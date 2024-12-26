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
	"os"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type StressLogger struct {
	logger  *zap.Logger
	logChan chan *LogEntry
	wg      sync.WaitGroup
	module  string
	file    *lumberjack.Logger
}

type LogEntry struct {
	level   string
	message string
}

func NewStressLogger(logDir, logFile, moduleName string) (*StressLogger, error) {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	fileWriter := &lumberjack.Logger{
		Filename:   logDir + logFile,
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	// 配置日志的编码器，增加对时间、模块、行号的支持
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 设置日志输出，控制台和文件同时输出
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)),
		zap.InfoLevel,
	)

	logger := zap.New(core)

	stressLogger := &StressLogger{
		logger:  logger,
		logChan: make(chan *LogEntry, 1000),
		module:  moduleName,
		file:    fileWriter,
	}
	stressLogger.start()
	return stressLogger, nil
}

func (l *StressLogger) Log(level string, message string) {
	// 创建日志条目
	logMessage := &LogEntry{
		level:   level,
		message: message,
	}

	// 将日志消息推送到通道，支持异步批量记录
	l.logChan <- logMessage
}

func (l *StressLogger) start() {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()

		// 批量异步记录日志
		var logs []LogEntry

		for logMsg := range l.logChan {
			// 将日志信息从通道读取并附加到日志数组中
			logs = append(logs, *logMsg)

			// 当数组中有超过10条日志时，批量处理
			if len(logs) >= 10 {
				l.flushLogs(logs)
				logs = nil
			}
		}

		// 处理剩余的日志
		if len(logs) > 0 {
			l.flushLogs(logs)
		}
	}()
}

func (l *StressLogger) flushLogs(logs []LogEntry) {
	for _, logMsg := range logs {
		// 获取调用栈信息
		_, file, line, ok := runtime.Caller(2) // 获取日志函数调用的堆栈信息
		if !ok {
			file = "unknown"
			line = 0
		}

		// 获取当前时间
		currentTime := time.Now().Format("2006-01-02 15:04:05.000")
		logEntry := map[string]interface{}{
			"timestamp": currentTime,
			"level":     logMsg.level,
			"module":    l.module,
			"message":   logMsg.message,
			"file":      file,
			"line":      line,
		}

		// 根据日志级别记录不同的日志
		switch logMsg.level {
		case "INFO":
			l.logger.Info(logMsg.message, zap.Any("details", logEntry))
		case "WARN":
			l.logger.Warn(logMsg.message, zap.Any("details", logEntry))
		case "ERROR":
			l.logger.Error(logMsg.message, zap.Any("details", logEntry))
		default:
			l.logger.Debug(logMsg.message, zap.Any("details", logEntry))
		}
	}
}

func (l *StressLogger) Close() {
	close(l.logChan) // 关闭日志通道
	l.wg.Wait()      // 等待所有日志写入完成
	if l.file != nil {
		l.file.Close()
	}
}
