package main

import "rookie-ninja/rk-logger"

func main() {
	ExampleNewZapLoggerWithBytes()
}

func ExampleNewZapLoggerWithBytes() {
	zapBytes := []byte(`{
      "level": "debug",
      "encoding": "console",
      "outputPaths": ["stdout","full path one","full path two"],
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

	logger.Debug("This is a DEBUG message")
	logger.Info("This is a INFO message")
	logger.Warn("This is a WARN message")
}
