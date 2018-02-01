package driver

import "github.com/cloudtask/common/models"
import "github.com/cloudtask/libtools/gounits/utils"
import "github.com/cloudtask/libtools/gounits/logger"

import (
	"errors"
	"time"
)

var (
	ErrAllScheduleIsEmpty = errors.New("job all schedules isempty.")
	ErrAllScheduleDisable = errors.New("job all schedules disable.")
	ErrAllScheduleInvalid = errors.New("job all schedules invalid.")
)

/*
每个job对象无论有无schedule，都会构建一个pcore对象，用于当无schedule时调度.
该对象因为没有schedule，所以执行完后没有计算nextat.
调用pcore流程由action命令发起.
                                    select set job.core
                                             |
action satart()  -----> cores > 0  -----> select() -----> execute()
                 -----> cores == 0 -----> pcore    -----> execute()

action stop()    -----> job.core != nil -----> job.core -----> stop()
                 -----> cores == 0      -----> pcore    -----> stop()
*/

type Job struct {
	JobId      string               //任务编号
	Name       string               //任务名称
	Root       string               //工作根目录
	FileCode   string               //文件编码
	WorkDir    string               //工作目录
	Cmd        string               //执行命令
	Env        []string             //环境变量
	Timeout    int                  //执行超时(秒)
	ExecMaxSec int64                //执行最长时长(时间戳：UNIX时间戳)
	State      JobState             //执行状态
	LastExecAt time.Time            //最后一次执行时间
	LastError  error                //最后一次错误信息
	cores      map[string]*ExecCore //每一个schedule对应一个core, cores为调度集合.
	core       *ExecCore            //当job有schedule时，从调度集合中选择出来的有效core，为当前或即将调度的对象，并可计算nextat.
	pcore      *ExecCore            //当job无schedule时，发起action可用tempcore执行.
}

func NewJob(root string, jobbase *models.JobBase, handler ICoreHandler) *Job {

	job := &Job{
		JobId:      jobbase.JobId,
		Name:       jobbase.JobName,
		Root:       root,
		FileCode:   jobbase.FileCode,
		WorkDir:    root + "/" + jobbase.JobId + "/" + jobbase.FileCode,
		Cmd:        jobbase.Cmd,
		Env:        jobbase.Env,
		Timeout:    jobbase.Timeout,
		ExecMaxSec: 0,
		State:      JOB_WAITING,
		cores:      make(map[string]*ExecCore, 0),
		core:       nil,
		pcore:      NewExecCore(jobbase.JobId, nil, handler),
	}

	for _, schedule := range jobbase.Schedule {
		logger.INFO("[#driver#] createjob %s execcore schedule:%s", jobbase.JobId, schedule.Id)
		job.cores[schedule.Id] = NewExecCore(jobbase.JobId, schedule, handler)
	}
	return job
}

func (job *Job) SetJob(jobbase *models.JobBase, handler ICoreHandler) {

	job.Name = jobbase.JobName
	job.FileCode = jobbase.FileCode
	job.WorkDir = job.Root + "/" + jobbase.JobId + "/" + jobbase.FileCode
	job.Cmd = jobbase.Cmd
	job.Timeout = jobbase.Timeout
	for scheduleid, core := range job.cores {
		found := false
		for _, schedule := range jobbase.Schedule {
			if scheduleid == schedule.Id { //修改已存在的schedule
				core.Schedule = schedule
				found = true
				logger.INFO("[#driver#] changejob %s execcore schedule:%s", jobbase.JobId, schedule.Id)
				break
			}
		}
		if !found { //删除已不存在的schedule
			core.Close(EXIT_STOP)
			delete(job.cores, scheduleid)
			logger.INFO("[#driver#] removejob %s execcore schedule:%s", jobbase.JobId, scheduleid)
		}
	}

	for _, schedule := range jobbase.Schedule { //添加新schedule到core
		if ret := utils.Contains(schedule.Id, job.cores); !ret {
			job.cores[schedule.Id] = NewExecCore(jobbase.JobId, schedule, handler)
			logger.INFO("[#driver#] createjob %s execcore schedule:%s", jobbase.JobId, schedule.Id)
		}
	}
}

func (job *Job) Select() error {

	logger.INFO("[#driver#] job %s select core...", job.JobId)
	selectcores := []*ExecCore{}
	seed := time.Now()
	disables := 0
	job.core = nil
	for _, core := range job.cores {
		if core.Schedule.Enabled == 0 { //忽略未开启的schedule
			disables += 1
			logger.INFO("[#driver#] job %s schedule %s disabled.", job.JobId, core.Schedule.Id)
			continue
		}
		nextat, err := CalcSchedule(core.Schedule, seed)
		if err != nil { //schedule计算失败，时间无效.
			logger.ERROR("[#driver#] job %s schedule %s error %s.", job.JobId, core.Schedule.Id, err.Error())
			continue
		}
		if ret, _ := isExpired(core.Schedule, seed); ret { //schedule已过期，超过了EndDate.
			logger.ERROR("[#driver#] job %s schedule %s expired.", job.JobId, core.Schedule.Id)
			continue
		}
		core.NextAt = nextat
		selectcores = append(selectcores, core)
	}

	corecount := len(job.cores)
	if corecount == 0 { //无schedules
		logger.INFO("[#driver#] job %s schedules isempty.", job.JobId)
		return ErrAllScheduleIsEmpty
	} else {
		if disables == corecount { //schedule enabled已全部关闭
			logger.INFO("[#driver#] job %s all schedule disabled.", job.JobId)
			return ErrAllScheduleDisable
		}
	}

	var core *ExecCore = nil
	if len(selectcores) > 0 { //找出最近一次nextat
		for i := 0; i < len(selectcores); i++ {
			if core == nil {
				core = selectcores[i]
			} else {
				if core.NextAt.Sub(selectcores[i].NextAt).Seconds() > ZERO_TICK {
					core = selectcores[i]
				}
			}
		}
		job.core = core
		logger.INFO("[#driver#] job %s schedule %s nextat %s", job.JobId, core.Schedule.Id, core.NextAt.String())
		return nil
	}
	logger.ERROR("[#driver#] job %s all schedule invalid.", job.JobId)
	return ErrAllScheduleInvalid
}

func (job *Job) CheckWithTimeout(seed time.Time) {

	if job.ExecMaxSec > 0 && seed.Unix()-job.ExecMaxSec > 0 {
		logger.INFO("[#driver#] job %s exec timeout.", job.JobId)
		job.Close(EXIT_DEADLINE)
	}
}

func (job *Job) Execute(seed time.Time, force bool) {

	if !force { //定时调度, 采用job.core对象
		if job.core != nil && seed.Sub(job.core.NextAt).Seconds() > ZERO_TICK {
			logger.INFO("[#driver#] job %s !force execute %s", job.JobId, job.WorkDir)
			calcMaxSec(job, seed)
			job.core.Execute(seed, job.WorkDir, job.Cmd, job.Env)

		}
	} else { //强制执行, 用job.pcore对象
		logger.INFO("[#driver#] job %s force execute %s", job.JobId, job.WorkDir)
		calcMaxSec(job, seed)
		job.pcore.Execute(seed, job.WorkDir, job.Cmd, job.Env)
	}
}

func (job *Job) Close(state ExitState) {

	job.ExecMaxSec = 0
	if job.core != nil {
		job.core.Close(state)
	}
	job.pcore.Close(state)
	logger.INFO("[#driver#] job %s execute close, state %s.", job.JobId, state.String())
}

func calcMaxSec(job *Job, seed time.Time) {

	job.ExecMaxSec = 0
	if job.Timeout > 0 {
		job.ExecMaxSec = seed.Unix() + (int64)(job.Timeout)
	}
}
