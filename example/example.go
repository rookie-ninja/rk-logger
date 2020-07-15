// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path"
	"rookie-ninja/rk-logger"
)

func main() {
	NewZapLoggerWithBytesExample()
	NewZapLoggerWithConfPathExample()
	NewZapLoggerWithConfExample()
	NewLumberjackLoggerWithBytesExample()
	NewLumberjackLoggerWithConfPathExample()
}

func NewLumberjackLoggerWithBytesExample() {
	bytes := []byte(`{
     "maxsize": 1,
     "maxage": 7,
     "maxbackups": 3,
     "localtime": true,
     "compress": true
    }`)
	_, err := rk_logger.NewLumberjackLoggerWithBytes(bytes, rk_logger.JSON)
	if err != nil {
		panic(err)
	}
}

func NewLumberjackLoggerWithConfPathExample() {
	// get current working directory
	dir, _ := os.Getwd()
	// init lumberjack logger
	_, err := rk_logger.NewLumberjackLoggerWithConfPath(path.Clean(path.Join(dir, "/assets/lumberjack.yaml")), rk_logger.YAML)
	if err != nil {
		panic(err)
	}
}

func NewZapLoggerWithConfExample() {
	encodingConfig := zapcore.EncoderConfig{
		TimeKey:        "zap_timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := &zap.Config{
		Level:         zap.NewAtomicLevelAt(zap.InfoLevel),
		EncoderConfig: encodingConfig,
		OutputPaths:   []string{"stdout", "logs/rk-logger.log"},
	}
	logger, _ := rk_logger.NewZapLoggerWithConf(config, &lumberjack.Logger{})
	logger.Info("NewZapLoggerWithConfExample")
}

func NewZapLoggerWithConfPathExample() {
	// get current working directory
	dir, _ := os.Getwd()
	// init lumberjack logger
	lumberjack, _ := rk_logger.NewLumberjackLoggerWithConfPath(path.Clean(path.Join(dir, "/assets/lumberjack.yaml")), rk_logger.YAML)
	// init zap logger
	logger, _, _ := rk_logger.NewZapLoggerWithConfPath(path.Clean(path.Join(dir, "/assets/zap.yaml")), rk_logger.YAML, lumberjack)
	// use it
	logger.Info("NewZapLoggerWithConfPathExample")
}

func NewZapLoggerWithBytesExample() {
	zapBytes := []byte(`{
      "level": "debug",
      "encoding": "console",
      "outputPaths": ["stdout", "logs/rk-logger.log"],
      "errorOutputPaths": ["stderr"],
      "initialFields": {"initFieldKey": "fieldValue"},
      "encoderConfig": {
        "messageKey": "message",
        "levelKey": "level",
        "nameKey": "logger",
        "timeKey": "time",
        "callerKey": "caller",
        "stacktraceKey": "stacktrace",
        "callstackKey": "callstack",
        "errorKey": "error",
        "timeEncoder": "iso8601",
        "fileKey": "file",
        "levelEncoder": "capital",
        "durationEncoder": "second",
        "callerEncoder": "full",
        "nameEncoder": "full",
        "sampling": {
            "initial": "3",
            "thereafter": "10"
        }
      }
    }`)

	lumberBytes := []byte(`{
     "maxsize": 1,
     "maxage": 7,
     "maxbackups": 3,
     "localtime": true,
     "compress": true
    }`)

	lumber, err := rk_logger.NewLumberjackLoggerWithBytes(lumberBytes, rk_logger.JSON)
	if err != nil {
		panic(err)
	}

	logger, _, err := rk_logger.NewZapLoggerWithBytes(zapBytes, rk_logger.JSON, lumber)

	if err != nil {
		panic(err)
	}

	logger.Info("NewZapLoggerWithBytesExample")
}
