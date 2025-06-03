package logger4gin

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/go-stack/stack"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// XRequestID 为每个请求分配的请求编号key和名称
// 1、优先从header头里读由nginx维护的并且转发过来的x-request-id
// 2、如果读取不到则使用当前纳秒时间戳字符串加前缀字符串
const (
	XRequestID          = "x-request-id"       // 请求ID名称
	XRequestIDPrefix    = "R"                  // 当使用纳秒时间戳作为请求ID时拼接的前缀字符串
	TextGinPanic        = "gin.panic.recovery" // gin panic日志标记
	TextGinRequest      = "gin.request"        // gin request请求日志标记
	TextGinResponseFail = "gin.response.fail"  // gin 业务层面失败响应日志标记
	TextGinPreflight    = "gin.preflight"      // gin preflight 预检options请求类型日志
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
			slog.Default().Error(
				TextGinPanic,
				"module", TextGinPanic,
				"url", ctx.Request.URL.Path,
				"request", string(httpRequest),
				"error", err,
				"stack", stack.Trace().TrimRuntime().String(),
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
//   - appendHandle 额外补充的自定义添加字段方法，可选参数，返回偶数键值对切片
func GinLogger(appendHandle func(ctx *gin.Context) []any) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		start := time.Now()

		// set XRequestID
		requestID := setRequestID(ctx)

		// +++++++++++++++++++++++++
		// 记录请求 body 体
		// Notice: http包里对*http.Request.Body这个Io是一次性读取，此处读取完需再次设置Body以便其他位置能顺利读取到参数内容
		// +++++++++++++++++++++++++
		bodyData := GetRequestBody(ctx, true)

		// executes at end
		ctx.Next()

		latencyTime := time.Now().Sub(start)

		fields := []any{
			"module", TextGinRequest,
			"ua", ctx.GetHeader("User-Agent"),
			"method", ctx.Request.Method,
			"req_id", requestID,
			"req_body", bodyData,
			"client_ip", ctx.ClientIP(),
			"url_path", ctx.Request.URL.Path,
			"url_query", ctx.Request.URL.RawQuery,
			"url", ctx.Request.URL.String(),
			"http_status", ctx.Writer.Status(),
			"duration", latencyTime,
		}

		// 额外自定义补充字段
		if appendHandle != nil {
			fields = append(fields, appendHandle(ctx)...)
		}

		if latencyTime.Seconds() > 0.5 {
			slog.Default().Warn(ctx.Request.URL.Path, fields...)
		} else {
			slog.Default().Info(ctx.Request.URL.Path, fields...)
		}
	}
}

// GinLogHttpFail gin框架失败响应日志处理
func GinLogHttpFail(ctx *gin.Context, err error) {
	if err != nil {
		slog.Default().Warn(
			TextGinResponseFail,
			"module", TextGinResponseFail,
			"ua", ctx.GetHeader("User-Agent"),
			"method", ctx.Request.Method,
			"req_id", GetRequestID(ctx),
			"client_ip", ctx.ClientIP(),
			"url_path", ctx.Request.URL.Path,
			"url_query", ctx.Request.URL.RawQuery,
			"url", ctx.Request.URL.String(),
			"http_status", ctx.Writer.Status(),
			"error", err,
			"stack", stack.Trace().TrimRuntime().String(),
		)
	}
}

// GinCors 为gin开启跨域功能<尽量通过nginx反代处理>
func GinCors(ctx *gin.Context) {
	var allowOrigin = "*"

	// detect origin
	if origin := ctx.Request.Header.Get("Origin"); origin != "" {
		allowOrigin = origin
	} else if referer := ctx.Request.Referer(); referer != "" {
		if ref, err := url.Parse(referer); err == nil {
			allowOrigin = ref.Scheme + "://" + ref.Host
		}
	}

	ctx.Header("Access-Control-Allow-Origin", allowOrigin)
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	ctx.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,X-App-Client,X-Requested-With,Authorization")
	ctx.Header("Access-Control-Allow-Methods", "GET,OPTIONS,POST,PUT,DELETE,PATCH")
	if ctx.Request.Method == http.MethodOptions {
		slog.Default().Debug(
			TextGinPreflight,
			"module", TextGinPreflight,
			"ua", ctx.GetHeader("User-Agent"),
			"method", ctx.Request.Method,
			"req_id", GetRequestID(ctx),
			"client_ip", ctx.ClientIP(),
			"url_path", ctx.Request.URL.Path,
			"url_query", ctx.Request.URL.RawQuery,
			"url", ctx.Request.URL.String(),
		)
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}
	ctx.Next()
}

// setRequestID 内部方法设置请求ID
func setRequestID(ctx *gin.Context) string {
	requestID := ctx.GetHeader(XRequestID)
	if requestID == "" {
		requestID = XRequestIDPrefix + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	ctx.Set(XRequestID, requestID)
	return requestID
}

// GetRequestID 暴露方法：读取当前请求ID
func GetRequestID(ctx *gin.Context) string {
	if reqId, exist := ctx.Get(XRequestID); exist {
		return reqId.(string)
	}
	return ""
}

// GetRequestBody 获取请求body体
//   - strip 是否要将JSON类型的body体去除反斜杠和大括号，以便于Es等不做深层字段解析而当做一个字符串
func GetRequestBody(ctx *gin.Context, strip bool) string {
	bodyData := ""

	// post、put、patch、delete四种类型请求记录body体
	if IsModifyMethod(ctx.Request.Method) {
		// 判断是否为JSON实体类型<application/json>，仅需要判断content-type包含/json字符串即可
		if strings.Contains(ctx.ContentType(), "/json") {
			buf, _ := io.ReadAll(ctx.Request.Body)
			bodyData = string(buf)
			_ = ctx.Request.Body.Close()
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(buf)) // 重要

			// strip json `\{}` to ignore transfer JSON struct
			if strip {
				bodyData = strings.Replace(bodyData, "\\", "", -1)
				bodyData = strings.Replace(bodyData, "{", "", -1)
				bodyData = strings.Replace(bodyData, "}", "", -1)
			}
		} else {
			_ = ctx.Request.ParseForm() // 尝试解析表单, 文件表单忽略
			bodyData = ctx.Request.PostForm.Encode()
		}
	}

	return bodyData
}

// IsModifyMethod 检查当前请求方式否为修改类型
//   - 即判断请求是否为post、put、patch、delete
func IsModifyMethod(method string) bool {
	return method == http.MethodPost ||
		method == http.MethodPut ||
		method == http.MethodPatch ||
		method == http.MethodDelete
}
