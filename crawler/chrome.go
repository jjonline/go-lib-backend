package crawler

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"net/url"
	"os"
	"strings"
	"time"
)

// Chrome chromedp浏览器控制结构
type Chrome struct {
	// context & cancelFunc
	Ctx               context.Context    // 执行任务的context
	instanceContext   context.Context    // 启动chrome的context
	instanceCloseFunc context.CancelFunc // 关闭chrome的cancelFunc
	cancelCtxFunc     context.CancelFunc // 关闭的cancelFunc
	cancel            context.CancelFunc // baseContext cancelFunc

	// logger
	debugLogger func(string, ...interface{}) // debug调试
	errorLogger func(string, ...interface{}) // error异常
	infoLogger  func(string, ...interface{}) // 普通日志

	// setting
	ua              string            // UA头
	proxyUrl        string            // 代理地址，为空则无代理
	headless        bool              // 是否无头模式，即打开chrome是否要显示界面，默认false不显示
	emulate         *device.Info      // 自定义仿真参数，UA头宽高尺寸Touch特性等，默认依据请求UA头智能使用 PcEmulate H5Emulate
	_screenshotPath string            // 单次使用的截图存储路径
	cookies         []chromedp.Action // 设置随请求发送的cookie
}

// New 初始化
func New() *Chrome {
	return &Chrome{
		Ctx:             nil,
		cancelCtxFunc:   nil,
		cancel:          nil,
		headless:        false,
		emulate:         nil,
		_screenshotPath: "",
		cookies:         nil,
	}
}

// SetShowBrowser 设置是否显示浏览器界面UI
//
//	-- show 是否显示浏览器UI
func (c *Chrome) SetShowBrowser(show bool) *Chrome {
	c.headless = !show
	return c
}

// SetProxy 设置代理服务器
//
//	-- proxyUrl 代理服务器url
//
// 参数错误不会报告异常，只有代理地址符合url规则才会被设置
//
// c = chrome.New().SetProxy("xxx")
func (c *Chrome) SetProxy(proxyUrl string) *Chrome {
	if proxyUrl != "" {
		if _, err := url.Parse(proxyUrl); err == nil {
			c.proxyUrl = proxyUrl
		}
	}

	return c
}

// SetDebugLogger 设置调试logger
func (c *Chrome) SetDebugLogger(f func(string, ...interface{})) *Chrome {
	c.debugLogger = f
	return c
}

// SetErrorLogger 设置错误logger
func (c *Chrome) SetErrorLogger(f func(string, ...interface{})) *Chrome {
	c.errorLogger = f
	return c
}

// SetInfoLogger 设置记录logger
func (c *Chrome) SetInfoLogger(f func(string, ...interface{})) *Chrome {
	c.infoLogger = f
	return c
}

// SetCookie 设置随请求发送出去的cookie
//   - key      cookie键名
//   - value    cookie键值
//   - domain   cookie的有效域名，例如：.a.com 或 www.a.com
//   - path     path路径，例如：/
//   - expired  cookie有效时长
//   - httpOnly 是否httpOnly
//   - secure   是否secure安全
func (c *Chrome) SetCookie(key, value, domain, path string, expired time.Duration, httpOnly, secure bool) *Chrome {
	cookie := chromedp.ActionFunc(func(ctx context.Context) error {
		expr := cdp.TimeSinceEpoch(time.Now().Add(expired))
		err := network.SetCookie(key, value).
			WithExpires(&expr).
			WithDomain(domain).
			WithPath(path).
			WithHTTPOnly(httpOnly).
			WithSecure(secure).
			Do(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	// add cookie
	if nil == c.cookies {
		c.cookies = make([]chromedp.Action, 0)
	}
	c.cookies = append(c.cookies, cookie)

	return c
}

// CleanCookie 清理启动时设置的cookie
func (c *Chrome) CleanCookie() *Chrome {
	c.cookies = nil
	return c
}

// SetCaptureScreenshot 设置截图存储文件夹路径
//
//   - savePath 存储截图的文件夹路径，路径需是存在可读写的， 例如 ./runtime/
//
// 使用示例：
//
// c = crawler.New()
//
// c.CaptureScreenshot("./runtime/").Crawler("https://www.baidu.com/", "", tasks)
func (c *Chrome) SetCaptureScreenshot(savePath string) *Chrome {
	if PathExists(savePath) {
		c._screenshotPath = savePath
		if !strings.HasSuffix(savePath, "/") {
			c._screenshotPath = savePath + "/" // 确保一定有后缀斜杠
		}
	}

	return c
}

// SetEmulate 设置爬虫抓取仿真参数
//   - emulate nil恢复默认行为即依据请求UA头自动选择使用 PcEmulate 或 H5Emulate，非nil则自定义
func (c *Chrome) SetEmulate(emulate *device.Info) *Chrome {
	c.emulate = emulate

	return c
}

// SetFromUA 设置请求来源UA头，如果要设置打开的chrome的自定义UA头请使用 SetEmulate 方法
//   - userAgent 设置请求来源UA头
//
// 本方法用于设置请求来源UA头设置
func (c *Chrome) SetFromUA(userAgentFrom string) *Chrome {
	c.ua = userAgentFrom

	return c
}

// OpenChrome 打开chrome并初始化各参数
func (c *Chrome) OpenChrome() {
	c.openChrome()
}

// InitContext 手动初始化控制器context，方便需要Execute之前获取ctx场景
func (c *Chrome) InitContext(timeout ...time.Duration) {
	c.initContext(timeout...)
}

// CloseChrome context形式关闭chrome
func (c *Chrome) CloseChrome() {
	c.closeChrome()
}

// Execute 执行任务方法抽象
//   - target   打开的网页URL
//   - node     打开网页指标元素ID，以等待网页已经渲染完毕
//   - task     补充的操作
//   - timeout  可选的超时时长，不传或传0表示发出一个请求不超时
//
// 返回值说明
//   - string            如果有html内容，则html字符串
//   - *network.Response chromedp原生resp
//   - error             是否报错
func (c *Chrome) Execute(target, node string, task chromedp.Tasks, timeout ...time.Duration) (string, *network.Response, error) {
	var (
		html    string
		png     []byte
		emulate device.Info
	)

	// 处理仿真参数：未自自定义则依据ua头侦测
	if c.emulate == nil {
		emulate = PcEmulate
		if mobileUA.MatchString(c.ua) {
			emulate = H5Emulate
		}
	}

	// 通用任务设置
	tasks := chromedp.Tasks{
		disableAutomationDetect(),
		chromedp.Emulate(emulate), // UA头
		// browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorDeny), // 禁用自动文件下载
		chromedp.Navigate(target), // chrome打开目标网页
	}
	if node != "" {
		tasks = append(tasks, WaitRootNodeVisible(node))
	}
	tasks = append(
		tasks,
		task, // 补充action任务
		chromedp.OuterHTML("html", &html, chromedp.ByQuery), // 最后获取网页HTMl内容
	)

	// 如果要求添加截图
	if c._screenshotPath != "" {
		tasks = append(tasks, chromedp.CaptureScreenshot(&png))
	}

	// 如果有cookie，写cookie应当在打开url之前
	if nil != c.cookies {
		tasks = append(c.cookies, tasks...)
	}

	// +++++++++++++++++++++++
	// 初始化chrome
	// +++++++++++++++++++++++
	if c.instanceContext == nil {
		c.openChrome()
	}

	// +++++++++++++++++++++++
	// 构造context
	// +++++++++++++++++++++++
	if nil == c.Ctx {
		c.initContext(timeout...)
	}

	// +++++++++++++++++++++++
	// 执行带响应返回
	// +++++++++++++++++++++++
	resp, err := chromedp.RunResponse(c.Ctx, tasks...)

	// 最后取消请求体context
	// defer c.cancelRequest()

	// 如果要求截图则保存截图
	if c._screenshotPath != "" && nil != png {
		_ = os.WriteFile(c._screenshotPath+hashSha1([]byte(target))+".png", png, 0644)
	}

	return html, resp, err
}

// initContext 初始化控制器context，方便需要再Execute之前获取ctx场景
func (c *Chrome) initContext(timeout ...time.Duration) {
	if c.instanceContext == nil {
		c.openChrome()
	}

	if c.Ctx != nil {
		return
	}

	if len(timeout) <= 0 || 0 == timeout[0] {
		// 构造无超时时长的context
		c.makeCancelContext()
	} else {
		// 构造指定超时时长的context
		c.makeTimeoutContext(timeout[0])
	}
}

// openChrome 初始化启动chrome
func (c *Chrome) openChrome() *Chrome {
	if c.instanceContext != nil {
		return c
	}

	// 参数配置
	options := []chromedp.ExecAllocatorOption{
		// 是否headless模式--true则没有浏览器界面 false则有一个浏览器界面
		// 打开浏览器界面便于调试，非开发环境则是headless模式
		chromedp.Flag("headless", c.headless),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("remote-debugging-port", "9222"),
		chromedp.Flag("window-size", "1920,1080"),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	// 如果有配置代理，则设置代理参数
	if c.proxyUrl != "" {
		options = append(options, chromedp.ProxyServer(c.proxyUrl))
	}

	// 启动chrome实例
	c.instanceContext, c.instanceCloseFunc = chromedp.NewExecAllocator(context.Background(), options...)

	return c
}

// closeChrome 关闭chrome
func (c *Chrome) closeChrome() {
	if nil != c.instanceContext {
		if nil != c.Ctx {
			c.cancelRequest() // 如果请求还未结束则先取消|结束请求
		}
		c.instanceCloseFunc()
		c.instanceContext = nil // reset
	}
}

// cancelRequest 取消当前上下文的锚定的请求
func (c *Chrome) cancelRequest() {
	// chrome实例ctx存在且请求实例ctx存在
	if nil != c.instanceContext && nil != c.Ctx {
		c.cancelCtxFunc()
		c.cancel()
		c.Ctx = nil // reset
	}
}

// makeCancelContext 构造手动关闭的ctx，即爬取网页没有超时
func (c *Chrome) makeCancelContext() {
	var instanceCtx context.Context
	instanceCtx, c.cancel = context.WithCancel(c.instanceContext)

	// logger自定义日志
	perhapsLogger := make([]chromedp.ContextOption, 0)
	if c.debugLogger != nil {
		perhapsLogger = append(perhapsLogger, chromedp.WithDebugf(c.debugLogger))
	}
	if c.errorLogger != nil {
		perhapsLogger = append(perhapsLogger, chromedp.WithErrorf(c.errorLogger))
	}
	if c.infoLogger != nil {
		perhapsLogger = append(perhapsLogger, chromedp.WithLogf(c.infoLogger))
	}

	if len(perhapsLogger) > 0 {
		c.Ctx, c.cancelCtxFunc = chromedp.NewContext(instanceCtx, perhapsLogger...)
	} else {
		c.Ctx, c.cancelCtxFunc = chromedp.NewContext(instanceCtx)
	}
}

// makeTimeoutContext 构造指定超时时长的ctx，即爬取网页这个时间后会超时
func (c *Chrome) makeTimeoutContext(timeout time.Duration) {
	var instanceCtx context.Context
	instanceCtx, c.cancel = context.WithTimeout(c.instanceContext, timeout)

	// logger自定义日志
	perhapsLogger := make([]chromedp.ContextOption, 0)
	if c.debugLogger != nil {
		perhapsLogger = append(perhapsLogger, chromedp.WithDebugf(c.debugLogger))
	}
	if c.errorLogger != nil {
		perhapsLogger = append(perhapsLogger, chromedp.WithErrorf(c.errorLogger))
	}
	if c.infoLogger != nil {
		perhapsLogger = append(perhapsLogger, chromedp.WithLogf(c.infoLogger))
	}

	if len(perhapsLogger) > 0 {
		c.Ctx, c.cancelCtxFunc = chromedp.NewContext(instanceCtx, perhapsLogger...)
	} else {
		c.Ctx, c.cancelCtxFunc = chromedp.NewContext(instanceCtx)
	}
}
