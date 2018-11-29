package jobs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/zjxpcyc/wechat-scheduler/lib"
)

// AuthorizerAccessToken 开放平台 授权公众号或者小程序 access_token
// ref: https://open.weixin.qq.com/cgi-bin/showdocument?action=dir_list&t=resource/res_list&verify=1&id=open1453779503&token=&lang=zh_CN
func AuthorizerAccessToken(tk *JobTask) *lib.JobServer {
	if tk.Execable != nil {
		return tk.Execable
	}

	taskName := "authorizer_access_token"
	t := tk.Job

	task := func() error {
		postData := map[string]string{
			"component_appid": t.AppID,
		}

		if tk.DynamicParams == nil {
			return errors.New("启动 " + taskName + " 任务失败, 注册须指定 DynamicParams 参数")
		}

		params := tk.DynamicParams(t.AppID, JOB_AUTHORIZER_ACCESS_TOKEN)
		if authApp, ok := params["authorizer_appid"]; ok {
			postData["authorizer_appid"] = authApp.(string)
		} else {
			return errors.New("获取 " + taskName + " 失败: 未找到有效 authorizer_appid")
		}

		// authorizer_refresh_token 先从业务系统查找
		if refresh, ok := params["authorizer_refresh_token"]; ok {
			postData["authorizer_refresh_token"] = refresh.(string)
		} else {
			// 如果找不到, 再从上次任务中查找
			if tk.Result == nil {
				return errors.New("获取 " + taskName + " 失败: 未找到有效 authorizer_refresh_token")
			} else {
				if refresh, ok = tk.Result["authorizer_refresh_token"]; ok {
					postData["authorizer_refresh_token"] = refresh.(string)
				} else {
					return errors.New("获取 " + taskName + " 失败: 未找到有效 authorizer_refresh_token")
				}
			}
		}

		query := url.Values{}

		// component_access_token 从两个地方来
		if token, ok := params["component_access_token"]; ok {
			// 1、业务系统传过来
			query.Add("component_access_token", token.(string))
		} else {
			// 2、从注册任务中查询
			caTk, ok := t.Tasks[JOB_COMPONENT_ACCESS_TOKEN]
			if !ok || caTk.Value == "" {
				return errors.New("获取 " + taskName + " 失败: 未找到有效 component_access_token")
			}

			query.Add("component_access_token", caTk.Value)
		}

		dt, err := json.Marshal(params)
		if err != nil {
			return err
		}

		res := map[string]interface{}{}
		_, err = lib.Request(tk.API, query, bytes.NewBuffer(dt), &res)
		if err != nil {
			return err
		}

		if code, ok := res["errcode"]; ok {
			return fmt.Errorf("%v - %s", code, res["errmsg"])
		}

		tk.Result = res
		tk.Value = res["authorizer_access_token"].(string)
		tk.LastTime = time.Now().Local()
		tk.Save()

		if tk.CallBack != nil {
			tk.CallBack(t.AppID, JOB_AUTHORIZER_ACCESS_TOKEN, res)
		}

		return nil
	}

	return lib.NewJobServer(t.AppID, taskName, task, tk.Freq)
}
