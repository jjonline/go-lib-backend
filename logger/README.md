# logger

> 基于标准库`log/slog`的logger封装，无任何第三方依赖

````
// 设置日志错误级别，使用 log/slog内置常量
lvl := &slog.LevelVar{}
lvl.Set(slog.LevelDebug)

// 实例化logger
tLogger := logger.New(lvl, "./runtime/", logger.Options{
    AddSource: false,           // 输出的日志是否添加code行数和位置
    UseText:   false,           // 日志格式仅支持：行文本 和 行JSON，默认是行JSON，如果要行文本该选项给true
    MaxSize:   2 * 1024 * 1024, // 文件日志时有效，单个文件最大体积，单位：byte
    MaxDays:   3,               // 文件日志时有效，文件日志时仅需配置文件存储路径，会自动按日期切割，配置保留日志天数
})

// 记录1条日志
tLogger.Info("info", "key", "value")

// 获取log/slog 实例
sLogger := tLogger.GetSlogLogger()
sLogger.Info("info", "key", "value")

// 获取底层writer
writer := tLogger.GetWriter()
````