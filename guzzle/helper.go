package guzzle

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// toQueryUrl url和查询字符串拼接成完整URL
func toQueryUrl(url string, query interface{}) string {
	if query != nil {
		// url里不存在 ? 符号 直接拼接返回
		if !strings.Contains(url, "?") {
			return url + "?" + BuildQuery(query)
		}

		// url里存在 ? 符号，去除右侧 & 符号后，再使用&符拼接返回
		return strings.TrimRight(url, "&") + "&" + BuildQuery(query)
	}
	return url
}

// toJsonReader 处理参数为JSON类型
func toJsonReader(param interface{}) io.Reader {
	switch pv := param.(type) {
	case nil:
		return nil
	case io.Reader:
		return pv
	case string:
		return strings.NewReader(pv)
	case []byte:
		return bytes.NewReader(pv)
	default:
		b, _ := json.Marshal(param)
		return bytes.NewReader(b)
	}
}

// toFormReader 处理参数为Form表单类型
//   - 支持的参数类型如下：
//   - nil
//   - io.Reader
//   - string
//   - []byte
//   - map[string]string
//   - map[string][]string <==> url.Values
func toFormReader(param interface{}) io.Reader {
	switch pv := param.(type) {
	case nil:
		return nil
	case io.Reader:
		return pv
	case string:
		return strings.NewReader(pv)
	case []byte:
		return bytes.NewReader(pv)
	case map[string]string, map[string][]string, url.Values:
		return strings.NewReader(BuildQuery(pv))
	default:
		return http.NoBody
	}
}

// BuildQuery 处理请求参数为URL里的Query键值对
//   - 支持的能构建的参数类型如下：
//   - map[string]string
//   - map[string][]string <==> url.Values
//   - 除了上述不支持的类型，其他类型将会忽略返回空字符串
func BuildQuery(param interface{}) string {
	switch pv := param.(type) {
	case map[string]string:
		values := make(url.Values)
		for k, v := range pv {
			values.Add(k, v)
		}
		return values.Encode()
	case map[string][]string:
		values := url.Values(pv)
		return values.Encode()
	case url.Values:
		return pv.Encode()
	default:
		return ""
	}
}
