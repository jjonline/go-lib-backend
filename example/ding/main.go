package main

import (
	"fmt"
	"github.com/jjonline/go-lib-backend/ding"
)

func main() {
	token := "***" // please set your token
	secret := "***" // please set your secret
	dingding := ding.New(token, secret, true, nil)

	// text
	err := dingding.Text("hello world", nil, false)
	fmt.Println(err)

	// markdown
	msg := fmt.Sprintf("> Account: %s  \n> Msg: %s  \n> Time:  %s \n", "account", "msg", "2021-07-23")
	err = dingding.Markdown("login info", msg, nil, false)
	fmt.Println(err)

	// link
	err = dingding.Link(
		"link test",
		"link description",
		"https://github.com/jjonline/go-lib-backend",
		"https://ec-image.hk01ec.com/-QzQHc0IMS5YfmqYCrmKPybJwLw=/public/images/202107/ea77fc7b-61c6-4942-9461-5e7a601cdf3f.png",
	)
	fmt.Println(err)

	// ActionCard
	err = dingding.ActionCard(
		"link test",
		"![sc](https://ec-image.hk01ec.com/-QzQHc0IMS5YfmqYCrmKPybJwLw=/public/images/202107/ea77fc7b-61c6-4942-9461-5e7a601cdf3f.png)",
		"11阅读11",
		"https://github.com/jjonline/go-lib-backend",
		)
	fmt.Println(err)

	// ActionCardWithMultiBtn
	err = dingding.ActionCardWithMultiBtn(
		"link test for multi",
		"![sc](https://ec-image.hk01ec.com/-QzQHc0IMS5YfmqYCrmKPybJwLw=/public/images/202107/ea77fc7b-61c6-4942-9461-5e7a601cdf3f.png) \n ## test for card",
		[]ding.Btn{{"试一下", "https://github.com/jjonline/go-lib-backend"}, {"赞一下", "https://github.com/jjonline/go-lib1-backend"}},
		false,
	)
	fmt.Println(err)

	// FeedCard
	err = dingding.FeedCard(
		[]ding.Feed{{"test one", "https://github.com/jjonline/go-lib-backend", "https://ec-image.hk01ec.com/-QzQHc0IMS5YfmqYCrmKPybJwLw=/public/images/202107/ea77fc7b-61c6-4942-9461-5e7a601cdf3f.png"},{"test two", "https://developers.dingtalk.com/document/app/custom-robot-access#section-e4x-4y8-9k0", "https://ec-image.hk01ec.com/UbDhaWYEzJQhD8Vz_trLQeUYd7k=/public/images/202107/58f811e8-b355-41d3-b8aa-e6880e91be77.png"}},
	)
	fmt.Println(err)
}
