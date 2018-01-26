package driver

import (
	"time"
)

/*
  DriverContext上下文定义
*/
type DriverContext struct {
	Job       *Job
	StdOut    string
	ErrOut    string
	ExecErr   string
	ExecAt    time.Time
	NextAt    time.Time
	ExecTimes float64
}

/*
  NewExecuteContext构造
*/
func (driver *Driver) NewExecuteContext(job *Job, core *ExecCore, nextat time.Time, err error) *DriverContext {

	context := &DriverContext{
		Job:    job,
		NextAt: nextat,
	}

	if err != nil {
		context.ExecErr = err.Error()
	}

	if core != nil {
		stdout, errout := core.GetExecDriverPipeBuffer()
		context.StdOut = string(stdout)
		context.ErrOut = string(errout)
		context.ExecAt = core.ExecAt
		context.ExecTimes = core.GetExecTimes()
	}
	return context
}

/*
  NewSelectContext构造
*/
func (driver *Driver) NewSelectContext(job *Job, nextat time.Time) *DriverContext {

	return &DriverContext{
		Job:    job,
		NextAt: nextat,
	}
}

/*
  NewStopedContext构造
*/
func (driver *Driver) NewStopedContext(job *Job, execat time.Time, nextat time.Time, err error) *DriverContext {

	context := &DriverContext{
		Job:    job,
		ExecAt: execat,
		NextAt: nextat,
	}

	if err != nil {
		context.ExecErr = err.Error()
	}
	return context
}

/*
  Driver回调handler定义
*/
type IDriverHandler interface {
	//DriverContext Code = ERR_SCHEDULE_EXECUTE
	OnDriverExecuteHandlerFunc(state int, context *DriverContext)
	//DriverContext Code = ERR_SCHEDULE_INVALID
	OnDriverSelectHandlerFunc(context *DriverContext)
	//DriverContext Code = ERR_SCHEDULE_STOPED
	OnDriverStopedHandlerFunc(state int, context *DriverContext)
}

type DriverExecuteHandlerFunc func(state int, context *DriverContext)

func (fn DriverExecuteHandlerFunc) OnDriverExecuteHandlerFunc(state int, context *DriverContext) {
	fn(state, context)
}

type DriverSelectHandlerFunc func(context *DriverContext)

func (fn DriverSelectHandlerFunc) OnDriverSelectHandlerFunc(context *DriverContext) {
	fn(context)
}

type DriverStopedHandlerFunc func(state int, context *DriverContext)

func (fn DriverStopedHandlerFunc) OnDriverStopedHandlerFunc(state int, context *DriverContext) {
	fn(state, context)
}

func (driver *Driver) ExecuteHandleFunc(state int, context *DriverContext) {

	if context.Job != nil {
		driver.handler.OnDriverExecuteHandlerFunc(state, context)
	}
}

func (driver *Driver) SelectHandleFunc(context *DriverContext) {

	if context.Job != nil {
		driver.handler.OnDriverSelectHandlerFunc(context)
	}
}

func (driver *Driver) StopedHandleFunc(state int, context *DriverContext) {

	if context.Job != nil {
		driver.handler.OnDriverStopedHandlerFunc(state, context)
	}
}

type ICoreHandler interface {
	OnCoreHandlerFunc(core *ExecCore, state int, err error)
}

type CoreHandlerFunc func(core *ExecCore, state int, err error)

func (fn CoreHandlerFunc) OnCoreHandlerFunc(core *ExecCore, state int, err error) {
	fn(core, state, err)
}
