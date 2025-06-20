# logger

> 基于标准库`log/slog`的logger实现，写入文件形式的支持配置按文件体积、日期自动切割轮转。

````
logger := logger.New(&logger.Options{Target: os.Stdout})

// can render info log
logger.Info("info", "info", "testing")

// change log level
logger.GetSlogLeveler().Set(slog.LevelInfo)
// none log render
logger.Info("info", "info", "testing")

// high performance
logger.Info("info", slog.String("info", "testing"))

// get slog.Logger instance
slogLogger := logger.GetSlogLogger()

// get log.Logger instance
logger.GetLogLogger()
````
