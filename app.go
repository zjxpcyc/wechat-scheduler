package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/zjxpcyc/wechat-scheduler/jobs"
)

// App is a http.Handler
type App struct{}

// ServeHTTP 实现接口
func (t *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if p := recover(); p != nil {
			// do nothing
		}
	}()

	r.ParseForm()

	ctrl := new(Controller)

	ctrl.Get = r.FormValue
	ctrl.Input = r
	ctrl.Output = w

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("获取请求内容失败: " + err.Error())
	}

	logger.Info("获取到 body 参数: " + string(body))
	ctrl.Body = body

	if strings.Index(r.URL.Path, "/registe") > -1 {
		ctrl.RegisteTasks()
	}

	if strings.Index(r.URL.Path, "/task") > -1 {
		ctrl.GetTaskValue()
	}
}

type Controller struct {
	Get  func(string) string
	Body []byte

	Input  *http.Request
	Output http.ResponseWriter
}

type RegisteTask struct {
	Typ    int    `json:"type"`
	Notify string `json:"notify"`
	Params string `json:"params"`
}

type RegisteParam struct {
	AppID     string        `json:"appid"`
	AppSecret string        `json:"appsecret"`
	Tasks     []RegisteTask `json:"tasks"`
}

// RegisteTasks 注册
// 注册传入的参数为 json 格式, 通过 http body 传入
// {
// 	"appid": "",
// 	"appsecret": "",
// 	"tasks": [
// 		{
// 			"type": 0,
// 			"notify": "",
// 			"params": "",
// 		},
// 		...
// 	]
// }
func (t *Controller) RegisteTasks() {
	// 解析传入参数
	if t.Body == nil || len(t.Body) == 0 {
		t.ResponseJSON(errors.New("注册失败: 注册参数不能为空"), http.StatusBadRequest)
	}

	params := RegisteParam{}
	if err := json.Unmarshal(t.Body, &params); err != nil {
		logger.Error("读取注册参数失败: " + err.Error())
		t.ResponseJSON(errors.New("注册失败: 读取参数失败"), http.StatusBadRequest)
	}

	if params.AppID == "" || params.AppSecret == "" {
		t.ResponseJSON(errors.New("注册失败: appid 或者 appsecret 不能为空"), http.StatusBadRequest)
	}

	tasks := params.Tasks
	if tasks == nil || len(tasks) == 0 {
		t.ResponseJSON("")
	}

	// 注册 Job
	job, err := jobs.NewJob(params.AppID, params.AppSecret)
	if err != nil {
		logger.Error("注册任务失败: (appid: " + params.AppID + ", appsecret: " + params.AppSecret + ") : " + err.Error())
		t.ResponseJSON(errors.New("注册任务失败, 请重试"), http.StatusBadRequest)
	}

	// 添加任务
	for _, tk := range tasks {
		if tk.Typ < 0 || tk.Typ >= jobs.JOB_MAX_LIMIT {
			t.ResponseJSON(errors.New("注册失败: 不支持的任务类型"), http.StatusBadRequest)
		}

		job.NewTask(tk.Typ, tk.Params, tk.Notify)
	}

	// 运行任务
	job.Run()
}

// GetTaskValue 获取当前任务的值
func (t *Controller) GetTaskValue() {
	ps := strings.Split(t.Input.URL.Path, "/")

	if len(ps) < 3 {
		t.ResponseJSON(errors.New("请求地址不存在"), http.StatusNotFound)
	}

	psLen := len(ps)
	typ, err := strconv.Atoi(ps[psLen-1])
	if err != nil {
		logger.Error("获取任务值出错: ", err.Error())
		t.ResponseJSON(errors.New("请求地址格式不正确"))
	}

	if typ < 0 || typ >= jobs.JOB_MAX_LIMIT {
		t.ResponseJSON(errors.New("非法的任务类型"))
	}

	logger.Info(jobs.AllJob)

	appid := ps[psLen-2]
	if job, ok := jobs.AllJob[appid]; !ok {
		t.ResponseJSON(errors.New("非法的 AppID"))
	} else {
		tk, has := job.Tasks[typ]
		if !has {
			t.ResponseJSON(errors.New("当前 AppID 并未注册指定类型的 任务"))
		}

		t.ResponseJSON(tk.Value)
	}
}

// ResponseJSON 统一约定返回 json
func (t *Controller) ResponseJSON(data interface{}, code ...int) {
	status := http.StatusOK
	if code != nil && len(code) > 0 {
		status = code[0]
	}

	var message string
	var result interface{}

	switch d := data.(type) {
	case error:
		message = d.Error()
		if status == http.StatusOK {
			status = http.StatusBadRequest
		}
	default:
		result = d
	}

	mapData := map[string]interface{}{
		"code":    status,
		"message": message,
		"result":  result,
	}

	rtn, err := json.Marshal(mapData)
	if err != nil {
		logger.Error("转换待返回数据失败: " + err.Error())
		// t.Output.Write([]byte("内部错误"))
		// t.Output.WriteHeader(http.StatusInternalServerError)
		http.Error(t.Output, "转换待返回数据失败", http.StatusInternalServerError)
	} else {
		t.Output.Header().Set("Content-Type", "application/json")
		t.Output.WriteHeader(http.StatusOK)
		t.Output.Write(rtn)
	}

	panic("")
}
