package feishu

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jjonline/go-lib-backend/guzzle"
	"net/http"
	"strings"
	"time"
)

// messageURL 钉钉消息api
var (
	messageURL = "https://open.feishu.cn/open-apis/bot/v2/hook/"
)

// 飞书机器人文档地址：https://www.feishu.cn/hc/zh-CN/articles/360024984973
// 飞书机器人文档地址：https://open.feishu.cn/document/ukTMukTMukTM/ucTM5YjL3ETO24yNxkjN
// 飞书机器人markdown说明：https://open.feishu.cn/document/ukTMukTMukTM/uADOwUjLwgDM14CM4ATN

// Robot 飞书机器人结构体
type Robot struct {
	token  string         // 令牌
	secret string         // 秘钥
	client *guzzle.Client // guzzle客户端
	enable bool           // 开关函数，每次发送消息时触发：true-发送，false-不发送
}

var (
	UTCZone8         = "Asia/Hong_Kong"
	UTCZone8Location = time.FixedZone(UTCZone8, 8*3600)
)

// New 实例化消息发送对象
//   - token  飞书access_token，飞书机器人设置时 Webhook 的URL里的hook后方UUID形式的值
//   - secret 飞书secret，飞书机器人设置时 启用加签获得的秘钥令牌
//   - enable 开关，true则真实发送 false则不真实发送<不用更改注释调用代码仅初始化时设置该值即可关闭真实发送逻辑>
//   - client 自定义 *http.Client 可自主控制http请求客户端，给 nil 不则使用默认
func New(token, secret string, enable bool, httpClient *http.Client) *Robot {
	return &Robot{
		token:  token,
		secret: secret,
		client: guzzle.New(httpClient, nil),
		enable: enable,
	}
}

// Info 提示（标题蓝色背景）
// title和content均支持emoji表情
// markdown写法说明：https://open.feishu.cn/document/ukTMukTMukTM/uADOwUjLwgDM14CM4ATN
func (s *Robot) Info(ctx context.Context, title, markdownText string, t time.Time) (err error) {
	now := time.Now().Unix()
	sign, err := s.sign(now)
	if err != nil {
		return fmt.Errorf("sign err:%s", err.Error())
	}

	params := s.buildParams(BgGreen, title, strings.TrimRight(markdownText, "\n")+
		"\nTime: "+t.In(UTCZone8Location).Format("2006-01-02 15:04:05"))
	params.Sign = sign
	params.Timestamp = now
	return s.send(ctx, params)
}

// Warning 告警（标题黄色背景）
// title和content均支持emoji表情
// markdown写法说明：https://open.feishu.cn/document/ukTMukTMukTM/uADOwUjLwgDM14CM4ATN
func (s *Robot) Warning(ctx context.Context, title, markdownText string, t time.Time) (err error) {
	now := time.Now().Unix()
	sign, err := s.sign(now)
	if err != nil {
		return fmt.Errorf("sign err:%s", err.Error())
	}

	params := s.buildParams(BgYellow, title, strings.TrimRight(markdownText, "\n")+
		"\nTime: "+t.In(UTCZone8Location).Format("2006-01-02 15:04:05"))
	params.Sign = sign
	params.Timestamp = now
	return s.send(ctx, params)
}

// Error 错误（标题红色背景）
// title和content均支持emoji表情
// markdown写法说明：https://open.feishu.cn/document/ukTMukTMukTM/uADOwUjLwgDM14CM4ATN
func (s *Robot) Error(ctx context.Context, title, markdownText string, t time.Time) (err error) {
	now := time.Now().Unix()
	sign, err := s.sign(now)
	if err != nil {
		return fmt.Errorf("sign err:%s", err.Error())
	}

	params := s.buildParams(BgRed, title, strings.TrimRight(markdownText, "\n")+
		"\nTime: "+t.In(UTCZone8Location).Format("2006-01-02 15:04:05"))
	params.Sign = sign
	params.Timestamp = now
	return s.send(ctx, params)
}

func (s *Robot) buildParams(bg, title, markdownText string) CardMsgParams {
	return CardMsgParams{
		MsgType: Interactive,
		Card: CardItem{
			Config: CardConfigItem{
				WideScreenMode: true,
				EnableForward:  true,
			},
			Header: CardHeaderItem{
				Title: CardHeaderTitleItem{
					Content: title,
					Tag:     "plain_text",
				},
				Template: bg,
			},
			Elements: []CardElementItem{
				{
					Tag:     "markdown",
					Content: markdownText,
				},
			},
		},
	}
}

// send 发送
func (s *Robot) send(ctx context.Context, params CardMsgParams) (err error) {
	if !s.enable {
		return
	}

	result, err := s.client.PostJSON(ctx, messageURL+s.token, params, nil)
	if err != nil {
		return
	}

	var resp SendResponse
	if err = json.Unmarshal(result.Body, &resp); err != nil {
		return
	}

	if resp.StatusCode == 0 && resp.Code == 0 {
		return nil
	}
	return fmt.Errorf("send msg err:(%d)%s", resp.Code, resp.Msg)
}

// sign 签名：timestamp + key 做sha256, 再进行base64 encode
func (s *Robot) sign(timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + s.secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}
