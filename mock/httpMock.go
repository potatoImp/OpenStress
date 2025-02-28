package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	port         = flag.Int("port", 8089, "服务监听端口")
	successRate  = flag.Float64("success-rate", 100.0, "成功响应比例 (0-100)")
	requestCount uint64
)

// 响应结构体
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Count   uint64 `json:"count"`
	Status  int    `json:"status"`
}

// 处理HTTP请求
func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 增加请求计数
	count := atomic.AddUint64(&requestCount, 1)

	// 获取客户端信息
	clientIP := r.RemoteAddr

	// 根据成功率决定响应
	success := rand.Float64()*100 <= *successRate

	// 准备响应数据
	resp := Response{
		Success: success,
		Message: "请求处理完成",
		Count:   count,
		Status:  2000,
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 如果不成功，设置500状态码
	if !success {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Message = "服务器内部错误"
	}

	// 编码并发送响应
	json.NewEncoder(w).Encode(resp)
	// 打印请求信息
	log.Printf("[请求 #%d] 来自 %s - 路径: %s - 响应状态: %v\n", count, clientIP, r.URL.Path, success)
}

// 处理长连接请求
func handleLongConnection(w http.ResponseWriter, r *http.Request) {
	// 设置响应头，启用长连接
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 获取客户端信息
	clientIP := r.RemoteAddr
	// 增加请求计数
	count := atomic.AddUint64(&requestCount, 1)

	// 创建通知通道
	done := make(chan bool)

	// 监听客户端断开连接
	go func() {
		<-r.Context().Done()
		close(done)
	}()

	// 定时发送数据
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Printf("[长连接 #%d] 建立连接 - 来自 %s - 路径: %s\n", count, clientIP, r.URL.Path)

	for {
		select {
		case <-done:
			log.Printf("[长连接 #%d] 连接断开 - 来自 %s\n", count, clientIP)
			return
		case <-ticker.C:
			// 根据成功率决定响应
			success := rand.Float64()*100 <= *successRate

			resp := Response{
				Success: success,
				Message: "长连接心跳",
				Count:   count,
				Status:  2000,
			}

			// 编码并发送响应
			data, _ := json.Marshal(resp)
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()

			log.Printf("[长连接 #%d] 发送心跳 - 来自 %s - 路径: %s - 响应状态: %v\n", count, clientIP, r.URL.Path, success)
		}
	}
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 配置日志输出
	log.SetFlags(log.Ltime | log.Lmsgprefix)

	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 设置路由
	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/stream", handleLongConnection)

	// 创建服务器
	server := &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%d", *port),
	}

	// 优雅关闭服务器
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动服务器
	go func() {
		log.Printf("服务器启动在 http://0.0.0.0:%d\n", *port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("服务器错误: %v\n", err)
		}
	}()

	// 等待中断信号
	<-sigChan
	log.Println("正在关闭服务器...")

	// 保持窗口不关闭
	log.Println("按回车键退出...")
	fmt.Scanln()
}
