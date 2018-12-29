package jobs

import (
	"errors"
	"net/url"
	"time"

	"github.com/zjxpcyc/wechat-scheduler/lib"
)

// JsApiTicket 获取网页 oauth2 授权需要的 access_token
// ref: https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421141115
func JsApiTicket(tk *JobTask) *lib.JobServer {
	if tk.Execable != nil {
		return tk.Execable
	}

	taskName := "jsapi_ticket"
	t := tk.Job

	task := func() error {
		query := url.Values{}

		// access_token 有两种来源
		accessToken := ""

		// 1、现有 APPID 的 access_token 任务
		accessTk, ok := t.Tasks[JOB_ACCESS_TOKEN]
		if !ok {
			if tk.DynamicParams == nil {
				return errors.New("启动 " + taskName + " 任务失败, 注册须指定 DynamicParams 参数")
			}

			// 2、是从业务 APP 获取过来
			params := tk.DynamicParams(t.AppID, JOB_ACCESS_TOKEN)
			token, has := params["access_token"]
			if !has {
				return errors.New("刷新 " + taskName + " 失败: 未找到有效 access_token")
			}

			accessToken = token.(string)
		} else {
			accessToken = accessTk.Value
		}

		if accessToken == "" {
			return errors.New("刷新 " + taskName + " 失败: 未找到有效 access_token")
		}

		query.Add("access_token", accessToken)

		res := map[string]interface{}{}
		_, err := lib.Request(tk.API, query, nil, &res)
		if err != nil {
			return err
		}

		if err := lib.CheckJSONResult(res); err != nil {
			return err
		}

		tk.Result = res
		tk.Value = res["ticket"].(string)
		tk.LastTime = time.Now().Local()
		tk.Save()

		if tk.CallBack != nil {
			tk.CallBack(t.AppID, JOB_JSAPI_TICKET, res)
		}

		return nil
	}

	return lib.NewJobServer(t.AppID, taskName, task, tk.Freq)
}
