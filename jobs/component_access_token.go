package jobs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zjxpcyc/wechat-scheduler/lib"
)

// ComponentAccessToken 开放平台 access_token
// ref: https://open.weixin.qq.com/cgi-bin/showdocument?action=dir_list&t=resource/res_list&verify=1&id=open1453779503&token=&lang=zh_CN
func ComponentAccessToken(tk *JobTask) *lib.JobServer {
	if tk.Execable != nil {
		return tk.Execable
	}

	taskName := "component_access_token"
	t := tk.Job

	task := func() error {
		postData := map[string]string{
			"component_appid":     t.AppID,
			"component_appsecret": t.AppSecret,
		}

		if tk.DynamicParams == nil {
			return errors.New("启动 " + taskName + " 任务失败, 注册须指定 DynamicParams 参数")
		}

		params := tk.DynamicParams(t.AppID, JOB_COMPONENT_ACCESS_TOKEN)
		if ticket, ok := params["component_verify_ticket"]; ok {
			postData["component_verify_ticket"] = ticket.(string)
		} else {
			return errors.New("获取 " + taskName + " 失败: 未找到有效 verify_ticket")
		}

		dt, err := json.Marshal(params)
		if err != nil {
			return err
		}

		res := map[string]interface{}{}
		_, err = lib.Request(tk.API, nil, bytes.NewBuffer(dt), &res)
		if err != nil {
			return err
		}

		if code, ok := res["errcode"]; ok {
			return fmt.Errorf("%v - %s", code, res["errmsg"])
		}

		tk.Result = res
		tk.Value = res["ComponentVerifyTicket"].(string)
		tk.LastTime = time.Now().Local()
		tk.Save()

		if tk.CallBack != nil {
			tk.CallBack(t.AppID, JOB_COMPONENT_ACCESS_TOKEN, res)
		}

		return nil
	}

	return lib.NewJobServer(t.AppID, taskName, task, tk.Freq)
}
