// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rklogger contains couple of utility functions for initializing zap logger.
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
	"reflect"
)

var (
	// Default zap logger encoder config whose output path is stdout.
	StdoutEncoderConfig = NewZapStdoutEncoderConfig()
	// Default zap logger config whose output path is stdout.
	StdoutLoggerConfig = &zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       true,
		Encoding:          "console",
		DisableStacktrace: true,
		EncoderConfig:     *StdoutEncoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}
	// Default zap logger whose output path is stdout.
	StdoutLogger, _ = StdoutLoggerConfig.Build()
	// Default zap noop logger.
	NoopLogger = zap.NewNop()

	// Default zap logger which is used by EventLogger.
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
	// Default EventLogger and EventLoggerConfig.
	EventLogger, EventLoggerConfig, _ = NewZapLoggerWithBytes(EventLoggerConfigBytes, JSON)

	// Default lumberjack config.
	LumberjackConfig = NewLumberjackConfigDefault()
)

// Config file type which support json and yaml currently.
// JSON: https://www.json.org/
// YAML: https://yaml.org/
type FileType int

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

// NewZapEventConfig creates new zap.Config for EventLogger
func NewZapEventConfig() *zap.Config {
	_, config, _ := NewZapLoggerWithBytes(EventLoggerConfigBytes, JSON)
	return config
}

// NewLumberjackConfigDefault creates new default lumberjack config
func NewLumberjackConfigDefault() *lumberjack.Logger {
	return &lumberjack.Logger{
		MaxSize:    1024,
		MaxAge:     7,
		MaxBackups: 3,
		LocalTime:  true,
		Compress:   true,
	}
}

// NewZapStdoutEncoderConfig creates new stdout encoder config
func NewZapStdoutEncoderConfig() *zapcore.EncoderConfig {
	return &zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// NewZapStdoutConfig creates new stdout config
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

// NewZapLoggerWithBytes inits zap logger with byte array from content of config file
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

// NewZapLoggerWithConfPath init zap logger with config file path
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

// NewZapLoggerWithConf inits zap logger with config
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
		if config.OutputPaths[i] != "stdout" && config.OutputPaths[i] != "stderr" {
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
			stdout, close, err := zap.Open(config.OutputPaths[i])
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
	errSink := make([]zapcore.WriteSyncer, 0, 0)
	if len(config.ErrorOutputPaths) > 0 {
		for i := range config.ErrorOutputPaths {
			if config.ErrorOutputPaths[i] != "stdout" && config.ErrorOutputPaths[i] != "stderr" {
				lumberNew := &lumberjack.Logger{
					Filename:   config.ErrorOutputPaths[i],
					MaxAge:     lumber.MaxAge,
					MaxBackups: lumber.MaxBackups,
					MaxSize:    lumber.MaxSize,
					Compress:   lumber.Compress,
					LocalTime:  lumber.LocalTime,
				}

				errSink = append(errSink, zapcore.AddSync(lumberNew))
			} else {
				stdout, close, err := zap.Open(config.ErrorOutputPaths[i])
				// just close the syncer if err occurs
				if err != nil {
					close()
				} else {
					errSink = append(errSink, stdout)
				}
			}
		}

		opts = append(opts, zap.ErrorOutput(zap.CombineWriteSyncers(errSink...)))
	}

	return zap.New(core, opts...).With(initialFields...), nil
}

// NewLumberjackLoggerWithBytes inits lumberjack logger as write sync with raw byte array of config file
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

// NewLumberjackLoggerWithConfPath inits lumberjack logger as write sync with lumberjack config file path
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
	}

	// default is console encoding
	return zapcore.NewConsoleEncoder(config.EncoderConfig)
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

// TransformToZapConfig transforms wrapped zap config into zap.Config
func TransformToZapConfig(wrap *ZapConfigWrap) *zap.Config {
	if wrap == nil {
		return nil
	}

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

// TransformToZapConfigWrap unmarshals zap.config
func TransformToZapConfigWrap(config *zap.Config) *ZapConfigWrap {
	return &ZapConfigWrap{
		Level:             config.Level.String(),
		Development:       config.Development,
		DisableCaller:     config.DisableCaller,
		DisableStacktrace: config.DisableStacktrace,
		Sampling:          config.Sampling,
		Encoding:          config.Encoding,
		EncoderConfig:     config.EncoderConfig,
		OutputPaths:       config.OutputPaths,
		ErrorOutputPaths:  config.ErrorOutputPaths,
		InitialFields:     config.InitialFields,
	}
}

// Marshal zapcore.NameEncoder
func marshalZapNameEncoder(encoder zapcore.NameEncoder) string {
	switch encoder {
	default:
		return "full"
	}
}

// Marshal zapcore.CallerEncoder
func marshalZapCallerEncoder(encoder zapcore.CallerEncoder) string {
	switch reflect.ValueOf(encoder).Pointer() {
	case reflect.ValueOf(zapcore.FullCallerEncoder).Pointer():
		return "full"
	default:
		return "short"
	}
}

// Marshal zapcore.DurationEncoder
func marshalZapDurationEncoder(encoder zapcore.DurationEncoder) string {
	switch reflect.ValueOf(encoder).Pointer() {
	case reflect.ValueOf(zapcore.StringDurationEncoder).Pointer():
		return "string"
	case reflect.ValueOf(zapcore.NanosDurationEncoder).Pointer():
		return "nanos"
	case reflect.ValueOf(zapcore.MillisDurationEncoder).Pointer():
		return "ms"
	default:
		return "secs"
	}
}

// Marshal zapcore.TimeEncoder
func marshalZapTimeEncoder(encoder zapcore.TimeEncoder) string {
	switch reflect.ValueOf(encoder).Pointer() {
	case reflect.ValueOf(zapcore.RFC3339NanoTimeEncoder).Pointer():
		return "RFC3339Nano"
	case reflect.ValueOf(zapcore.RFC3339TimeEncoder).Pointer():
		return "RFC3339"
	case reflect.ValueOf(zapcore.ISO8601TimeEncoder).Pointer():
		return "ISO8601"
	case reflect.ValueOf(zapcore.EpochMillisTimeEncoder).Pointer():
		return "millis"
	case reflect.ValueOf(zapcore.EpochNanosTimeEncoder).Pointer():
		return "nanos"
	default:
		return "seconds"
	}
}

// Marshal zapcore.LevelEncoder
func marshalZapLevelEncoder(encoder zapcore.LevelEncoder) string {
	switch reflect.ValueOf(encoder).Pointer() {
	case reflect.ValueOf(zapcore.CapitalLevelEncoder).Pointer():
		return "capital"
	case reflect.ValueOf(zapcore.CapitalColorLevelEncoder).Pointer():
		return "capitalColor"
	case reflect.ValueOf(zapcore.LowercaseColorLevelEncoder).Pointer():
		return "color"
	default:
		return "lower"
	}
}

// ZapConfigWrap wraps zap config which copied from zap.Config
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

// MarshalJSON marshals ZapConfigWrap
func (wrap *ZapConfigWrap) MarshalJSON() ([]byte, error) {
	encoderWrap := &ZapEncoderConfigWrap{
		MessageKey:       wrap.EncoderConfig.MessageKey,
		LevelKey:         wrap.EncoderConfig.LevelKey,
		TimeKey:          wrap.EncoderConfig.TimeKey,
		NameKey:          wrap.EncoderConfig.NameKey,
		CallerKey:        wrap.EncoderConfig.CallerKey,
		FunctionKey:      wrap.EncoderConfig.FunctionKey,
		StacktraceKey:    wrap.EncoderConfig.StacktraceKey,
		LineEnding:       wrap.EncoderConfig.LineEnding,
		EncodeLevel:      marshalZapLevelEncoder(wrap.EncoderConfig.EncodeLevel),
		EncodeTime:       marshalZapTimeEncoder(wrap.EncoderConfig.EncodeTime),
		EncodeDuration:   marshalZapDurationEncoder(wrap.EncoderConfig.EncodeDuration),
		EncodeCaller:     marshalZapCallerEncoder(wrap.EncoderConfig.EncodeCaller),
		EncodeName:       marshalZapNameEncoder(wrap.EncoderConfig.EncodeName),
		ConsoleSeparator: wrap.EncoderConfig.ConsoleSeparator,
	}

	// Create an inner config since zap.EncoderConfig would throw an error while marshalling
	type innerZapConfig struct {
		Level             string                 `json:"level" yaml:"level"`
		Development       bool                   `json:"development" yaml:"development"`
		DisableCaller     bool                   `json:"disableCaller" yaml:"disableCaller"`
		DisableStacktrace bool                   `json:"disableStacktrace" yaml:"disableStacktrace"`
		Sampling          *zap.SamplingConfig    `json:"sampling" yaml:"sampling"`
		Encoding          string                 `json:"encoding" yaml:"encoding"`
		EncoderConfig     *ZapEncoderConfigWrap  `json:"encoderConfig" yaml:"encoderConfig"`
		OutputPaths       []string               `json:"outputPaths" yaml:"outputPaths"`
		ErrorOutputPaths  []string               `json:"errorOutputPaths" yaml:"errorOutputPaths"`
		InitialFields     map[string]interface{} `json:"initialFields" yaml:"initialFields"`
	}

	return json.Marshal(&innerZapConfig{
		Level:             wrap.Level,
		Development:       wrap.Development,
		DisableCaller:     wrap.DisableCaller,
		DisableStacktrace: wrap.DisableStacktrace,
		Sampling:          wrap.Sampling,
		Encoding:          wrap.Encoding,
		EncoderConfig:     encoderWrap,
		OutputPaths:       wrap.OutputPaths,
		ErrorOutputPaths:  wrap.ErrorOutputPaths,
		InitialFields:     wrap.InitialFields,
	})
}

// UnmarshalJSON unmarshal ZapConfigWrap
func (wrap *ZapConfigWrap) UnmarshalJSON([]byte) error {
	return nil
}

// ZapEncoderConfigWrap wraps zap EncoderConfig which copied from zapcore.EncoderConfig
// This is used while parsing zap yaml config to zapcore.EncoderConfig with viper
// because Level would throw an error since it is not a type of string
type ZapEncoderConfigWrap struct {
	MessageKey       string `json:"messageKey" yaml:"messageKey"`
	LevelKey         string `json:"levelKey" yaml:"levelKey"`
	TimeKey          string `json:"timeKey" yaml:"timeKey"`
	NameKey          string `json:"nameKey" yaml:"nameKey"`
	CallerKey        string `json:"callerKey" yaml:"callerKey"`
	FunctionKey      string `json:"functionKey" yaml:"functionKey"`
	StacktraceKey    string `json:"stacktraceKey" yaml:"stacktraceKey"`
	LineEnding       string `json:"lineEnding" yaml:"lineEnding"`
	EncodeLevel      string `json:"levelEncoder" yaml:"levelEncoder"`
	EncodeTime       string `json:"timeEncoder" yaml:"timeEncoder"`
	EncodeDuration   string `json:"durationEncoder" yaml:"durationEncoder"`
	EncodeCaller     string `json:"callerEncoder" yaml:"callerEncoder"`
	EncodeName       string `json:"nameEncoder" yaml:"nameEncoder"`
	ConsoleSeparator string `json:"consoleSeparator" yaml:"consoleSeparator"`
}
