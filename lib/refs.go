package lib

import "net/url"

// WechatAPI 微信接口
type WechatAPI struct {
	// 接口名称
	Name string

	// 接口地址
	URL string

	// 请求方式
	Method string

	// 请求及返回格式
	ContentType string
}

// SetQueryParams 设置 URL query 参数
func (w *WechatAPI) SetQueryParams(params url.Values) *WechatAPI {
	u, e := url.Parse(w.URL)
	if e != nil {
		return w
	}

	q := u.Query()
	for k := range q {
		if v, ok := params[k]; ok {
			q[k] = v
		}
	}

	u.RawQuery = q.Encode()
	w.URL = u.String()
	return w
}

var (
	// MimeJSON mime-type json
	MimeJSON = "application/json"

	// MimeXML mime-type xml
	MimeXML = "text/xml"
)
