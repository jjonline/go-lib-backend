package logger4gin

import (
	"bytes"
	"errors"
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

// responseRecorder 自定义ResponseWriter用于捕获响应内容
type responseRecorder struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseRecorder) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseRecorder) WriteString(s string) (n int, err error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// GinLogger Gin中间件形式logger支持
//
// 底层使用slog.Default()，可搭配本仓库一同提供的 logger.Logger 自动实现日志级别，或slog.Default()获取到slog.Logger对象后自主重设级别
// slog.LevelInfo 及以上不记录response
// slog.LevelDebug 及以下记录response便于调试
func GinLogger() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		start := time.Now()

		// 设置响应记录器以捕获响应内容
		writer := &responseRecorder{
			ResponseWriter: ctx.Writer,
			body:           bytes.NewBufferString(""),
		}
		ctx.Writer = writer

		// set XRequestID
		requestID := setRequestID(ctx)

		// +++++++++++++++++++++++++
		// 记录请求 body 体
		// Notice: http包里对*http.Request.Body这个Io是一次性读取，此处读取完需再次设置Body以便其他位置能顺利读取到参数内容
		// +++++++++++++++++++++++++
		bodyData := GetRequestBody(ctx, true)

		// executes at end
		ctx.Next()

		// 存储响应内容到上下文
		ctx.Set("response_body", writer.body.String())
		ctx.Set("response_status", ctx.Writer.Status())

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

		// debug则记录response，可能会降低响应性能
		if slog.Default().Enabled(ctx, slog.LevelDebug) {
			fields = append(fields, "response", writer.body.String())
		}

		// 响应时间超过0.5秒则是warn级别
		if latencyTime.Seconds() > 0.5 {
			slog.Default().Warn(ctx.Request.URL.Path, fields...)
		} else {
			slog.Default().Info(ctx.Request.URL.Path, fields...)
		}
	}
}

// GinRecovery zap实现的gin-recovery日志中间件<gin.HandlerFunc的实现>
func GinRecovery() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// dump出http请求相关信息
				httpRequest, _ := httputil.DumpRequest(ctx.Request, false)

				// 检查是否为tcp管道中断错误：这样就没办法给客户端通知消息
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne.Err, &se) {
						if CauseByLostConnection(se) {
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
}

// GinCors 为gin开启跨域功能<尽量通过nginx反代处理>
//
//	-- specifyOrigin 指定0个或1个CORS的固定的origin（不建议固定）
func GinCors(specifyOrigin ...string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var allowOrigin = "*"
		if len(specifyOrigin) == 0 {
			// detect origin
			if origin := ctx.Request.Header.Get("Origin"); origin != "" {
				allowOrigin = origin
			} else if referer := ctx.Request.Referer(); referer != "" {
				if ref, err := url.Parse(referer); err == nil {
					allowOrigin = ref.Scheme + "://" + ref.Host
				}
			}
		} else {
			// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Reference/Headers/Access-Control-Allow-Origin
			// 自定义CORS的域名，则只能指定1个，多个不允许
			// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Guides/CORS/Errors/CORSMultipleAllowOriginNotAllowed
			allowOrigin = specifyOrigin[0]
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
}

// GinLogResponseFail gin框架失败响应日志处理
func GinLogResponseFail(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

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

// GetRes 获取响应内容和状态码
func GetRes(ctx *gin.Context, strip bool) (string, int) {
	if writer, ok := ctx.Writer.(*responseRecorder); ok {
		bodyData := writer.body.String()

		// strip json `\{}` to ignore transfer JSON struct
		if strip {
			bodyData = strings.Replace(bodyData, "\\", "", -1)
			bodyData = strings.Replace(bodyData, "{", "", -1)
			bodyData = strings.Replace(bodyData, "}", "", -1)
		}

		return bodyData, writer.Status()
	}
	return "", 0
}

// CauseByLostConnection 字符串匹配方式检查是否为断开连接导致出错
func CauseByLostConnection(err error) bool {
	if err == nil || "" == err.Error() {
		return false
	}

	needles := []string{
		"server has gone away",
		"no connection to the server",
		"lost connection",
		"is dead or not enabled",
		"error while sending",
		"decryption failed or bad record mac",
		"server closed the connection unexpectedly",
		"ssl connection has been closed unexpectedly",
		"error writing data to the connection",
		"resource deadlock avoided",
		"transaction() on null",
		"child connection forced to terminate due to client_idle_limit",
		"query_wait_timeout",
		"reset by peer",
		"broken pipe",
		"connection refused",
	}

	msg := strings.ToLower(err.Error())
	for _, needle := range needles {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

// IsModifyMethod 检查当前请求方式否为修改类型
//   - 即判断请求是否为post、put、patch、delete
func IsModifyMethod(method string) bool {
	return method == http.MethodPost ||
		method == http.MethodPut ||
		method == http.MethodPatch ||
		method == http.MethodDelete
}
