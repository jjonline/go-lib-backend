package logger

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"
)

// XRequestID 为每个请求分配的请求编号key和名称
// 1、优先从header头里读由nginx维护的并且转发过来的x-request-id
// 2、如果读取不到则使用当前纳秒时间戳字符串
const (
	XRequestID          = "x-request-id"       // 请求ID名称
	TextGinPanic        = "gin.panic.recovery" // gin panic日志标记
	TextGinRequest      = "gin.request"        // gin request请求日志标记
	TextGinResponseFail = "gin.response.fail"  // gin 业务层面失败响应日志标记
)

// GinRecovery zap实现的gin-recovery日志中间件<gin.HandlerFunc的实现>
func GinRecovery(ctx *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			// dump出http请求相关信息
			httpRequest, _ := httputil.DumpRequest(ctx.Request, false)

			// 检查是否为tcp管道中断错误：这样就没办法给客户端通知消息
			var brokenPipe bool
			if ne, ok := err.(*net.OpError); ok {
				if se, ok := ne.Err.(*os.SyscallError); ok {
					if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
						brokenPipe = true
					}
				}
			}

			// record log
			zapLogger.Error(
				TextGinPanic,
				zap.String("module", TextGinPanic),
				zap.String("url", ctx.Request.URL.Path),
				zap.String("request", string(httpRequest)),
				zap.Any("error", err),
				zap.Stack("stack"),
			)

			if brokenPipe {
				// tcp中断导致panic，终止无输出
				_ = ctx.Error(err.(error))
				ctx.Abort()
			} else {
				// 非tcp中断导致panic，响应500错误
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}
	}()
	ctx.Next()
}

// GinLogger zap实现的gin-logger日志中间件<gin.HandlerFunc的实现>
func GinLogger(ctx *gin.Context) {
	start := time.Now()

	// set XRequestID
	requestID := setRequestID(ctx)

	// +++++++++++++++++++++++++
	// 记录请求 body 体
	// Notice: http包里对*http.Request.Body这个Io是一次性读取，此处读取完需再次设置Body以便其他位置能顺利读取到参数内容
	// +++++++++++++++++++++++++
	bodyData := GetRequestBody(ctx)

	// executes at end
	ctx.Next()

	latencyTime := time.Now().Sub(start)
	fields := []zap.Field{
		zap.String("module", TextGinRequest),
		zap.String("ua", ctx.GetHeader("User-Agent")),
		zap.String("method", ctx.Request.Method),
		zap.String("req_id", requestID),
		zap.String("req_body", bodyData),
		zap.String("client_ip", ctx.ClientIP()),
		zap.String("url_path", ctx.Request.URL.Path),
		zap.String("url_query", ctx.Request.URL.RawQuery),
		zap.String("url", ctx.Request.URL.String()),
		zap.Int("http_status", ctx.Writer.Status()),
		zap.Duration("duration", latencyTime),
	}

	if latencyTime.Seconds() > 0.5 {
		zapLogger.Warn(ctx.Request.URL.Path, fields...)
	} else {
		zapLogger.Info(ctx.Request.URL.Path, fields...)
	}
}

// GinLogHttpFail gin框架失败响应日志处理
func GinLogHttpFail(ctx *gin.Context, err error) {
	if err != nil && zapLogger.Core().Enabled(zap.InfoLevel) {
		zapLogger.Warn(
			TextGinResponseFail,
			zap.String("module", TextGinResponseFail),
			zap.String("ua", ctx.GetHeader("User-Agent")),
			zap.String("method", ctx.Request.Method),
			zap.String("req_id", GetRequestID(ctx)),
			zap.String("client_ip", ctx.ClientIP()),
			zap.String("url_path", ctx.Request.URL.Path),
			zap.String("url_query", ctx.Request.URL.RawQuery),
			zap.String("url", ctx.Request.URL.String()),
			zap.Int("http_status", ctx.Writer.Status()),
			zap.Error(err),
			zap.StackSkip("stack", 2),
		)
	}
}

// GinCors 为gin开启跨域功能<尽量通过nginx反代处理>
func GinCors(ctx *gin.Context)  {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,App-Client,x-requested-with,Authorization")
	ctx.Header("Access-Control-Allow-Methods", "GET,OPTIONS,POST,PUT,DELETE,PATCH")
	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}
	ctx.Next()
}

// setRequestID 内部方法设置请求ID
func setRequestID(ctx *gin.Context) string {
	requestID := ctx.GetHeader(XRequestID)
	if requestID == "" {
		requestID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	ctx.Set(XRequestID, requestID)
	return requestID
}

// GetRequestID 暴露方法：读取当前请求ID
func GetRequestID(ctx *gin.Context) string {
	if req_id, exist := ctx.Get(XRequestID); exist {
		return req_id.(string)
	}
	return ""
}

// GetRequestBody 获取请求body体
func GetRequestBody(ctx *gin.Context) string {
	bodyData := ""

	// post、put、patch、delete四种类型请求记录body提
	if ctx.Request.Method == http.MethodPost ||
		ctx.Request.Method == http.MethodPut ||
		ctx.Request.Method == http.MethodPatch ||
		ctx.Request.Method == http.MethodDelete {
		if ctx.ContentType() == "application/json" {
			buf, _ := ioutil.ReadAll(ctx.Request.Body)
			bodyData = string(buf)
			_ = ctx.Request.Body.Close()
			ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(buf)) // 重要

			// strip json `\{}` to ignore transfer JSON struct
			bodyData = strings.Replace(bodyData, "\\", "", -1)
			bodyData = strings.Replace(bodyData, "{", "", -1)
			bodyData = strings.Replace(bodyData, "}", "", -1)
		} else {
			_ = ctx.Request.ParseForm() // 尝试解析表单, 文件表单忽略
			bodyData = ctx.Request.PostForm.Encode()
		}
	}

	return bodyData
}
