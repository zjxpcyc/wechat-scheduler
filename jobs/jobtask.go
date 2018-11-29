package jobs

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/zjxpcyc/wechat-scheduler/lib"
)

// 任务类型
// 使用递增数字, 主要是为了解决部分任务是有依赖关系的
// 如果启动的任务列表有这种依赖关系的任务, 那么初始启动任务的时候会按照这个顺序启动
const (
	// access_token
	JOB_ACCESS_TOKEN = iota

	// web access_token
	JOB_WEB_ACCESS_TOKEN

	// jsapi_ticket
	JOB_JSAPI_TICKET

	// component_access_token
	JOB_COMPONENT_ACCESS_TOKEN

	// authorizer_access_token
	JOB_AUTHORIZER_ACCESS_TOKEN

	// 不支持的类型, 同时也作为边界判定值
	JOB_MAX_LIMIT
)

// FREQUENCY 目前已知的微信定时任务都是 7200 秒, 这里提前 200s 启动
const FREQUENCY = 7000

// JobTask 单个任务
// 比如 access_token 获取等
type JobTask struct {
	// Typ 任务类型
	Typ int

	// DynamicParams 需要动态传入过来的参数
	// 例如, 微信开放平台的 component_token 需要 verify_ticket 来获取.
	// 但是 verify_ticket 的刷新频率是 10 分钟一次
	DynamicParams func(appid string, typ int) map[string]interface{}

	// Result 当前任务结果
	// 与 Value 有区别
	Result map[string]interface{}

	// Value 当前任务的目的结果
	// 比如, access_token 任务的 Result 是 { "access_token":"ACCESS_TOKEN", "expires_in":7200 }
	// 但是 Value 则直接就是 access_token 的值
	Value string

	// LastTime 上次执行时间
	LastTime time.Time

	// API 接口
	API lib.WechatAPI

	// Freq 刷新频率
	Freq time.Duration

	// CallBack 成功之后的回调
	CallBack func(appid string, typ int, result map[string]interface{})

	// Execable 任务Server
	Execable *lib.JobServer

	// Job 所属 Job
	Job *Job
}

// JobAPIs 目前支持的微信 api
var JobAPIs = map[int]lib.WechatAPI{
	JOB_ACCESS_TOKEN: lib.WechatAPI{
		Name:        "access_token",
		URL:         "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=APPID&secret=APPSECRET",
		Method:      http.MethodGet,
		ContentType: lib.MimeJSON,
	},

	JOB_WEB_ACCESS_TOKEN: lib.WechatAPI{
		Name:        "web_access_token",
		URL:         "https://api.weixin.qq.com/sns/oauth2/refresh_token?appid=APPID&grant_type=refresh_token&refresh_token=REFRESH_TOKEN",
		Method:      http.MethodGet,
		ContentType: lib.MimeJSON,
	},

	JOB_JSAPI_TICKET: lib.WechatAPI{
		Name:        "jsapi_ticket",
		URL:         "https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=ACCESS_TOKEN&type=jsapi",
		Method:      http.MethodGet,
		ContentType: lib.MimeJSON,
	},

	JOB_COMPONENT_ACCESS_TOKEN: lib.WechatAPI{
		Name:        "component_token",
		URL:         "https://api.weixin.qq.com/cgi-bin/component/api_component_token",
		Method:      http.MethodPost,
		ContentType: lib.MimeJSON,
	},

	JOB_AUTHORIZER_ACCESS_TOKEN: lib.WechatAPI{
		Name:        "authorizer_access_token",
		URL:         "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token=xxxxx",
		Method:      http.MethodPost,
		ContentType: lib.MimeJSON,
	},
}

// Start task
func (t *JobTask) Start() {
	if t.Execable == nil {
		return
	}

	if t.Execable.Status == lib.TASK_STARTED {
		return
	}

	diff := time.Now().Local().Sub(t.LastTime)
	delay := diff - FREQUENCY

	if delay > 0 && delay < FREQUENCY*time.Second {
		t.Execable.Start(delay)
	} else {
		t.Execable.Start()
	}

	logger.Info("任务已加入启动队列...")
}

// Stop task
func (t *JobTask) Stop() {
	if t.Execable == nil {
		return
	}

	if t.Execable.Status == lib.TASK_STARTED {
		t.Execable.Stop()
	}
}

// Save into the job model
func (t *JobTask) Save() {
	prefix := "task-" + strconv.Itoa(t.Typ)

	lasttime := t.LastTime.Format("2006-01-02 15:04:05")
	t.Job.Model.Update(prefix+"-lasttime", lasttime)

	result, err := json.Marshal(t.Result)
	if err != nil {
		logger.Error("保存任务结果失败: ", err.Error())
	} else {
		t.Job.Model.Update(prefix+"-result", string(result))
	}

	t.Job.Model.Update(prefix+"-value", t.Value)
}

// Set property of task
func (t *JobTask) Set(k, v string) {
	switch k {
	case "lasttime":
		tm, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local)
		if err != nil {
			logger.Error("设置 task Lasttime 属性出错:", err.Error())
			break
		}

		t.LastTime = tm
	case "result":
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(v), &result); err != nil {
			logger.Error("设置 task Result 属性出错:", err.Error())
			break
		}

		t.Result = result
	case "value":
		t.Value = v
	default:
	}
}
