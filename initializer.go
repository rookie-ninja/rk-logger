// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rklogger

import (
	"encoding/json"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
)

type FileType int

var (
	StdoutEncoderConfig = NewZapStdoutEncoderConfig()
	StdoutLoggerConfig  = &zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       true,
		Encoding:          "console",
		DisableStacktrace: true,
		EncoderConfig:     *StdoutEncoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}
	StdoutLogger, _ = StdoutLoggerConfig.Build()
	NoopLogger      = zap.NewNop()

	EventLoggerConfigBytes = []byte(`{
     "level": "info",
     "encoding": "console",
     "outputPaths": ["stdout"],
     "errorOutputPaths": ["stderr"],
     "initialFields": {},
     "encoderConfig": {
       "messageKey": "msg",
       "levelKey": "",
       "nameKey": "",
       "timeKey": "",
       "callerKey": "",
       "stacktraceKey": "",
       "callstackKey": "",
       "errorKey": "",
       "timeEncoder": "iso8601",
       "fileKey": "",
       "levelEncoder": "capital",
       "durationEncoder": "second",
       "callerEncoder": "full",
       "nameEncoder": "full"
     },
    "maxsize": 1024,
    "maxage": 7,
    "maxbackups": 3,
    "localtime": true,
    "compress": true
   }`)
	EventLogger, EventLoggerConfig, _ = NewZapLoggerWithBytes(EventLoggerConfigBytes, JSON)

	LumberjackConfig = NewLumberjackConfigDefault()
)

// Config file type which support json, yaml, toml and hcl
// JSON: https://www.json.org/
// YAML: https://yaml.org/
const (
	JSON FileType = 0
	YAML FileType = 1
)

// Stringfy above config file types.
func (fileType FileType) String() string {
	names := [...]string{"JSON", "YAML"}

	// Please do not forget to change the boundary while adding a new config file types
	if fileType < JSON || fileType > YAML {
		return "UNKNOWN"
	}

	return names[fileType]
}

func NewZapEventConfig() *zap.Config {
	_, config, _ := NewZapLoggerWithBytes(EventLoggerConfigBytes, JSON)
	return config
}

// Create new default lumberjack config
func NewLumberjackConfigDefault() *lumberjack.Logger {
	return &lumberjack.Logger{
		MaxSize:    1,
		MaxAge:     7,
		MaxBackups: 3,
		LocalTime:  true,
		Compress:   true,
	}
}

// Create new stdout encoder config
func NewZapStdoutEncoderConfig() *zapcore.EncoderConfig {
	return &zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// Create new stdout config
func NewZapStdoutConfig() *zap.Config {
	return &zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       true,
		Encoding:          "console",
		DisableStacktrace: true,
		EncoderConfig:     *NewZapStdoutEncoderConfig(),
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}
}

// Init zap logger with byte array from content of config file
// lumberjack.Logger could be empty, if not provided,
// then, we will use default write sync
func NewZapLoggerWithBytes(raw []byte, fileType FileType, opts ...zap.Option) (*zap.Logger, *zap.Config, error) {
	if raw == nil {
		return nil, nil, errors.New("input byte array is nil")
	}

	if len(raw) == 0 {
		return nil, nil, errors.New("byte array is empty")
	}

	// Initialize zap logger from config file
	var logger *zap.Logger
	var err error
	zapConfig := &zap.Config{}
	lumberConfig := &lumberjack.Logger{}

	if fileType == JSON {
		// parse zap json file
		if err := json.Unmarshal(raw, zapConfig); err != nil {
			return nil, nil, err
		}

		// parse lumberjack json file
		if err := json.Unmarshal(raw, lumberConfig); err != nil {
			return nil, nil, err
		}

		logger, err = NewZapLoggerWithConf(zapConfig, lumberConfig, opts...)
	} else if fileType == YAML {
		// parse zap yaml file
		if err := yaml.Unmarshal(raw, zapConfig); err != nil {
			return nil, nil, err
		}

		// parse lumberjack yaml file
		if err := yaml.Unmarshal(raw, lumberConfig); err != nil {
			return nil, nil, err
		}

		logger, err = NewZapLoggerWithConf(zapConfig, lumberConfig, opts...)
	} else {
		logger, err = nil, errors.New("invalid config file")
	}

	// make sure we return nil for logger and logger config
	if err != nil {
		return nil, nil, err
	}

	return logger, zapConfig, err
}

// Init zap logger with config file path
// File path needs to be absolute path
// lumberjack.Logger could be empty, if not provided,
// then, we will use default write sync
func NewZapLoggerWithConfPath(filePath string, fileType FileType, opts ...zap.Option) (*zap.Logger, *zap.Config, error) {
	if len(filePath) == 0 {
		return nil, nil, errors.New("file path is empty")
	}

	// Initialize zap logger from config file
	var logger *zap.Logger
	var err error
	var config *zap.Config

	err = validateFilePath(filePath)

	if err == nil {
		bytes, readErr := ioutil.ReadFile(filePath)
		if readErr != nil {
			return logger, config, readErr
		}

		logger, config, err = NewZapLoggerWithBytes(bytes, fileType, opts...)
	}

	return logger, config, err
}

// Init zap logger with config
// File path needs to be absolute path
// lumberjack.Logger could be empty, if not provided,
// then, we will use default write sync
func NewZapLoggerWithConf(config *zap.Config, lumber *lumberjack.Logger, opts ...zap.Option) (*zap.Logger, error) {
	// Validate parameters
	if config == nil {
		return nil, errors.New("zap config is nil")
	}

	if lumber == nil {
		return config.Build(opts...)
	}

	sync := make([]zapcore.WriteSyncer, 0, 0)
	// Iterate output path and attach to lumberjack
	// Remember, each logger will use same lumberjack logger configuration
	for i := range config.OutputPaths {
		if config.OutputPaths[i] != "stdout" {
			lumberNew := &lumberjack.Logger{
				Filename:   config.OutputPaths[i],
				MaxAge:     lumber.MaxAge,
				MaxBackups: lumber.MaxBackups,
				MaxSize:    lumber.MaxSize,
				Compress:   lumber.Compress,
				LocalTime:  lumber.LocalTime,
			}

			sync = append(sync, zapcore.AddSync(lumberNew))
		} else {
			stdout, close, err := zap.Open("stdout")
			// just close the syncer if err occurs
			if err != nil {
				close()
			} else {
				sync = append(sync, stdout)
			}
		}
	}

	core := zapcore.NewCore(
		generateEncoder(config),
		zap.CombineWriteSyncers(sync...),
		config.Level)

	// add initial fields
	initialFields := make([]zap.Field, 0, 0)
	for k, v := range config.InitialFields {
		initialFields = append(initialFields, zap.Any(k, v))
	}

	// add error output sync
	if len(config.ErrorOutputPaths) > 0 {
		errSink, _, err := zap.Open(config.ErrorOutputPaths...)
		if err != nil {
			return nil, err
		}
		opts = append(opts, zap.ErrorOutput(errSink))
	}

	return zap.New(core, opts...).With(initialFields...), nil
}

// Init lumberjack logger as write sync with raw byte array of config file
func NewLumberjackLoggerWithBytes(raw []byte, fileType FileType) (*lumberjack.Logger, error) {
	if raw == nil {
		return nil, errors.New("input byte array is nil")
	}

	if len(raw) == 0 {
		return nil, errors.New("byte array is empty")
	}

	logger := &lumberjack.Logger{}
	// unmarshal as yaml
	if fileType == YAML {
		if err := yaml.Unmarshal(raw, logger); err != nil {
			return nil, err
		}
	} else if fileType == JSON {
		if err := json.Unmarshal(raw, logger); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unknown type")
	}

	return logger, nil
}

// Init lumberjack logger as write sync with lumberjack config file path
// File path needs to be absolute path
func NewLumberjackLoggerWithConfPath(filePath string, fileType FileType) (*lumberjack.Logger, error) {
	if len(filePath) == 0 {
		return nil, errors.New("file path is empty")
	}

	var logger *lumberjack.Logger
	var err error

	err = validateFilePath(filePath)

	if err == nil {
		bytes, readErr := ioutil.ReadFile(filePath)

		if readErr == nil {
			logger, err = NewLumberjackLoggerWithBytes(bytes, fileType)
		} else {
			err = readErr
		}
	}

	return logger, err
}

func validateFilePath(filePath string) error {
	_, err := os.Stat(filePath)

	if err != nil {
		if os.IsNotExist(err) {
			err = errors.Errorf("file does not exists, filePath:%s", filePath)
		} else {
			err = errors.Errorf("error thrown while reading file, filePath:%s", filePath)
		}
	}

	return err
}

// Generate zap encoder from zap config
func generateEncoder(config *zap.Config) zapcore.Encoder {
	if config.Encoding == "json" {
		return zapcore.NewJSONEncoder(config.EncoderConfig)
	} else {
		// default is console encoding
		return zapcore.NewConsoleEncoder(config.EncoderConfig)
	}
}

// Parse relative path, convert it to current working directory
func toAbsoluteWorkingDir(filePath string) (string, error) {
	if path.IsAbs(filePath) {
		return filePath, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// relative path, add current working directory
	return path.Clean(path.Join(dir, filePath)), nil
}

// Transform wrapped zap config into zap.Config
func TransformToZapConfig(wrap *ZapConfigWrap) *zap.Config {
	level := zap.NewAtomicLevel()

	if err := level.UnmarshalText([]byte(wrap.Level)); err != nil {
		level.SetLevel(zapcore.InfoLevel)
	}

	config := &zap.Config{
		Level:             level,
		Development:       wrap.Development,
		DisableCaller:     wrap.DisableCaller,
		DisableStacktrace: wrap.DisableStacktrace,
		Sampling:          wrap.Sampling,
		Encoding:          wrap.Encoding,
		EncoderConfig:     wrap.EncoderConfig,
		OutputPaths:       wrap.OutputPaths,
		ErrorOutputPaths:  wrap.ErrorOutputPaths,
		InitialFields:     wrap.InitialFields,
	}

	return config
}

// A wrapper zap config which copied from zap.Config
// This is used while parsing zap yaml config to zap.Config with viper
// because Level would throw an error since it is not a type of string
type ZapConfigWrap struct {
	// Level is the minimum enabled logging level. Note that this is a dynamic
	// level, so calling Config.Level.SetLevel will atomically change the log
	// level of all loggers descended from this config.
	Level string `json:"level" yaml:"level"`
	// Development puts the logger in development mode, which changes the
	// behavior of DPanicLevel and takes stacktraces more liberally.
	Development bool `json:"development" yaml:"development"`
	// DisableCaller stops annotating logs with the calling function's file
	// name and line number. By default, all logs are annotated.
	DisableCaller bool `json:"disableCaller" yaml:"disableCaller"`
	// DisableStacktrace completely disables automatic stacktrace capturing. By
	// default, stacktraces are captured for WarnLevel and above logs in
	// development and ErrorLevel and above in production.
	DisableStacktrace bool `json:"disableStacktrace" yaml:"disableStacktrace"`
	// Sampling sets a sampling policy. A nil SamplingConfig disables sampling.
	Sampling *zap.SamplingConfig `json:"sampling" yaml:"sampling"`
	// Encoding sets the logger's encoding. Valid values are "json" and
	// "console", as well as any third-party encodings registered via
	// RegisterEncoder.
	Encoding string `json:"encoding" yaml:"encoding"`
	// EncoderConfig sets options for the chosen encoder. See
	// zapcore.EncoderConfig for details.
	EncoderConfig zapcore.EncoderConfig `json:"encoderConfig" yaml:"encoderConfig"`
	// OutputPaths is a list of URLs or file paths to write logging output to.
	// See Open for details.
	OutputPaths []string `json:"outputPaths" yaml:"outputPaths"`
	// ErrorOutputPaths is a list of URLs to write internal logger errors to.
	// The default is standard error.
	//
	// Note that this setting only affects internal errors; for sample code that
	// sends error-level logs to a different location from info- and debug-level
	// logs, see the package-level AdvancedConfiguration example.
	ErrorOutputPaths []string `json:"errorOutputPaths" yaml:"errorOutputPaths"`
	// InitialFields is a collection of fields to add to the root logger.
	InitialFields map[string]interface{} `json:"initialFields" yaml:"initialFields"`
}
