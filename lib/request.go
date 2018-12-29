package lib

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Request 请求数据
// 请求远程 http 服务数据, 针对微信相关接口进行了特殊的处理
// 请求的地址, 方式等通过 api 指定, url 中的 search 参数通过 query 指定
// 如果请求中需要包含 body 内容, 则传入 body 参数, 没有传 nil 即可
// result 是可选参数, 用来承载或者格式化远程请求的结果, 比如 远程返回的实际上是 json 字串, 那么 result 可以为该 json 对应的 struct 指针
// data 返回值是远程请求的原始结果内容
func Request(api WechatAPI, query url.Values, body io.Reader, result ...interface{}) (data []byte, err error) {
	// 请求地址
	if query != nil {
		(&api).SetQueryParams(query)
	}
	addr := api.URL

	logger.Info("远程请求接口 URL ", addr)
	logger.Info("远程请求方法 ", api.Method)

	// 请求 Body
	var bodyData io.Reader
	if api.Method != http.MethodGet && body != nil {
		bodyData = body

		b := &bytes.Buffer{}
		io.Copy(b, body)
		logger.Info("远程请求体内容 ", b.String())
	} else {
		bodyData = nil
	}

	// 构造 http 请求
	var req *http.Request
	var res *http.Response
	client := new(http.Client)

	req, err = http.NewRequest(api.Method, addr, bodyData)
	if err != nil {
		logger.Error("初始化 http 客户端失败 ", err.Error())
		return
	}

	req.Header.Add("Content-type", api.ContentType)

	res, err = client.Do(req)
	if err != nil {
		logger.Error("http 请求数据失败 ", err.Error())
		return
	}

	data, err = ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		logger.Error("读取 http 请求结果失败 ", err.Error())
		return
	} else {
		logger.Info("远程请求结果 ", string(data))
	}

	// 格式化结果
	if result != nil && len(result) > 0 {
		if api.ContentType == MimeJSON {
			err = json.Unmarshal(data, result[0])
			if err != nil {
				logger.Error("格式化 http 请求结果失败 ", err.Error())
				return
			}
		} else if api.ContentType == MimeXML {
			err = xml.Unmarshal(data, result[0])
			if err != nil {
				logger.Error("格式化 http 请求结果失败 ", err.Error())
				return
			}
		}
	}

	return
}

// CheckJSONResult 校验结果
func CheckJSONResult(res map[string]interface{}) error {
	code, ok := res["errcode"]
	if !ok {
		return nil
	}

	switch status := code.(type) {
	case float64:
		if int(status) == 0 {
			return nil
		}
	case string:
		if status == "0" || status == "" {
			return nil
		}
	}

	return fmt.Errorf("%v - %s", code, res["errmsg"])
}
