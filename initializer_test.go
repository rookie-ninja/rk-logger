// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_logger

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"os"
	"path"
	"testing"
)

func TestConfigFileType_Indexing(t *testing.T) {
	assert.Equal(t, FileType(0), JSON)
	assert.Equal(t, FileType(1), YAML)
	assert.Equal(t, FileType(2), TOML)
	assert.Equal(t, FileType(3), HCL)
}

func TestConfigFileType_String_HappyCase(t *testing.T) {
	assert.Equal(t, "JSON", JSON.String())
	assert.Equal(t, "YAML", YAML.String())
	assert.Equal(t, "TOML", TOML.String())
	assert.Equal(t, "HCL", HCL.String())
}

func TestConfigFileType_String_Overflow_LeftBoundary(t *testing.T) {
	assert.Equal(t, "UNKNOWN", FileType(-1).String())
}

func TestConfigFileType_String_Overflow_RightBoundary(t *testing.T) {
	assert.Equal(t, "UNKNOWN", FileType(4).String())
}

// With nil byte array
func TestNewZapLoggerWithBytes_With_NilByteArray(t *testing.T) {
	logger, config, err := NewZapLoggerWithBytes(nil, YAML, nil)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With empty byte array
func TestNewZapLoggerWithBytes_With_EmptyByteArray(t *testing.T) {
	logger, config, err := NewZapLoggerWithBytes(make([]byte, 0, 0), YAML, nil)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With invalid json
func TestNewZapLoggerWithBytes_With_InvalidJson(t *testing.T) {
	invalidJson := `{"key":"value"`
	logger, config, err := NewZapLoggerWithBytes([]byte(invalidJson), JSON, nil)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With invalid yaml
func TestNewZapLoggerWithBytes_With_InvalidYaml(t *testing.T) {
	invalidYaml := `"key"="value"`
	logger, config, err := NewZapLoggerWithBytes([]byte(invalidYaml), YAML, nil)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With invalid toml
func TestNewZapLoggerWithBytes_With_InvalidToml(t *testing.T) {
	invalidToml := `key: "value"`
	logger, config, err := NewZapLoggerWithBytes([]byte(invalidToml), TOML, nil)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With invalid toml
func TestNewZapLoggerWithBytes_With_InvalidHCL(t *testing.T) {
	invalidHCL := `"key" : "value"`
	logger, config, err := NewZapLoggerWithBytes([]byte(invalidHCL), HCL, nil)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With unmatched type
func TestNewZapLoggerWithBytes_With_InvalidType(t *testing.T) {
	json := `{"key":"value"}`
	logger, config, err := NewZapLoggerWithBytes([]byte(json), 10, nil)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With unmatched type
func TestNewZapLoggerWithBytes_With_UnmatchedType(t *testing.T) {
	json := `{"key":"value"}`
	logger, config, err := NewZapLoggerWithBytes([]byte(json), YAML, nil)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With invalid zap config
func TestNewZapLoggerWithBytes_With_InvalidZapConfig(t *testing.T) {
	logger, config, err := NewZapLoggerWithBytes([]byte(`{}`), JSON, nil)
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
      }
    }`)
	logger, config, err := NewZapLoggerWithBytes(bytes, JSON, nil)
	assert.NotNil(t, logger)
	assert.NotNil(t, config)
	assert.Nil(t, err)
}

// With empty file path
func TestNewZapLoggerWithConfPath_With_EmptyString(t *testing.T) {
	logger, config, err := NewZapLoggerWithConfPath("", YAML, nil)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With invalid file path
func TestNewZapLoggerWithConfPath_With_InvalidFilePath(t *testing.T) {
	logger, config, err := NewZapLoggerWithConfPath("///invalid", YAML, nil)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// With non exist file path
func TestNewZapLoggerWithConfPath_With_NonExistFilePath(t *testing.T) {
	logger, config, err := NewZapLoggerWithConfPath("/NonExistExpected.invalid", YAML, nil)
	assert.Nil(t, logger)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}

// Happy case
func TestNewZapLoggerWithConfPath_HappyCase(t *testing.T) {
	// get current working directory
	dir, err := os.Getwd()
	assert.Nil(t, err)

	logger, config, err := NewZapLoggerWithConfPath(dir+"/assets/zap.yaml", YAML, nil)
	assert.NotNil(t, logger)
	assert.NotNil(t, config)
	assert.Nil(t, err)
}

// With nil config
func TestNewZapLoggerWithConf_With_NilConfig(t *testing.T) {
	logger, err := NewZapLoggerWithConf(nil, nil)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// Happy case
func TestNewZapLoggerWithConf_HappyCae(t *testing.T) {
	// get current working directory
	dir, err := os.Getwd()
	assert.Nil(t, err)
	// create zap config with existing config file
	_, config, _ := NewZapLoggerWithConfPath(dir+"/assets/zap.yaml", YAML, nil)

	logger, err := NewZapLoggerWithConf(config, nil)
	assert.NotNil(t, logger)
	assert.Nil(t, err)
}

// With nil lumberjack config
func TestNewLumberjackLoggerWithBytes_With_NilByteArray(t *testing.T) {
	logger, err := NewLumberjackLoggerWithBytes(nil, YAML)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// With empty lumberjack config
func TestNewLumberjackLoggerWithBytes_With_EmptyByteArray(t *testing.T) {
	logger, err := NewLumberjackLoggerWithBytes(make([]byte, 0, 0), YAML)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// With invalid yaml
func TestNewLumberjackLoggerWithBytes_With_InvalidYaml(t *testing.T) {
	logger, err := NewLumberjackLoggerWithBytes([]byte("key=value"), YAML)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// With invalid json
func TestNewLumberjackLoggerWithBytes_With_InvalidJson(t *testing.T) {
	logger, err := NewLumberjackLoggerWithBytes([]byte(`{"key":"value"`), JSON)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
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
func TestNewLumberjackLoggerWithConfPath_With_EmptyString(t *testing.T) {
	logger, err := NewLumberjackLoggerWithConfPath("", YAML)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// With invalid file path
func TestNewLumberjackLoggerWithConfPath_With_InvalidFilePath(t *testing.T) {
	logger, err := NewLumberjackLoggerWithConfPath("///invalid", YAML)
	assert.Nil(t, logger)
	assert.NotNil(t, err)
}

// With non exist file path
func TestNewLumberjackLoggerWithConfPath_With_NonExistFilePath(t *testing.T) {
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
func TestValidateFilePath_With_InvalidFilePath(t *testing.T) {
	assert.NotNil(t, validateFilePath("///invalid"))
}

// With non exist file path
func TestValidateFilePath_With_NonExistFilePath(t *testing.T) {
	assert.NotNil(t, validateFilePath("/NonExistExpected.invalid"))
}

// Happy case
func TestValidateFilePath_HappyCase(t *testing.T) {
	dir, err := os.Getwd()
	assert.Nil(t, err)
	assert.Nil(t, validateFilePath(dir+"/assets/lumberjack.yaml"))
}

// With json encoder
func TestGenerateEncoder_With_JsonEncoder(t *testing.T) {
	config := &zap.Config{Encoding: "json"}
	assert.NotNil(t, generateEncoder(config))
}

// With console encoder
func TestGenerateEncoder_With_ConsoleEncoder(t *testing.T) {
	config := &zap.Config{Encoding: "console"}
	assert.NotNil(t, generateEncoder(config))
}

// Absolute path
func TestToAbsoluteWorkingDir_With_AbsolutePath(t *testing.T) {
	abs, err := toAbsoluteWorkingDir("/tmp")
	assert.Nil(t, err)
	assert.True(t, path.IsAbs(abs))
}

// Relative path
func TestToAbsoluteWorkingDir_With_RelativePath(t *testing.T) {
	abs, err := toAbsoluteWorkingDir("logs/rk-logger.log")
	assert.Nil(t, err)
	assert.True(t, path.IsAbs(abs))
}
