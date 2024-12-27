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

package pool

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// StressLogger 表示一个日志记录器
type StressLogger struct {
	logger  *zap.Logger
	logChan chan *LogEntry
	wg      sync.WaitGroup
	module  string
	file    *lumberjack.Logger
}

// LogEntry 表示一条日志记录
type LogEntry struct {
	level   string
	message string
}

// Declare a global variable to hold the logger instance
var globalLogger *StressLogger

// This function is now only responsible for starting the logger if not already started
func GetLogger() (*StressLogger, error) {
	if globalLogger == nil {
		return nil, fmt.Errorf("logger not initialized")
	}
	return globalLogger, nil
}

var once sync.Once

func InitializeLogger(logDir, logFile, moduleName string) (*StressLogger, error) {
	var err error
	once.Do(func() {
		if globalLogger != nil {
			return
		}

		// Ensure the log directory exists
		if err = os.MkdirAll(logDir, os.ModePerm); err != nil {
			return
		}

		fileWriter := &lumberjack.Logger{
			Filename:   logDir + logFile,
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)),
			zap.InfoLevel,
		)

		logger := zap.New(core)

		stressLogger = &StressLogger{
			logger:  logger,
			logChan: make(chan *LogEntry, 1000),
			module:  moduleName,
			file:    fileWriter,
		}

		// Start the logger's asynchronous processing
		stressLogger.start()

		globalLogger = stressLogger
	})
	return globalLogger, err
}

// Log records a log entry
func (l *StressLogger) Log(level string, message string) {
	// Create a log entry
	logMessage := &LogEntry{
		level:   level,
		message: message,
	}

	// Push the log message into the channel for asynchronous batch processing
	l.logChan <- logMessage
}

// start begins the process of handling log messages asynchronously
func (l *StressLogger) start() {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()

		// Batch logs asynchronously
		var logs []LogEntry

		for logMsg := range l.logChan {
			logs = append(logs, *logMsg)

			// Process logs when there are 10 or more
			if len(logs) >= 10 {
				l.flushLogs(logs)
				logs = nil
			}
		}

		// Process any remaining logs
		if len(logs) > 0 {
			l.flushLogs(logs)
		}
	}()
}

// flushLogs writes a batch of logs to the storage
func (l *StressLogger) flushLogs(logs []LogEntry) {
	for _, logMsg := range logs {
		// Get stack trace information
		_, file, line, ok := runtime.Caller(2) // Get the stack trace of the log function call
		if !ok {
			file = "unknown"
			line = 0
		}

		// Get the current timestamp
		currentTime := time.Now().Format("2006-01-02 15:04:05.000")
		logEntry := map[string]interface{}{
			"timestamp": currentTime,
			"level":     logMsg.level,
			"module":    l.module,
			"message":   logMsg.message,
		}

		// Record logs based on their level
		switch logMsg.level {
		case "INFO":
			// Only record essential info
			l.logger.Info(logMsg.message, zap.Any("details", logEntry))
		case "ERROR", "DEBUG":
			// ERROR and DEBUG levels include file and line information
			logEntry["file"] = file
			logEntry["line"] = line
			l.logger.Error(logMsg.message, zap.Any("details", logEntry))
		default:
			l.logger.Debug(logMsg.message, zap.Any("details", logEntry))
		}
	}
}

// Close stops the logger and ensures all logs are written
func (l *StressLogger) Close() {
	close(l.logChan) // Close the log channel
	l.wg.Wait()      // Wait for all logs to be processed
	if l.file != nil {
		l.file.Close()
	}
}
