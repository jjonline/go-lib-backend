package crawler

import (
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"time"
)

// WaitRootNodeVisible 等待根节点可见
//   - rootNodeID html根节点ID
func WaitRootNodeVisible(rootNodeID string) chromedp.Action {
	return chromedp.WaitVisible(rootNodeID, chromedp.ByID)
}

// SetCookie 设置cookie
//   - key     cookie键名
//   - value   cookie键值
//   - domain  cookie的有效域名，例如：a.com 或 www.a.com
//   - expired cookie有效时长
func SetCookie(key, value, domain string, expired time.Duration) chromedp.Action {
	expr := cdp.TimeSinceEpoch(time.Now().Add(expired))
	return network.SetCookie(key, value).WithExpires(&expr).WithDomain(domain)
}

// ScrollToBottom 下拉滚动条到最底部
//
// 通过JS下拉网页滚动条到最底部以触发一些滚动加载的逻辑
func ScrollToBottom() chromedp.Action {
	return chromedp.EvaluateAsDevTools("window.scrollTo(0,document.body.scrollHeight);", nil)
}

// IsSoft404NotFound 读取html的meta头判断当前页面是否为一个软404页面
//
//		前端需遵循软404页面通过js添加<meta name="robots" content="noindex">
//
//		- result 执行JS获取meta后返回的结果对象引用
//
//	 判断是否有有添加软404的meta标记 result != nil && string(is404.Value) == "true"
func IsSoft404NotFound(result **runtime.RemoteObject) chromedp.Action {
	var js = `
(function() {
	_meta = document.querySelector('meta[name=robots]');
	if (_meta) {
		return _meta.getAttribute('content') === 'noindex';
	}
	return false;
})();
`
	return chromedp.EvaluateAsDevTools(js, result)
}

// WaitDefineMetaReady 等待自定义meta标签出现
//   - metaName meta标签name
//
// 对于SPA页面，可以约定SPA渲染完毕后往head标签里添加一个meta标签 例如：<meta name="metaName" content="true">
// 然后crawler监听是否出现这个标签以侦测渲染完毕
// 每50毫秒侦测1次直到10秒后超时
func WaitDefineMetaReady(metaName string) chromedp.Action {
	return chromedp.Poll(
		"document.querySelector('meta[name="+metaName+"]') != null",
		nil,
		chromedp.WithPollingInterval(50*time.Millisecond),
		chromedp.WithPollingTimeout(10*time.Second),
	)
}

// GetLastURL 页面可能会出现Redirect，用于获取最后停留页面URL
//   - lastUrl 最后停留页面的URL
func GetLastURL(lastUrl *string) chromedp.Action {
	return chromedp.Location(lastUrl)
}
