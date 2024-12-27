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

// Authentication module

// AuthMode Authentication mode
type AuthMode int

const (
	// ModeLocal Local mode (uses configuration file only)
	ModeLocal AuthMode = iota
	// ModeRedis Redis mode (configuration file + Redis)
	ModeRedis
)

// Permission Permission type
type Permission string

const (
	// PermissionSubmit Permission to submit tasks
	PermissionSubmit  Permission = "submit"
	// PermissionManage Permission to manage tasks
	PermissionManage  Permission = "manage"
	// PermissionMonitor Permission to monitor
	PermissionMonitor Permission = "monitor"
)

// UserAuth User authentication information
type UserAuth struct {
	Username    string       `yaml:"username" json:"username"`
	Password    string       `yaml:"password" json:"-"`              // Password in the configuration file
	APIKey      string       `yaml:"api_key" json:"api_key"`         // API key
	Permissions []Permission `yaml:"permissions" json:"permissions"` // List of permissions
}

// AuthConfig Authentication configuration
type AuthConfig struct {
	Users []UserAuth `yaml:"users"`
}

// RedisState Redis connection state
type RedisState int

const (
	// StateDisconnected Redis disconnected
	StateDisconnected RedisState = iota
	// StateConnecting Redis connecting
	StateConnecting
	// StateConnected Redis connected
	StateConnected
)

// AuthCache Authentication cache
type AuthCache struct {
	cache sync.Map // Local cache
	ttl   time.Duration
}

// cacheItem Cache item
type cacheItem struct {
	auth      *UserAuth
	timestamp time.Time
}

// AuthManager Authentication manager
type AuthManager struct {
	mu              sync.RWMutex
	config          *AuthConfig
	mode            AuthMode
	redisClient     *redis.Client
	redisOpts       *redis.Options
	logger          *pool.StressLogger // Using StressLogger
	ctx             context.Context
	cancel          context.CancelFunc
	redisState      RedisState
	reconnectChan   chan struct{}
	stateChangeChan chan RedisState
	localCache      *AuthCache
	validateChan    chan validateReq
}

// Logger Logging interface
type Logger interface {
	Log(level string, message string)
}

// validateReq Validation request
type validateReq struct {
	apiKey   string
	respChan chan validateResp
}

// validateResp Validation response
type validateResp struct {
	auth *UserAuth
	err  error
}

// NewAuthManager Create authentication manager
func NewAuthManager(configPath string, redisOpts *redis.Options) (*AuthManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create logger
	logger, err := pool.NewStressLogger("logs/", "auth.log", "auth")
	if err != nil {
		cancel() // Ensure cancel is called on error return
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

	// Load configuration file
	if err := am.loadConfig(configPath); err != nil {
		cancel() // Ensure cancel is called on error return
		logger.Log("ERROR", fmt.Sprintf("Failed to load config: %v", err))
		am.Close() // Clean up resources
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	logger.Log("INFO", "Configuration loaded successfully")

	// Initialize authentication mode
	if err := am.initMode(); err != nil {
		cancel() // Ensure cancel is called on error return
		logger.Log("ERROR", fmt.Sprintf("Failed to init auth mode: %v", err))
		am.Close() // Clean up resources
		return nil, fmt.Errorf("failed to init auth mode: %v", err)
	}
	logger.Log("INFO", fmt.Sprintf("Auth manager initialized in %v mode", am.mode))

	// Start Redis state manager
	go am.redisStateManager()

	// Start validation processing worker
	go am.validateWorker()

	// Start cache cleaning worker
	go am.cacheCleaner()

	return am, nil
}

// loadConfig Load configuration file
func (am *AuthManager) loadConfig(configPath string) error {
	// Implement configuration file loading
	if configPath == "" {
		return fmt.Errorf("config path cannot be empty")
	}

	// Read and parse configuration file
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

// initMode Initialize authentication mode
func (am *AuthManager) initMode() error {
	// Attempt to connect to Redis
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

// redisStateManager Redis state manager
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

// asyncConnect Asynchronously connect to Redis
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

// asyncHealthCheck Asynchronous health check
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

// GetRedisState Get Redis connection state
func (am *AuthManager) GetRedisState() RedisState {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.redisState
}

// validateWorker Validation processing worker
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

// cacheCleaner Cache cleaning worker
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

// ValidateAPIKey Asynchronously validate API key
func (am *AuthManager) ValidateAPIKey(apiKey string) (*UserAuth, error) {
	// Check local cache first
	if auth := am.checkLocalCache(apiKey); auth != nil {
		am.logger.Log("DEBUG", fmt.Sprintf("API key validation successful (cache hit): %s", apiKey))
		return auth, nil
	}

	// Create response channel
	respChan := make(chan validateResp, 1)

	// Send validation request
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

	// Wait for response
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

// checkLocalCache Check local cache
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

// validateAPIKeyInternal Internally validate API key
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
		// Update local cache
		am.localCache.cache.Store(apiKey, cacheItem{
			auth:      auth,
			timestamp: time.Now(),
		})
	}

	return auth, err
}

// validateAPIKeyRedis Validate API key using Redis
func (am *AuthManager) validateAPIKeyRedis(apiKey string) (*UserAuth, error) {
	am.mu.RLock()
	client := am.redisClient
	am.mu.RUnlock()

	if client == nil {
		return am.validateAPIKeyLocal(apiKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Attempt to get from Redis
	data, err := client.Get(ctx, fmt.Sprintf("apikey:%s", apiKey)).Bytes()
	if err == nil {
		var auth UserAuth
		if err := json.Unmarshal(data, &auth); err == nil {
			return &auth, nil
		}
	}

	// Not found in Redis, check local configuration
	auth, err := am.validateAPIKeyLocal(apiKey)
	if err != nil {
		return nil, err
	}

	// Asynchronously update Redis cache
	go am.updateRedisCache(apiKey, auth)

	return auth, nil
}

// updateRedisCache Asynchronously update Redis cache
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

// validateAPIKeyLocal Validate API key using local configuration
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

// HasPermission Check if the user has the specified permission
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

// Close Close authentication manager
func (am *AuthManager) Close() error {
	am.logger.Log("INFO", "Shutting down auth manager")

	// Cancel context first, notify all goroutines to stop
	am.cancel()

	// Wait for a short time to allow goroutines to complete
	time.Sleep(100 * time.Millisecond)

	am.mu.Lock()
	defer am.mu.Unlock()

	if am.redisClient != nil {
		if err := am.redisClient.Close(); err != nil {
			am.logger.Log("ERROR", fmt.Sprintf("Error closing Redis connection: %v", err))
			return err
		}
	}

	// Close logger
	am.logger.Close()

	return nil
}
