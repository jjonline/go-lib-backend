package crawler

import (
	"github.com/chromedp/chromedp/device"
	"regexp"
	"time"
)

var (
	// PcEmulate Pc版设备仿真参数
	PcEmulate = device.Info{
		Name:      "chrome",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
		Width:     1512, // 14寸mac默认宽度
		Height:    982,
		Scale:     1.000000,
		Landscape: false,
		Mobile:    false,
		Touch:     false,
	}
	// H5Emulate H5版设备仿真参数，也可以使用 device.IPhone13 之类的常量，但是无法自定义UA头
	H5Emulate = device.Info{
		Name:      "iPhone 13",
		UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1",
		Width:     390,
		Height:    844,
		Scale:     3.000000,
		Landscape: false,
		Mobile:    true,
		Touch:     true,
	}
	// 等待渲染完成的最大超时时长
	defaultRenderTimeout = 30 * time.Second
	// 判断UA头为H5的正则
	mobileUA = regexp.MustCompile("(?i)Android|Windows Phone|iPhone|iPod|Mobile|WhatsApp")
)
