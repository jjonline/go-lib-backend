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
const XRequestID = "x-request-id"

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
				"gin.panic.recovery",
				zap.String("module", "gin.panictx.recovery"),
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
	requestID := ctx.GetHeader(XRequestID)
	if requestID == "" {
		requestID = strconv.FormatInt(start.UnixNano(), 10)
	}
	ctx.Set(XRequestID, requestID)

	// +++++++++++++++++++++++++
	// 记录请求 body 体
	// Notice: http包里对*http.Request.Body这个Io是一次性读取，此处读取完需再次设置Body以便其他位置能顺利读取到参数内容
	// +++++++++++++++++++++++++
	bodyData := GetRequestBody(ctx)

	// executes at end
	ctx.Next()

	latencyTime := time.Now().Sub(start)
	fields := []zap.Field{
		zap.String("module", "request"),
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
