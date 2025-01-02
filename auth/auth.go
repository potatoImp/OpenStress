// auth.go
package auth

import (
	"OpenStress/pool"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"gopkg.in/yaml.v2"
	// "github.com/your-repo/OpenStress/pool"  // 导入 pool 包
)

// 认证模块

// AuthMode 认证模式
type AuthMode int

const (
	// ModeLocal 本地模式（仅使用配置文件）
	ModeLocal AuthMode = iota
	// ModeRedis Redis模式（配置文件+Redis）
	ModeRedis
)

// Permission 权限类型
type Permission string

const (
	PermissionSubmit  Permission = "submit"  // 提交任务权限
	PermissionManage  Permission = "manage"  // 管理任务权限
	PermissionMonitor Permission = "monitor" // 监控权限
)

// UserAuth 用户认证信息
type UserAuth struct {
	Username    string       `yaml:"username" json:"username"`
	Password    string       `yaml:"password" json:"-"`              // 配置文件中的密码
	APIKey      string       `yaml:"api_key" json:"api_key"`         // API密钥
	Permissions []Permission `yaml:"permissions" json:"permissions"` // 权限列表
}

// AuthConfig 认证配置
type AuthConfig struct {
	Users []UserAuth `yaml:"users"`
}

// RedisState Redis连接状态
type RedisState int

const (
	// StateDisconnected Redis断开连接
	StateDisconnected RedisState = iota
	// StateConnecting Redis正在连接
	StateConnecting
	// StateConnected Redis已连接
	StateConnected
)

// AuthCache 认证缓存
type AuthCache struct {
	cache sync.Map // 本地缓存
	ttl   time.Duration
}

// cacheItem 缓存项
type cacheItem struct {
	auth      *UserAuth
	timestamp time.Time
}

// AuthManager 认证管理器
type AuthManager struct {
	mu              sync.RWMutex
	config          *AuthConfig
	mode            AuthMode
	redisClient     *redis.Client
	redisOpts       *redis.Options
	logger          *pool.StressLogger // 使用 StressLogger
	ctx             context.Context
	cancel          context.CancelFunc
	redisState      RedisState
	reconnectChan   chan struct{}
	stateChangeChan chan RedisState
	localCache      *AuthCache
	validateChan    chan validateReq
}

// Logger 日志接口
type Logger interface {
	Log(level string, message string)
}

// validateReq 验证请求
type validateReq struct {
	apiKey   string
	respChan chan validateResp
}

// validateResp 验证响应
type validateResp struct {
	auth *UserAuth
	err  error
}

// NewAuthManager 创建认证管理器
func NewAuthManager(configPath string, redisOpts *redis.Options) (*AuthManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建日志记录器
	logger, err := pool.InitializeLogger("logs/", "auth.log", "auth")
	if err != nil {
		cancel() // 确保在错误返回时调用 cancel
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	am := &AuthManager{
		redisOpts:       redisOpts,
		ctx:             ctx,
		cancel:          cancel,
		reconnectChan:   make(chan struct{}, 1),
		stateChangeChan: make(chan RedisState, 10),
		redisState:      StateDisconnected,
		localCache:      &AuthCache{ttl: 5 * time.Minute},
		validateChan:    make(chan validateReq, 100),
		logger:          logger,
	}

	// 加载配置文件
	if err := am.loadConfig(configPath); err != nil {
		cancel() // 确保在错误返回时调用 cancel
		logger.Log("ERROR", fmt.Sprintf("Failed to load config: %v", err))
		am.Close() // 清理资源
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	logger.Log("INFO", "Configuration loaded successfully")

	// 初始化认证模式
	if err := am.initMode(); err != nil {
		cancel() // 确保在错误返回时调用 cancel
		logger.Log("ERROR", fmt.Sprintf("Failed to init auth mode: %v", err))
		am.Close() // 清理资源
		return nil, fmt.Errorf("failed to init auth mode: %v", err)
	}
	logger.Log("INFO", fmt.Sprintf("Auth manager initialized in %v mode", am.mode))

	// 启动Redis状态管理器
	go am.redisStateManager()

	// 启动验证处理器
	go am.validateWorker()

	// 启动缓存清理
	go am.cacheCleaner()

	return am, nil
}

// loadConfig 加载配置文件
func (am *AuthManager) loadConfig(configPath string) error {
	// 实现配置文件加载
	if configPath == "" {
		return fmt.Errorf("config path cannot be empty")
	}

	// 读取并解析配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	config := &AuthConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	am.config = config
	return nil
}

// initMode 初始化认证模式
func (am *AuthManager) initMode() error {
	// 尝试连接Redis
	if am.redisOpts != nil {
		if err := am.asyncConnect(); err != nil {
			am.logger.Log("WARNING", fmt.Sprintf("Failed to connect to Redis: %v, falling back to local mode", err))
			am.mode = ModeLocal
		} else {
			am.mode = ModeRedis
		}
	} else {
		am.mode = ModeLocal
	}

	am.logger.Log("INFO", fmt.Sprintf("Auth manager initialized in %v mode", am.mode))
	return nil
}

// redisStateManager Redis状态管理器
func (am *AuthManager) redisStateManager() {
	reconnectTimer := time.NewTimer(0)
	defer reconnectTimer.Stop()

	healthCheckTicker := time.NewTicker(5 * time.Second)
	defer healthCheckTicker.Stop()

	var consecutiveFailures int

	for {
		select {
		case <-am.ctx.Done():
			am.logger.Log("INFO", "Redis state manager stopping")
			return

		case newState := <-am.stateChangeChan:
			am.mu.Lock()
			oldState := am.redisState
			am.redisState = newState
			am.mu.Unlock()

			switch newState {
			case StateConnected:
				consecutiveFailures = 0
				am.logger.Log("INFO", "Redis connection established")
			case StateDisconnected:
				backoff := time.Duration(0)
				if consecutiveFailures == 0 {
					reconnectTimer.Reset(0)
				} else {
					backoff = time.Duration(math.Min(float64(consecutiveFailures)*2, 30)) * time.Second
					reconnectTimer.Reset(backoff)
				}
				consecutiveFailures++
				am.logger.Log("WARNING", fmt.Sprintf("Redis disconnected (state change from %v), attempt %d, next retry in %v", oldState, consecutiveFailures, backoff))
			}

		case <-reconnectTimer.C:
			if am.redisState == StateDisconnected {
				am.logger.Log("INFO", "Attempting to reconnect to Redis")
				go am.asyncConnect()
			}

		case <-healthCheckTicker.C:
			if am.redisState == StateConnected {
				go am.asyncHealthCheck()
			}
		}
	}
}

// asyncConnect 异步连接Redis
func (am *AuthManager) asyncConnect() error {
	if am.redisOpts == nil {
		return fmt.Errorf("redis options not configured")
	}

	errChan := make(chan error, 1)
	go func() {
		am.stateChangeChan <- StateConnecting
		am.logger.Log("INFO", "Starting Redis connection attempt")

		client := redis.NewClient(am.redisOpts)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			am.logger.Log("ERROR", fmt.Sprintf("Redis connection failed: %v", err))
			client.Close()
			am.stateChangeChan <- StateDisconnected
			errChan <- err
			return
		}

		am.mu.Lock()
		if am.redisClient != nil {
			am.logger.Log("INFO", "Closing existing Redis connection")
			am.redisClient.Close()
		}
		am.redisClient = client
		am.mode = ModeRedis
		am.mu.Unlock()

		am.logger.Log("INFO", "Redis connection established successfully")
		am.stateChangeChan <- StateConnected
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(6 * time.Second):
		return fmt.Errorf("connection timeout")
	}
}

// asyncHealthCheck 异步健康检查
func (am *AuthManager) asyncHealthCheck() {
	am.mu.RLock()
	client := am.redisClient
	am.mu.RUnlock()

	if client == nil {
		am.stateChangeChan <- StateDisconnected
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		am.logger.Log("WARNING", fmt.Sprintf("Redis health check failed: %v", err))
		am.stateChangeChan <- StateDisconnected
	}
}

// GetRedisState 获取Redis连接状态
func (am *AuthManager) GetRedisState() RedisState {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.redisState
}

// validateWorker 验证处理工作器
func (am *AuthManager) validateWorker() {
	am.logger.Log("INFO", "Starting validation worker")
	for {
		select {
		case <-am.ctx.Done():
			am.logger.Log("INFO", "Validation worker stopping")
			return
		case req := <-am.validateChan:
			auth, err := am.validateAPIKeyInternal(req.apiKey)
			select {
			case <-am.ctx.Done():
				return
			case req.respChan <- validateResp{auth: auth, err: err}:
			}
		}
	}
}

// cacheCleaner 缓存清理工作器
func (am *AuthManager) cacheCleaner() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			am.localCache.cache.Range(func(key, value interface{}) bool {
				if item, ok := value.(cacheItem); ok {
					if now.Sub(item.timestamp) > am.localCache.ttl {
						am.localCache.cache.Delete(key)
					}
				}
				return true
			})
		}
	}
}

// ValidateAPIKey 异步验证API密钥
func (am *AuthManager) ValidateAPIKey(apiKey string) (*UserAuth, error) {
	// 先检查本地缓存
	if auth := am.checkLocalCache(apiKey); auth != nil {
		am.logger.Log("DEBUG", fmt.Sprintf("API key validation successful (cache hit): %s", apiKey))
		return auth, nil
	}

	// 创建响应通道
	respChan := make(chan validateResp, 1)

	// 发送验证请求
	select {
	case am.validateChan <- validateReq{apiKey: apiKey, respChan: respChan}:
		am.logger.Log("DEBUG", fmt.Sprintf("Validation request queued for API key: %s", apiKey))
	case <-am.ctx.Done():
		am.logger.Log("WARNING", "Auth manager is shutting down during API key validation")
		return nil, fmt.Errorf("auth manager is shutting down")
	default:
		am.logger.Log("WARNING", "Validation channel full, falling back to synchronous validation")
		return am.validateAPIKeyInternal(apiKey)
	}

	// 等待响应
	select {
	case resp := <-respChan:
		if resp.err != nil {
			am.logger.Log("ERROR", fmt.Sprintf("API key validation failed: %v", resp.err))
		} else {
			am.logger.Log("INFO", fmt.Sprintf("API key validation successful: %s", apiKey))
		}
		return resp.auth, resp.err
	case <-am.ctx.Done():
		am.logger.Log("WARNING", "Auth manager shutdown during validation wait")
		return nil, fmt.Errorf("auth manager is shutting down")
	case <-time.After(3 * time.Second):
		am.logger.Log("ERROR", fmt.Sprintf("API key validation timeout: %s", apiKey))
		return nil, fmt.Errorf("validation timeout")
	}
}

// checkLocalCache 检查本地缓存
func (am *AuthManager) checkLocalCache(apiKey string) *UserAuth {
	if value, ok := am.localCache.cache.Load(apiKey); ok {
		if item, ok := value.(cacheItem); ok {
			if time.Since(item.timestamp) <= am.localCache.ttl {
				return item.auth
			}
			am.localCache.cache.Delete(apiKey)
		}
	}
	return nil
}

// validateAPIKeyInternal 内部验证API密钥
func (am *AuthManager) validateAPIKeyInternal(apiKey string) (*UserAuth, error) {
	am.mu.RLock()
	mode := am.mode
	am.mu.RUnlock()

	var auth *UserAuth
	var err error

	switch mode {
	case ModeRedis:
		auth, err = am.validateAPIKeyRedis(apiKey)
	default:
		auth, err = am.validateAPIKeyLocal(apiKey)
	}

	if err == nil && auth != nil {
		// 更新本地缓存
		am.localCache.cache.Store(apiKey, cacheItem{
			auth:      auth,
			timestamp: time.Now(),
		})
	}

	return auth, err
}

// validateAPIKeyRedis 使用Redis验证API密钥
func (am *AuthManager) validateAPIKeyRedis(apiKey string) (*UserAuth, error) {
	am.mu.RLock()
	client := am.redisClient
	am.mu.RUnlock()

	if client == nil {
		return am.validateAPIKeyLocal(apiKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 尝试从Redis获取
	data, err := client.Get(ctx, fmt.Sprintf("apikey:%s", apiKey)).Bytes()
	if err == nil {
		var auth UserAuth
		if err := json.Unmarshal(data, &auth); err == nil {
			return &auth, nil
		}
	}

	// Redis中未找到，从本地配置查找
	auth, err := am.validateAPIKeyLocal(apiKey)
	if err != nil {
		return nil, err
	}

	// 异步更新Redis缓存
	go am.updateRedisCache(apiKey, auth)

	return auth, nil
}

// updateRedisCache 异步更新Redis缓存
func (am *AuthManager) updateRedisCache(apiKey string, auth *UserAuth) {
	am.mu.RLock()
	client := am.redisClient
	am.mu.RUnlock()

	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if authData, err := json.Marshal(auth); err == nil {
		client.Set(ctx, fmt.Sprintf("apikey:%s", apiKey), authData, 24*time.Hour)
	}
}

// validateAPIKeyLocal 使用本地配置验证API密钥
func (am *AuthManager) validateAPIKeyLocal(apiKey string) (*UserAuth, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for _, user := range am.config.Users {
		if user.APIKey == apiKey {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("invalid API key")
}

// HasPermission 检查用户是否有指定权限
func (am *AuthManager) HasPermission(auth *UserAuth, perm Permission) bool {
	if auth == nil {
		return false
	}

	for _, p := range auth.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// Close 关闭认证管理器
func (am *AuthManager) Close() error {
	am.logger.Log("INFO", "Shutting down auth manager")

	// 先取消上下文，通知所有 goroutine 停止
	am.cancel()

	// 等待一小段时间，让 goroutine 有机会完成
	time.Sleep(100 * time.Millisecond)

	am.mu.Lock()
	defer am.mu.Unlock()

	if am.redisClient != nil {
		if err := am.redisClient.Close(); err != nil {
			am.logger.Log("ERROR", fmt.Sprintf("Error closing Redis connection: %v", err))
			return err
		}
	}

	// 关闭日志记录器
	am.logger.Close()

	return nil
}
