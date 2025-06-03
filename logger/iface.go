package logger

// IFace 日志接口定义, slog.Logger 已实现该interface
type IFace interface {
	// Debug debug级别输出的日志
	//   - msg 日志消息文本描述
	//   - keyValue 按顺序一个key一个value，len(keyValue)一定是偶数<注意0也是偶数>
	Debug(msg string, keyValue ...any)
	// Info info级别输出的日志
	//   - msg 日志消息文本描述
	//   - keyValue 按顺序一个key一个value，len(keyValue)一定是偶数<注意0也是偶数>
	Info(msg string, keyValue ...any)
	// Warn warn级别输出的日志
	//   - msg 日志消息文本描述
	//   - keyValue 按顺序一个key一个value，len(keyValue)一定是偶数<注意0也是偶数>
	Warn(msg string, keyValue ...any)
	// Error error级别输出的日志
	//   - msg 日志消息文本描述
	//   - keyValue 按顺序一个key一个value，len(keyValue)一定是偶数<注意0也是偶数>
	Error(msg string, keyValue ...any)
}
