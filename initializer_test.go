// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rklogger

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"testing"
)

func TestNewZapLoggerWithOverride(t *testing.T) {
	// with json, output path
	logger, err := NewZapLoggerWithOverride("json", "logs/ut.log")
	assert.NotNil(t, logger)
	assert.Nil(t, err)

	// with console, output path
	logger, err = NewZapLoggerWithOverride("console", "logs/ut.log")
	assert.NotNil(t, logger)
	assert.Nil(t, err)

	// with invalid, output path
	logger, err = NewZapLoggerWithOverride("invalid", "logs/ut.log")
	assert.NotNil(t, logger)
	assert.Nil(t, err)

	// without output paths
	logger, err = NewZapLoggerWithOverride("invalid")
	assert.NotNil(t, logger)
	assert.Nil(t, err)
}

func TestConfigFileType_Indexing(t *testing.T) {
	assert.Equal(t, FileType(0), JSON)
	assert.Equal(t, FileType(1), YAML)
}

func TestConfigFileType_String_HappyCase(t *testing.T) {
	assert.Equal(t, "JSON", JSON.String())
	assert.Equal(t, "YAML", YAML.String())
}

func TestConfigFileType_String_Overflow_LeftBoundary(t *testing.T) {
	assert.Equal(t, "UNKNOWN", FileType(-1).String())
}

func TestConfigFileType_String_Overflow_RightBoundary(t *testing.T) {
	assert.Equal(t, "UNKNOWN", FileType(4).String())
}

func TestNewZapEventConfig_HappyCase(t *testing.T) {
	config := NewZapEventConfig()
	assert.NotNil(t, config)
}

func TestNewZapStdoutConfig_HappyCase(t *testing.T) {
	config := NewZapStdoutConfig()
	assert.NotNil(t, config)
	assert.Equal(t, zap.InfoLevel.String(), config.Level.String())
	assert.True(t, config.Development)
	assert.Equal(t, "console", config.Encoding)
	assert.True(t, config.DisableStacktrace)
	assert.NotNil(t, config.EncoderConfig)
	assert.Len(t, config.OutputPaths, 1)
	assert.Equal(t, "stdout", config.OutputPaths[0])
	assert.Len(t, config.ErrorOutputPaths, 1)
	assert.Equal(t, "stderr", config.ErrorOutputPaths[0])
}

// With nil byte array
func TestNewZapLoggerWithBytes_WithNilByteArray(t *testing.T) {
	logger, config, err := NewZapLoggerWithBytes(nil, YAML)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With empty byte array
func TestNewZapLoggerWithBytes_WithEmptyByteArray(t *testing.T) {
	logger, config, err := NewZapLoggerWithBytes(make([]byte, 0, 0), YAML)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With invalid json
func TestNewZapLoggerWithBytes_WithInvalidJson(t *testing.T) {
	invalidJson := `{"key":"value"`
	logger, config, err := NewZapLoggerWithBytes([]byte(invalidJson), JSON)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With invalid yaml
func TestNewZapLoggerWithBytes_WithInvalidYaml(t *testing.T) {
	invalidYaml := `"key"="value"`
	logger, config, err := NewZapLoggerWithBytes([]byte(invalidYaml), YAML)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With unmatched type
func TestNewZapLoggerWithBytes_WithInvalidType(t *testing.T) {
	json := `{"key":"value"}`
	logger, config, err := NewZapLoggerWithBytes([]byte(json), 10)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// Happy case
func TestNewZapLoggerWithBytes_HappyCase(t *testing.T) {
	bytes := []byte(`{
      "level": "debug",
      "encoding": "console",
      "outputPaths": ["stdout"],
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
	logger, config, err := NewZapLoggerWithBytes(bytes, JSON)
	assert.NotNil(t, logger)
	assert.NotNil(t, config)
	assert.Nil(t, err)
}

// With empty file path
func TestNewZapLoggerWithConfPath_WithEmptyString(t *testing.T) {
	logger, config, err := NewZapLoggerWithConfPath("", YAML)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With invalid file path
func TestNewZapLoggerWithConfPath_WithInvalidFilePath(t *testing.T) {
	logger, config, err := NewZapLoggerWithConfPath("///invalid", YAML)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With non exist file path
func TestNewZapLoggerWithConfPath_WithNonExistFilePath(t *testing.T) {
	logger, config, err := NewZapLoggerWithConfPath("/NonExistExpected.invalid", YAML)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// Happy case
func TestNewZapLoggerWithConfPath_HappyCase(t *testing.T) {
	// get current working directory
	dir, err := os.Getwd()
	assert.Nil(t, err)

	logger, config, err := NewZapLoggerWithConfPath(dir+"/assets/zap.yaml", YAML)
	assert.NotNil(t, logger)
	assert.NotNil(t, config)
	assert.Nil(t, err)
}

// With nil config
func TestNewZapLoggerWithConf_WithNilConfig(t *testing.T) {
	logger, err := NewZapLoggerWithConf(nil, nil)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// New lumberjack instance would be created if output path is one of stdout or stderr
func TestNewZaploggerWithConf_NonStdoutOutputPath(t *testing.T) {
	// get current working directory
	dir, err := os.Getwd()
	assert.Nil(t, err)
	// create zap config with existing config file
	_, config, _ := NewZapLoggerWithConfPath(dir+"/assets/zap.yaml", YAML)

	config.OutputPaths = []string{"ut.log"}

	logger, err := NewZapLoggerWithConf(config, LumberjackConfig)
	assert.NotNil(t, logger)
	assert.Nil(t, err)
}

// New lumberjack instance would be created if output path is one of stdout or stderr
func TestNewZaploggerWithConf_NonStdoutErrOutputPath(t *testing.T) {
	// get current working directory
	dir, err := os.Getwd()
	assert.Nil(t, err)
	// create zap config with existing config file
	_, config, _ := NewZapLoggerWithConfPath(dir+"/assets/zap.yaml", YAML)

	config.ErrorOutputPaths = []string{"ut-err.log"}

	logger, err := NewZapLoggerWithConf(config, LumberjackConfig)
	assert.NotNil(t, logger)
	assert.Nil(t, err)
}

// Happy case
func TestNewZapLoggerWithConf_HappyCae(t *testing.T) {
	// get current working directory
	dir, err := os.Getwd()
	assert.Nil(t, err)
	// create zap config with existing config file
	_, config, _ := NewZapLoggerWithConfPath(dir+"/assets/zap.yaml", YAML)

	logger, err := NewZapLoggerWithConf(config, nil)
	assert.NotNil(t, logger)
	assert.Nil(t, err)
}

// With nil lumberjack config
func TestNewLumberjackLoggerWithBytes_WithNilByteArray(t *testing.T) {
	logger, err := NewLumberjackLoggerWithBytes(nil, YAML)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// With empty lumberjack config
func TestNewLumberjackLoggerWithBytes_WithEmptyByteArray(t *testing.T) {
	logger, err := NewLumberjackLoggerWithBytes(make([]byte, 0, 0), YAML)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// With invalid yaml
func TestNewLumberjackLoggerWithBytes_WithInvalidYaml(t *testing.T) {
	logger, err := NewLumberjackLoggerWithBytes([]byte("key=value"), YAML)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// With invalid json
func TestNewLumberjackLoggerWithBytes_WithInvalidJson(t *testing.T) {
	logger, err := NewLumberjackLoggerWithBytes([]byte(`{"key":"value"`), JSON)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// With unknown file type
func TestNewLumberjackLoggerWithBytes_WithUnkownFileType(t *testing.T) {
	logger, err := NewLumberjackLoggerWithBytes([]byte(`{"key":"value"}`), 2)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

func TestTransformToZapConfigWrap(t *testing.T) {
	config := NewZapStdoutConfig()
	wrap := TransformToZapConfigWrap(config)

	assert.Equal(t, config.Level.String(), wrap.Level)
	assert.Equal(t, config.Development, wrap.Development)
	assert.Equal(t, config.DisableCaller, wrap.DisableCaller)
	assert.Equal(t, config.DisableStacktrace, wrap.DisableStacktrace)
	assert.Equal(t, config.Sampling, wrap.Sampling)
	assert.Equal(t, config.Encoding, wrap.Encoding)
	assert.Equal(t, config.OutputPaths, wrap.OutputPaths)
	assert.Equal(t, config.ErrorOutputPaths, wrap.ErrorOutputPaths)
	assert.Equal(t, config.InitialFields, wrap.InitialFields)
}

func TestMarshalZapNameEncoder(t *testing.T) {
	config := NewZapStdoutConfig()
	assert.Equal(t, "full", marshalZapNameEncoder(config.EncoderConfig.EncodeName))
}

func TestMarshalZapCallerEncoder(t *testing.T) {
	assert.Equal(t, "full", marshalZapCallerEncoder(zapcore.FullCallerEncoder))
	assert.Equal(t, "short", marshalZapCallerEncoder(zapcore.ShortCallerEncoder))
}

func TestMarshalZapDurationEncoder(t *testing.T) {
	assert.Equal(t, "string", marshalZapDurationEncoder(zapcore.StringDurationEncoder))
	assert.Equal(t, "nanos", marshalZapDurationEncoder(zapcore.NanosDurationEncoder))
	assert.Equal(t, "ms", marshalZapDurationEncoder(zapcore.MillisDurationEncoder))
	assert.Equal(t, "secs", marshalZapDurationEncoder(zapcore.SecondsDurationEncoder))
}

func TestMarshalZapTimeEncoder(t *testing.T) {
	assert.Equal(t, "RFC3339Nano", marshalZapTimeEncoder(zapcore.RFC3339NanoTimeEncoder))
	assert.Equal(t, "RFC3339", marshalZapTimeEncoder(zapcore.RFC3339TimeEncoder))
	assert.Equal(t, "ISO8601", marshalZapTimeEncoder(zapcore.ISO8601TimeEncoder))
	assert.Equal(t, "millis", marshalZapTimeEncoder(zapcore.EpochMillisTimeEncoder))
	assert.Equal(t, "nanos", marshalZapTimeEncoder(zapcore.EpochNanosTimeEncoder))
	assert.Equal(t, "seconds", marshalZapTimeEncoder(zapcore.EpochTimeEncoder))
}

func TestMarshalZapLevelEncoder(t *testing.T) {
	assert.Equal(t, "capital", marshalZapLevelEncoder(zapcore.CapitalLevelEncoder))
	assert.Equal(t, "capitalColor", marshalZapLevelEncoder(zapcore.CapitalColorLevelEncoder))
	assert.Equal(t, "color", marshalZapLevelEncoder(zapcore.LowercaseColorLevelEncoder))
	assert.Equal(t, "lower", marshalZapLevelEncoder(zapcore.LowercaseLevelEncoder))
}

// Happy case
func TestNewLumberjackLoggerWithBytes_HappyCase(t *testing.T) {
	bytes := []byte(`{
     "maxsize": 1,
     "maxage": 7,
     "maxbackups": 3,
     "localtime": true,
     "compress": true
    }`)

	logger, err := NewLumberjackLoggerWithBytes(bytes, JSON)
	assert.NotNil(t, logger)
	assert.Nil(t, err)
}

// With empty file path
func TestNewLumberjackLoggerWithConfPath_WithEmptyString(t *testing.T) {
	logger, err := NewLumberjackLoggerWithConfPath("", YAML)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// With invalid file path
func TestNewLumberjackLoggerWithConfPath_WithInvalidFilePath(t *testing.T) {
	logger, err := NewLumberjackLoggerWithConfPath("///invalid", YAML)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// With non exist file path
func TestNewLumberjackLoggerWithConfPath_WithNonExistFilePath(t *testing.T) {
	logger, err := NewLumberjackLoggerWithConfPath("/NonExistExpected.invalid", YAML)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// Happy case
func TestNewLumberjackLoggerWithConfPath_HappyCase(t *testing.T) {
	// get current working directory
	dir, err := os.Getwd()
	assert.Nil(t, err)

	logger, err := NewLumberjackLoggerWithConfPath(dir+"/assets/lumberjack.yaml", YAML)
	assert.NotNil(t, logger)
	assert.Nil(t, err)
}

// With invalid file path
func TestValidateFilePath_WithInvalidFilePath(t *testing.T) {
	assert.NotNil(t, validateFilePath("///invalid"))
}

// With non exist file path
func TestValidateFilePath_WithNonExistFilePath(t *testing.T) {
	assert.NotNil(t, validateFilePath("/NonExistExpected.invalid"))
}

// Happy case
func TestValidateFilePath_HappyCase(t *testing.T) {
	dir, err := os.Getwd()
	assert.Nil(t, err)
	assert.Nil(t, validateFilePath(dir+"/assets/lumberjack.yaml"))
}

// With json encoder
func TestGenerateEncoder_WithJsonEncoder(t *testing.T) {
	config := &zap.Config{Encoding: "json"}
	assert.NotNil(t, generateEncoder(config))
}

// With console encoder
func TestGenerateEncoder_WithConsoleEncoder(t *testing.T) {
	config := &zap.Config{Encoding: "console"}
	assert.NotNil(t, generateEncoder(config))
}

// Absolute path
func TestToAbsoluteWorkingDir_WithAbsolutePath(t *testing.T) {
	abs, err := toAbsoluteWorkingDir("/tmp")
	assert.Nil(t, err)
	assert.True(t, path.IsAbs(abs))
}

// Relative path
func TestToAbsoluteWorkingDir_WithRelativePath(t *testing.T) {
	abs, err := toAbsoluteWorkingDir("logs/rk-logger.log")
	assert.Nil(t, err)
	assert.True(t, path.IsAbs(abs))
}

func TestTransformToZapConfig_WithNilInput(t *testing.T) {
	assert.Nil(t, TransformToZapConfig(nil))
}

func TestTransformToZapConfig_WithInvalidLevel(t *testing.T) {
	wrap := &ZapConfigWrap{
		Level: "invalid",
	}

	zapConfig := TransformToZapConfig(wrap)
	assert.NotNil(t, zapConfig)
	assert.Equal(t, zap.InfoLevel, zapConfig.Level.Level())
}

func TestTransformToZapConfig_HappyCase(t *testing.T) {
	wrap := &ZapConfigWrap{
		Level:             "info",
		Development:       true,
		DisableCaller:     true,
		DisableStacktrace: true,
		Encoding:          "json",
		OutputPaths:       []string{"ut.log"},
		ErrorOutputPaths:  []string{"ut.log"},
	}

	zapConfig := TransformToZapConfig(wrap)
	assert.NotNil(t, zapConfig)
	assert.Equal(t, zap.InfoLevel, zapConfig.Level.Level())
	assert.True(t, zapConfig.Development)
	assert.True(t, zapConfig.DisableCaller)
	assert.True(t, zapConfig.DisableStacktrace)
	assert.Equal(t, "json", zapConfig.Encoding)
	assert.Contains(t, zapConfig.OutputPaths, "ut.log")
	assert.Contains(t, zapConfig.ErrorOutputPaths, "ut.log")
}

func TestZapConfigWrap_MarshalJSON_HappyCase(t *testing.T) {
	wrap := TransformToZapConfigWrap(NewZapStdoutConfig())

	bytes, err := wrap.MarshalJSON()
	assert.Nil(t, err)
	assert.NotNil(t, bytes)
}

func TestZapConfigWrap_UnmarshalJSON(t *testing.T) {
	wrap := TransformToZapConfigWrap(NewZapStdoutConfig())

	// unmarshal is not supported yet!
	assert.Nil(t, wrap.UnmarshalJSON([]byte{}))
}
