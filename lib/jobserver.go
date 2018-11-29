package lib

import (
	"time"
)

// 任务状态
const (
	// 未开始
	TASK_NOT_START = iota

	// 已开始
	TASK_STARTED
)

// JobServer 定时任务服务
type JobServer struct {
	AppID  string
	Name   string
	Status int
	done   chan bool
	task   func() error
	freq   time.Duration
}

// NewJobServer 实例化 JobServer
// query 主要是 微信 一系列接口在调用的时候, url query 参数
// body 是部分接口需要发送的数据
func NewJobServer(appid string, name string, task func() error, freq time.Duration) *JobServer {
	return &JobServer{
		AppID:  appid,
		Name:   name,
		Status: TASK_NOT_START,
		done:   make(chan bool),
		task:   task,
		freq:   freq,
	}
}

// ID 返回任务 id
func (t *JobServer) ID() string {
	return t.AppID
}

// Start 启动任务
// delay 为首次启动任务的延迟时间
func (t *JobServer) Start(delay ...time.Duration) {
	if t.Status == TASK_STARTED {
		return
	}

	t.Status = TASK_STARTED
	go t.start(delay...)
}

// Stop 停止任务
func (t *JobServer) Stop() {
	t.Status = TASK_NOT_START
	t.done <- true
}

func (t *JobServer) start(delay ...time.Duration) {
	maxTimes := 30
	tryTimes := 0
	firstTime := true

	// 默认是立即开始
	d := 0 * time.Second
	if delay != nil && len(delay) > 0 {
		d = delay[0]
	}

	for {
		select {
		case done := <-t.done:
			if done {
				return
			}
		default:
		}

		if firstTime {
			time.Sleep(d)
			firstTime = false
		}

		logger.Info("任务 " + t.Name + " 开始 ...")
		err := t.task()

		if err != nil {
			if tryTimes >= maxTimes {
				t.Stop()
				return
			}

			tryTimes++
			logger.Error("任务 "+t.Name+" 执行失败, ", err.Error())

			// 30 秒后重试
			logger.Error("30 秒后自动重试 ...")
			time.Sleep(30 * time.Second)
			continue
		}

		tryTimes = 0
		logger.Info("任务 " + t.Name + " 结束")
		time.Sleep(t.freq)
	}
}
