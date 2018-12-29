package jobs

import (
	"errors"
	"net/url"
	"time"

	"github.com/zjxpcyc/wechat-scheduler/lib"
)

// WebAccessToken 获取网页 oauth2 授权需要的 access_token
// ref: https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140842
func WebAccessToken(tk *JobTask) *lib.JobServer {
	if tk.Execable != nil {
		return tk.Execable
	}

	taskName := "web_access_token"
	t := tk.Job

	task := func() error {
		query := url.Values{}
		query.Add("appid", t.AppID)

		// refresh_token 有两种来源
		refreshToken := ""

		// 1、直接是上次任务结果的返回
		refresh := tk.Result["refresh_token"]
		if refresh == nil || refresh.(string) == "" {
			if tk.DynamicParams == nil {
				return errors.New("启动 " + taskName + " 任务失败, 注册须指定 DynamicParams 参数")
			}

			// 2、是从业务 APP 获取过来
			params := tk.DynamicParams(t.AppID, JOB_WEB_ACCESS_TOKEN)
			var ok bool
			refresh, ok = params["refresh_token"]
			if !ok {
				return errors.New("刷新 " + taskName + " 失败: 未找到有效 refresh_token")
			}
		}

		refreshToken = refresh.(string)
		if refreshToken == "" {
			return errors.New("刷新 " + taskName + " 失败: 未找到有效 refresh_token")
		}

		query.Add("refresh_token", refreshToken)

		res := map[string]interface{}{}
		_, err := lib.Request(tk.API, query, nil, &res)
		if err != nil {
			return err
		}

		if err := lib.CheckJSONResult(res); err != nil {
			return err
		}

		tk.Result = res
		tk.Value = res["access_token"].(string)
		tk.LastTime = time.Now().Local()
		tk.Save()

		if tk.CallBack != nil {
			tk.CallBack(t.AppID, JOB_WEB_ACCESS_TOKEN, res)
		}

		return nil
	}

	return lib.NewJobServer(t.AppID, taskName, task, tk.Freq)
}
