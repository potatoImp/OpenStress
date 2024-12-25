// log.go
// 日志模块
// 本文件负责管理系统日志的记录、格式化和输出。
// 主要功能包括：
// - 创建日志目录
// - 初始化日志记录器
// - 提供日志记录接口
// - 支持异步日志记录
// - 日志切割与压缩
// - 日志过期清理
//
// 技术实现细节：
// 1. 使用 lumberjack 包实现日志文件的切割与压缩。
// 2. 提供多级别的日志记录功能，包括 DEBUG、INFO、WARNING 和 ERROR。
// 3. 支持异步日志记录，提高日志写入性能。
// 4. 提供日志文件的自动管理，包括文件大小和保留天数的配置。

package pool

import (
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// StressLogger 结构体
// 记录器结构体
// 包含异步日志通道和日志文件
// 以及日志切割配置

type StressLogger struct {
	file    *lumberjack.Logger
	logger  *log.Logger
	logChan chan string
	wg      sync.WaitGroup
	module  string // 当前模块名
}

// NewStressLogger 创建新的 StressLogger 实例，支持输出到文件和控制台
func NewStressLogger(logDir, logFile, moduleName string) (*StressLogger, error) {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	stressLogger := &StressLogger{
		file: &lumberjack.Logger{
			Filename:   logDir + logFile,
			MaxSize:    10, // MB
			MaxBackups: 3,
			MaxAge:     28, // days
			Compress:   true,
		},
		logger:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile),
		logChan: make(chan string, 1000),
		module:  moduleName,
	}
	stressLogger.start()
	return stressLogger, nil
}

// Log 记录日志
func (l *StressLogger) Log(level string, message string) {
	currentTime := time.Now().Format("2006-01-02 15:04:05.000") // 毫秒级时间
	logMessage := currentTime + " [" + level + "] [" + l.module + "] " + message
	l.logChan <- logMessage
}

// start 启动异步日志记录
func (l *StressLogger) start() {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		for msg := range l.logChan {
			l.logger.Println(msg)
			l.file.Write([]byte(msg + "\n"))
		}
	}()
}

// Close 关闭日志文件和等待异步日志写入完成
func (l *StressLogger) Close() {
	close(l.logChan) // 关闭日志通道
	l.wg.Wait()      // 等待所有日志写入完成
	if l.file != nil {
		l.file.Close()
	}
}
