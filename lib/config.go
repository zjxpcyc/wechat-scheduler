package lib

import (
	"io/ioutil"

	"github.com/tidwall/gjson"
)

// GetConfig 获取配置信息
// 通过 gjson 库读取配置文件
// https://github.com/tidwall/gjson
func GetConfig(fpath string) gjson.Result {
	raw, err := ioutil.ReadFile(fpath)
	if err != nil {
		logger.Error("配置文件读取失败", err.Error())
		return gjson.Parse("{}")
	}

	return gjson.ParseBytes(raw)
}
