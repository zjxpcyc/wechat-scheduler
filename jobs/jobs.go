package jobs

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/zjxpcyc/tinylogger"
	"github.com/zjxpcyc/wechat-scheduler/database"
	"github.com/zjxpcyc/wechat-scheduler/lib"
)

// Job 任务集合
// 以 appid 为分组key, 将单个账号下的所有任务集合到一起
type Job struct {
	// AppID 微信应用ID
	AppID string

	// AppSecret 微信应用 Secret
	AppSecret string

	// Tasks 当前 Job 所有的注册 task
	Tasks map[int]*JobTask

	// Model
	Model *database.Model
}

// AllJob 所有注册的 Job
var AllJob = make(map[string]*Job)

var logger tinylogger.LogService

// NewJob 新建一个 Job
// 如果同一个 appid 多次创建, 那么返回的是同一个 Job
// 因此 可以支持 runtime 更新 Job
func NewJob(appid, appsecret string) (*Job, error) {
	if appid == "" || appsecret == "" {
		return nil, errors.New("新建 Job 失败, appid 或者 appsecret 不能为空")
	}

	m, ok := database.AllModel[appid]
	if !ok {
		var err error
		m, err = database.NewModel(appid)
		if err != nil {
			return nil, err
		}
	}

	// 重复注册, 以最后一次为准
	// 因为有可能出现 appsecret 变更的情况
	m.Update("appid", appid)
	m.Update("appsecret", appsecret)

	if j, ok := AllJob[appid]; ok {
		j.AppID = appid
		j.AppSecret = appsecret

		return j, nil
	}

	j := &Job{
		AppID:     appid,
		AppSecret: appsecret,
		Tasks:     make(map[int]*JobTask),
		Model:     m,
	}

	AllJob[appid] = j
	return j, nil
}

// NewTask 新建一个 Task
// 支持任务的重复创建
func (t *Job) NewTask(typ int, dynAddr, cbAddr string) *JobTask {
	// 不支持的类型
	if typ < 0 || typ >= JOB_MAX_LIMIT {
		logger.Error("创建 Task 失败, 不支持的 Task 类型: " + strconv.Itoa(typ))
		return nil
	}

	t.Model.Update("dyn-"+strconv.Itoa(typ), dynAddr)
	t.Model.Update("cb-"+strconv.Itoa(typ), cbAddr)

	if tk, ok := t.Tasks[typ]; ok {
		tk.DynamicParams = lib.DynamicFuncFactory(dynAddr)
		tk.CallBack = lib.CallBackFuncFactory(cbAddr)

		return tk
	}

	freq := FREQUENCY * time.Second

	task := &JobTask{
		Typ:           typ,
		DynamicParams: lib.DynamicFuncFactory(dynAddr),
		API:           JobAPIs[typ],
		Freq:          freq,
		CallBack:      lib.CallBackFuncFactory(cbAddr),
		Job:           t,
	}

	switch typ {
	case JOB_ACCESS_TOKEN:
		task.Execable = AccessToken(task)
	case JOB_WEB_ACCESS_TOKEN:
		task.Execable = WebAccessToken(task)
	case JOB_JSAPI_TICKET:
		task.Execable = JsApiTicket(task)
	case JOB_COMPONENT_ACCESS_TOKEN:
		task.Execable = ComponentAccessToken(task)
	case JOB_AUTHORIZER_ACCESS_TOKEN:
		task.Execable = AuthorizerAccessToken(task)
	}

	// 更新 model
	tasklist, _ := t.Model.Query("tasklist")
	if tasklist == "" {
		tasklist = strconv.Itoa(typ)
	} else {
		tasklist = lib.DistinctStr(tasklist + "," + strconv.Itoa(typ))
	}

	t.Model.Update("tasklist", tasklist)

	t.Tasks[typ] = task
	return task
}

// Run 自动运行任务, 自动运行不支持任务停止
// 如果需要手动运行, 请直接调用相关任务的方法
func (t *Job) Run() {
	for i := 0; i < JOB_MAX_LIMIT; i++ {
		tk, ok := t.Tasks[i]
		if !ok {
			continue
		}

		tk.Start()
	}
}

// Init 从数据库文件进行系统初始化
func Init() {
	logger = lib.GetLogger()
	logger.Info("开始进行任务列表初始化 ...")

	if database.AllModel == nil || len(database.AllModel) == 0 {
		return
	}

	for appid, m := range database.AllModel {
		appsecret, err := m.Query("appsecret")
		if err != nil {
			logger.Error("初始化 Job-"+appid+" 失败: ", err.Error())
			continue
		}

		job, err := NewJob(appid, appsecret)
		if err != nil {
			logger.Error("初始化 Job-"+appid+" 失败: ", err.Error())
			continue
		}

		tasklist, err := m.Query("tasklist")
		if err != nil {
			logger.Error("初始化 Job-"+appid+" tasklist 失败: ", err.Error())
			continue
		}

		if tasklist == "" {
			continue
		}

		allTasks := strings.Split(tasklist, ",")
		for _, typStr := range allTasks {
			dyn, err := m.Query("dyn-" + typStr)
			if err != nil {
				logger.Error("初始化 Job-"+appid+" task["+typStr+"] 失败: ", err.Error())
				continue
			}

			cb, err := m.Query("cb-" + typStr)
			if err != nil {
				logger.Error("初始化 Job-"+appid+" task["+typStr+"] 失败: ", err.Error())
				continue
			}

			typ, _ := strconv.Atoi(typStr)
			tk := job.NewTask(typ, dyn, cb)

			prefix := "task-" + strconv.Itoa(typ)

			result, err := m.Query(prefix + "-result")
			if err != nil {
				logger.Error("初始化 Job-"+appid+" task["+typStr+"] Result 失败: ", err.Error())
				continue
			}

			lasttime, err := m.Query(prefix + "-lasttime")
			if err != nil {
				logger.Error("初始化 Job-"+appid+" task["+typStr+"] Lasttime 失败: ", err.Error())
				continue
			}

			value, err := m.Query(prefix + "-value")
			if err != nil {
				logger.Error("初始化 Job-"+appid+" task["+typStr+"] Value 失败: ", err.Error())
				continue
			}

			tk.Set("result", result)
			tk.Set("lasttime", lasttime)
			tk.Set("value", value)
		}

		job.Run()
		AllJob[appid] = job
	}
}
