package qywx

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jjonline/go-lib-backend/guzzle"
	"net/http"
	"net/url"
)

// messageURL 钉钉消息api
var (
	messageURL = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send"
)

// WeWork 企业微信机器人结构
type WeWork struct {
	key    string
	client *guzzle.Client
	enable bool
}

// response 响应结构
type response struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// Article 图文消息结构
type Article struct {
	Title       string `json:"title" comment:"必填，标题，不超过128个字节，超过会自动截断"`
	Description string `json:"description,omitempty" comment:"选填，描述，不超过512个字节，超过会自动截断"`
	URL         string `json:"url" comment:"必填，点击后跳转的链接"`
	PicUrl      string `json:"picurl,omitempty" comment:"选填，图文消息的图片链接，支持JPG、PNG格式，较好的效果为大图 1068*455，小图150*150"`
}

// New 创建企业微信机器人客户端 - 20条/分钟
//   - key    企业微信机器人key，企业微信机器人设置时 Webhook 的URL里的key值
//   - enable 开关，true则真实发送 false则不真实发送<不用更改注释调用代码仅初始化时设置该值即可关闭真实发送逻辑>
//   - client 自定义 *http.Client 可自主控制http请求客户端，给 nil 则使用默认
func New(key string, enable bool, client *http.Client) *WeWork {
	return &WeWork{
		key:    key,
		client: guzzle.New(client),
		enable: enable,
	}
}

// send 底层执行发送方法
func (w *WeWork) send(message interface{}) error {
	params := url.Values{}
	params.Set("key", w.key)

	res, err := w.client.PostJSON(context.TODO(), guzzle.ToQueryURL(messageURL, params), message, nil)
	if err != nil {
		return err
	}

	// check response
	body := response{}
	err = json.Unmarshal(res.Body, &body)
	if err != nil {
		return err
	}
	if body.ErrCode != 0 {
		return fmt.Errorf("%s", body.ErrMsg)
	}

	return nil
}

// Text 发送文本信息
//   - msg       文本消息内容，不宜过长，最大2048字节
//   - atMobiles 需要 at 的人的手机号，不需要@任何人时给nil
//   - atUserIds 需要 at 的人的企业微信用户id，不需要@任何人时给nil，要@全员使用 @all 即可
func (w *WeWork) Text(msg string, atMobiles []string, atUserIds []string) error {
	if !w.enable {
		return nil
	}

	message := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content":               msg,
			"mentioned_list":        atUserIds,
			"mentioned_mobile_list": atMobiles,
		},
	}
	return w.send(message)
}

// Markdown 发送markdown信息
//   - markDownText markdown格式文本
func (w *WeWork) Markdown(markDownText string) error {
	if !w.enable {
		return nil
	}
	message := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": markDownText,
		},
	}
	return w.send(message)
}

// Image 发送图片信息
//   - base64image 图片base64编码后字符串
//   - imageMd5    图片原文件md5
func (w *WeWork) Image(base64image, imageMd5 string) error {
	if !w.enable {
		return nil
	}
	message := map[string]interface{}{
		"msgtype": "image",
		"markdown": map[string]string{
			"image":  base64image,
			"base64": imageMd5,
		},
	}
	return w.send(message)
}

// News 发送多图文混合类型信息
//   - article 多条图文切片，1到8条图文--即最多8条
func (w *WeWork) News(article []Article) error {
	if !w.enable {
		return nil
	}
	message := map[string]interface{}{
		"msgtype": "news",
		"news": map[string][]Article{
			"articles": article,
		},
	}
	return w.send(message)
}
