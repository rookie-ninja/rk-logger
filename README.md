# rk-logger
Log initializer written with golang.
Currently, support zap logger as default logger and lumberjack as log rotation

- [zap](https://github.com/uber-go/zap)
- [lumberjack](https://github.com/natefinch/lumberjack)

## Installation
`go get -u rookie-ninja/rk-logger`

## Quick Start
We combined zap config and lumberjack config in the same config file
Both of the configs could keep same format as it 

In order to init zap logger with full log rotation, rk-logger support three different utility functions
- With zap+lumberjack config file path
- With zap+lumberjack config as byte array
- With zap config and lumberjack config

### With Config file path
config:
```yaml
---
level: debug
encoding: console
outputPaths:
  - stdout
  - logs/rk-logger.log
errorOutputPaths:
  - stderr
initialFields:
  initFieldKey: fieldValue
encoderConfig:
  messageKey: messagea
  levelKey: level
  nameKey: logger
  timeKey: time
  callerKey: caller
  stacktraceKey: stacktrace
  callstackKey: callstack
  errorKey: error
  timeEncoder: iso8601
  fileKey: file
  levelEncoder: capital
  durationEncoder: second
  callerEncoder: full
  nameEncoder: full
  sampling:
    initial: '3'
    thereafter: '10'
maxsize: 1
maxage: 7
maxbackups: 3
localtime: true
compress: true
```

Example:
```go
func NewZapLoggerWithConfPathExample() {
    // get current working directory
    dir, _ := os.Getwd()

    // init logger 
    logger, _, _ := rk_logger.NewZapLoggerWithConfPath(path.Clean(path.Join(dir, "/assets/zap.yaml")), rk_logger.YAML)
    
    // use it 
    logger.Info("NewZapLoggerWithConfPathExample")
}
```
### With Config as byte array
Example:
```go
func NewZapLoggerWithBytesExample() {
    bytes := []byte(`{
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

    logger, _, err := rk_logger.NewZapLoggerWithBytes(bytes, rk_logger.JSON)
    
    logger.Info("NewZapLoggerWithBytesExample")
}
```
### With Config
```go
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
        Level: zap.NewAtomicLevelAt(zap.InfoLevel),
        EncoderConfig: encodingConfig,
        OutputPaths: []string{"stdout", "logs/rk-logger.log"},
    }

    logger, _ := rk_logger.NewZapLoggerWithConf(config, &lumberjack.Logger{})
    logger.Info("NewZapLoggerWithConfExample")
}
```

### Development Status: Stable

### Contributing
We encourage and support an active, healthy community of contributors &mdash;
including you! Details are in the [contribution guide](CONTRIBUTING.md) and
the [code of conduct](CODE_OF_CONDUCT.md). The rk maintainers keep an eye on
issues and pull requests, but you can also report any negative conduct to
dongxuny@gmail.com. That email list is a private, safe space; even the zap
maintainers don't have access, so don't hesitate to hold us to a high
standard.

<hr>

Released under the [MIT License](LICENSE).

