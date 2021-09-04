// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package main contains example of rklogger usages.
package main

import (
	"github.com/rookie-ninja/rk-logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path"
)

// Main entrance.
func main() {
	NewZapLoggerWithBytesExample()
	//NewZapLoggerWithConfPathExample()
	//NewZapLoggerWithConfExample()
	//NewLumberjackLoggerWithBytesExample()
	//NewLumberjackLoggerWithConfPathExample()
}

// Create a new lumberjack instance with raw bytes.
func NewLumberjackLoggerWithBytesExample() {
	bytes := []byte(`{
     "maxsize": 1,
     "maxage": 7,
     "maxbackups": 3,
     "localtime": true,
     "compress": true
    }`)
	_, err := rklogger.NewLumberjackLoggerWithBytes(bytes, rklogger.JSON)
	if err != nil {
		panic(err)
	}
}

// Create a new lumberjack instance with config file.
func NewLumberjackLoggerWithConfPathExample() {
	// get current working directory
	dir, _ := os.Getwd()
	// init lumberjack logger
	_, err := rklogger.NewLumberjackLoggerWithConfPath(path.Clean(path.Join(dir, "/assets/lumberjack.yaml")), rklogger.YAML)
	if err != nil {
		panic(err)
	}
}

// Create a new zap logger instance with zap config.
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
	logger, _ := rklogger.NewZapLoggerWithConf(config, &lumberjack.Logger{})
	logger.Info("NewZapLoggerWithConfExample")
}

// Create a new zap logger instance with config file.
func NewZapLoggerWithConfPathExample() {
	// get current working directory
	dir, _ := os.Getwd()
	// init zap logger
	logger, _, _ := rklogger.NewZapLoggerWithConfPath(path.Clean(path.Join(dir, "/assets/zap.yaml")), rklogger.YAML)
	// use it
	logger.Info("NewZapLoggerWithConfPathExample")
}

// Create a new zap logger instance with raw bytes.
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
      },
     "maxsize": 1,
     "maxage": 7,
     "maxbackups": 3,
     "localtime": true,
     "compress": true
    }`)

	logger, _, err := rklogger.NewZapLoggerWithBytes(zapBytes, rklogger.JSON)

	if err != nil {
		panic(err)
	}

	logger.Info("NewZapLoggerWithBytesExample")
}
