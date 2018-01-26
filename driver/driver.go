package driver

import "github.com/cloudtask/common/models"
import "github.com/cloudtask/libtools/gounits/logger"

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

//Driver is exported
type Driver struct {
	sync.RWMutex
	CoreHandler
	Root    string
	jobs    map[string]*Job
	handler IDriverHandler
}

//NewDirver is exported
func NewDirver(root string, handler IDriverHandler) *Driver {

	return &Driver{
		Root:    root,
		jobs:    make(map[string]*Job, 0),
		handler: handler,
	}
}

//Set is exported
func (driver *Driver) Set(jobbase *models.JobBase) {

	driver.Lock()
	if _, ret := driver.jobs[jobbase.JobId]; ret {
		logger.INFO("[#driver#] driver jobChange %s.", jobbase.JobId)
		driver.jobChange(jobbase)
	} else {
		logger.INFO("[#driver#] driver jobCreate %s.", jobbase.JobId)
		driver.jobCreate(jobbase)
	}
	driver.Unlock()
}

//Remove is exported
func (driver *Driver) Remove(jobid string) {

	driver.Lock()
	if job, ret := driver.jobs[jobid]; ret {
		job.Close(EXIT_STOP)
		delete(driver.jobs, jobid)
		logger.INFO("[#driver#] driver jobRemove %s.", jobid)
	}
	driver.Unlock()
}

//Clear is exported
func (driver *Driver) Clear() {

	driver.Lock()
	for _, job := range driver.jobs {
		job.Close(EXIT_STOP)
		delete(driver.jobs, job.JobId)
		logger.INFO("[#driver#] driver clearjob %s.", job.JobId)
	}
	driver.Unlock()
}

//Dispatch is exported
func (driver *Driver) Dispatch() {

	driver.Lock()
	for _, job := range driver.jobs {
		seed := time.Now()
		switch job.State {
		case JOB_WAITING:
			job.Execute(seed, false) //调度正处于等待状态的job
		case JOB_RUNNING:
			job.CheckWithTimeout(seed) //检查是否超过执行时间
		}
	}
	driver.Unlock()
}

//Action is exported
func (driver *Driver) Action(jobid string, action string) {

	driver.Lock()
	job := driver.jobs[jobid]
	if job != nil {
		logger.INFO("[#driver#] driver job %s action %s.", jobid, action)
		switch strings.ToLower(action) {
		case "start":
			{
				if job.State == JOB_WAITING {
					logger.INFO("[#driver#] driver start job %s.", job.JobId)
					job.Execute(time.Now(), true)
				}
			}
		case "stop":
			{
				if job.State == JOB_RUNNING {
					logger.INFO("[#driver#] driver stop job %s.", job.JobId)
					job.Close(EXIT_STOP)
				} else {
					var (
						err   error
						state int
					)
					state = models.STATE_STOPED
					if job.LastError != nil {
						err = job.LastError
						state = models.STATE_FAILED
					}
					if len(job.cores) > 0 {
						if e := job.Select(); e != nil {
							if e == ErrAllScheduleInvalid {
								state = models.STATE_FAILED
								if err != nil {
									err = fmt.Errorf("#1,%s\n#2,%s", err, e)
								} else {
									err = e
								}
							}
						}
					}
					execat := job.LastExecAt
					nextat := time.Time{}
					if job.core != nil {
						nextat = job.core.NextAt
					}
					context := driver.NewStopedContext(job, execat, nextat, err)
					driver.StopedHandleFunc(state, context)
				}
			}
		}
	}
	driver.Unlock()
}

func (driver *Driver) jobChange(jobbase *models.JobBase) {

	job := driver.jobs[jobbase.JobId]
	if job != nil {
		job.SetJob(jobbase, driver)
		driver.jobSelect(job)
	}
}

func (driver *Driver) jobCreate(jobbase *models.JobBase) {

	job := NewJob(driver.Root, jobbase, driver)
	if job != nil {
		driver.jobSelect(job)
		driver.jobs[jobbase.JobId] = job //加入到调度器
	}
}

func (driver *Driver) jobSelect(job *Job) {

	nextat := time.Time{}
	if job.State != JOB_RUNNING {
		if err := job.Select(); err != nil {
			if err == ErrAllScheduleInvalid {
				context := driver.NewExecuteContext(job, nil, nextat, err)
				driver.ExecuteHandleFunc(models.STATE_FAILED, context)
			}
		} else {
			if job.core != nil {
				nextat = job.core.NextAt
			}
			context := driver.NewSelectContext(job, nextat)
			driver.SelectHandleFunc(context)
		}
	}
}

func (driver *Driver) OnCoreHandlerFunc(core *ExecCore, state int, err error) {

	driver.Lock()
	job := driver.jobs[core.JobId]
	if job != nil {
		nextat := time.Time{}
		if state == models.STATE_STARTED {
			if core.ExecDriver != nil {
				job.State = JOB_RUNNING
			}
		} else {
			job.State = JOB_WAITING
			if len(job.cores) > 0 { //当有schedule时再计算nextat并选择core.
				if e := job.Select(); e != nil {
					if e == ErrAllScheduleInvalid {
						state = models.STATE_FAILED
						if err != nil {
							err = fmt.Errorf("#1,%s\n#2,%s", err, e)
						} else {
							err = e
						}
					}
				}
			}
			if job.core != nil {
				nextat = job.core.NextAt
			}
			job.LastExecAt = core.ExecAt
			job.LastError = err
		}
		//回调执行状态(启动/停止)
		context := driver.NewExecuteContext(job, core, nextat, err)
		driver.ExecuteHandleFunc(state, context)
	}
	driver.Unlock()
}
