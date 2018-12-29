package jobs

import (
	"net/url"
	"time"

	"github.com/zjxpcyc/wechat-scheduler/lib"
)

// AccessToken 获取 access_token
// ref: https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140183
func AccessToken(tk *JobTask) *lib.JobServer {
	if tk.Execable != nil {
		return tk.Execable
	}

	taskName := "access_token"
	t := tk.Job

	task := func() error {
		query := url.Values{}
		query.Add("appid", t.AppID)
		query.Add("secret", t.AppSecret)

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
			tk.CallBack(t.AppID, JOB_ACCESS_TOKEN, res)
		}

		return nil
	}

	return lib.NewJobServer(t.AppID, taskName, task, tk.Freq)
}
