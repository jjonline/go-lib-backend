# logger4gin

> 基于标准库`log/slog`的`slog.Default()`实现gin请求日志功能支持

````
router = gin.New()

// set Logger Recovery And CORS
router.Use(logger4gin.GinLogger(), logger4gin.GinRecovery(), logger4gin.GinCors())

# helper

# 当响应出现error时快捷记录日志
logger4gin.GinLogResponseFail(ctx, err)

# 获取http请求体body内容，可选支持移除JSON结构
logger4gin.GetRequestBody(ctx, false)

# 获取http响应体body内容，可选支持移除JSON结构
logger4gin.GetResponseBody(ctx, false)

# 通过error判断是否为失去底层tcp链接导致--通过error的文本匹配实现
logger4gin.CauseByLostConnection(err)
````
