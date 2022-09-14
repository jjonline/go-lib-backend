package feishu

// 消息类型
const (
	Text        = "text"        // 文本消息
	Interactive = "interactive" // 卡片消息
)

// 卡片标题颜色
const (
	BgGreen  = "green"  // 绿色
	BgYellow = "yellow" // 黄色
	BgRed    = "red"    // 红色
)

type CardMsgParams struct {
	Timestamp int64    `json:"timestamp"`
	Sign      string   `json:"sign"`
	MsgType   string   `json:"msg_type"`
	Card      CardItem `json:"card"`
}

type CardItem struct {
	Config   CardConfigItem    `json:"config"`
	Header   CardHeaderItem    `json:"header"`
	Elements []CardElementItem `json:"elements"`
}

type CardConfigItem struct {
	WideScreenMode bool `json:"wide_screen_mode"` // true
	EnableForward  bool `json:"enable_forward"`   // true
}

type CardHeaderItem struct {
	Title    CardHeaderTitleItem `json:"title"`
	Template string              `json:"template"` // 卡片标题颜色：blue red
}

type CardHeaderTitleItem struct {
	Content string `json:"content"` // 卡片标题
	Tag     string `json:"tag"`     // plain_text
}

type CardElementItem struct {
	Tag     string `json:"tag"`     // markdown
	Content string `json:"content"` // markdown内容
}

type SendResponse struct {
	StatusCode    int    `json:"StatusCode"`    // 成功：StatusCode=0
	StatusMessage string `json:"StatusMessage"` // 成功：StatusMessage=success
	Code          int    `json:"code"`          // 失败错误码
	Msg           string `json:"msg"`           // 失败错误信息
}
