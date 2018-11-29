package lib

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

// DynamicFuncFactory 获取动态参数函数
func DynamicFuncFactory(addr string) func(string, int) map[string]interface{} {
	if addr == "" {
		return nil
	}

	f := func(appid string, typ int) map[string]interface{} {
		res := make(map[string]interface{})

		api := WechatAPI{
			Name:        "dynamic function",
			URL:         addr,
			Method:      http.MethodGet,
			ContentType: MimeJSON,
		}

		query := url.Values{}
		query.Add("appid", appid)
		query.Add("type", strconv.Itoa(typ))

		Request(api, query, nil, &res)
		return res
	}

	return f
}

// CallBackFuncFactory 获取回调函数
func CallBackFuncFactory(addr string) func(string, int, map[string]interface{}) {
	if addr == "" {
		return nil
	}

	f := func(appid string, typ int, result map[string]interface{}) {
		api := WechatAPI{
			Name:        "callback function",
			URL:         addr,
			Method:      http.MethodPost,
			ContentType: MimeJSON,
		}

		query := url.Values{}
		query.Add("appid", appid)
		query.Add("type", strconv.Itoa(typ))

		dt, _ := json.Marshal(result)

		Request(api, query, bytes.NewBuffer(dt))
		return
	}

	return f
}
