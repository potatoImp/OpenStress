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
	logger       *zap.Logger
	logChan      chan *LogEntry
	wg           sync.WaitGroup
	module       string
	file         *lumberjack.Logger
	closed       bool
	mu           sync.Mutex // Protects the closed flag and channels
	currentLevel zapcore.Level
}

// LogEntry 表示一条日志记录
type LogEntry struct {
	level   string
	message string
}

// Declare a global variable to hold the logger instance
var globalLogger *StressLogger

// DefaultLogLevel 默认日志级别，初始化为 INFO
var DefaultLogLevel zapcore.Level = zap.InfoLevel

// This function is now only responsible for starting the logger if not already started
func GetLogger() (*StressLogger, error) {
	if globalLogger == nil {
		return nil, fmt.Errorf("logger not initialized")
	}
	return globalLogger, nil
}

var once sync.Once

// InitializeLogger 创建并初始化日志记录器
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
			// Write only to file
			zapcore.AddSync(fileWriter),
			DefaultLogLevel, // Use the global default level
		)

		logger := zap.New(core)

		stressLogger = &StressLogger{
			logger:       logger,
			logChan:      make(chan *LogEntry, 1000),
			module:       moduleName,
			file:         fileWriter,
			closed:       false,
			currentLevel: DefaultLogLevel,
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

	// Locking here to make sure the channel is not closed while logging
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return // If the logger is closed, do not log
	}

	// Only log the message if its level is >= current log level
	if levelPriority(level) >= levelPriority(l.currentLevel.String()) {
		// Push the log message into the channel for asynchronous processing
		l.logChan <- logMessage
	}
}

// levelPriority returns the integer priority for a log level.
func levelPriority(level string) int {
	switch level {
	case "DEBUG":
		return 1
	case "INFO":
		return 2
	case "WARN":
		return 3
	case "ERROR":
		return 4
	default:
		return 0
	}
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
			l.logger.Info(logMsg.message, zap.Any("details", logEntry))
		case "ERROR", "DEBUG":
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
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return // If the logger is already closed, return
	}

	// Close the log channel
	l.closed = true
	close(l.logChan) // Close the log channel
	l.wg.Wait()      // Wait for all logs to be processed
	if l.file != nil {
		l.file.Close()
	}
}

// SetLogLevel 动态设置日志级别
func SetLogLevel(level string) error {
	var zapLevel zapcore.Level
	switch level {
	case "DEBUG":
		zapLevel = zap.DebugLevel
	case "INFO":
		zapLevel = zap.InfoLevel
	case "WARN":
		zapLevel = zap.WarnLevel
	case "ERROR":
		zapLevel = zap.ErrorLevel
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}

	// Update the global logger level
	DefaultLogLevel = zapLevel

	// Update the logger core
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	fileWriter := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(fileWriter),
		DefaultLogLevel, // Use the updated global level
	)

	// Recreate the logger with the new level
	globalLogger.logger = zap.New(core)

	return nil
}
