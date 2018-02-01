package server

import "github.com/cloudtask/cloudtask-agent/cache"
import "github.com/cloudtask/cloudtask-agent/driver"
import "github.com/cloudtask/libtools/gounits/logger"
import "github.com/cloudtask/libtools/gzkwrapper"
import "github.com/cloudtask/common/models"

import (
	"time"
)

func (server *NodeServer) OnZkWrapperNodeHandlerFunc(nodestore *gzkwrapper.NodeStore) {
}

func (server *NodeServer) OnZkWrapperPulseHandlerFunc(key string, nodedata *gzkwrapper.NodeData, err error) {

	if err != nil {
		logger.ERROR("[#server#] pulse keepalive error, %s", err)
		return
	}
}

func (server *NodeServer) OnCacheAllocWatchHandlerFunc(path string, data []byte, err error) {

	if err != nil {
		logger.ERROR("[#server#] watch alloc %s error, %s", path, err)
		return
	}

	originVersion := server.Cache.GetAllocVersion()
	version, err := server.Cache.SetAllocBuffer(server.Key, data)
	if err != nil {
		logger.ERROR("[#server#] watch alloc %s setting error, %s", server.AllocPath, err)
		return
	}

	if originVersion < version {
		logger.INFO("[#server] watch alloc %s changed, version is %d", server.AllocPath, version)
	}
}

func (server *NodeServer) OnJobCacheChangedHandlerFunc(event cache.CacheEvent, jobbase *models.JobBase) {

	if jobbase != nil {
		logger.INFO("[#server#] jobcache changed, jobid %s version %d event %s", jobbase.JobId, jobbase.Version, event)
		server.disposeDriver(event, jobbase)
	}
}

func (server *NodeServer) OnJobCacheExceptionHandlerFunc(event cache.CacheEvent, workdir string, jobget *cache.JobGet, jobgeterror *cache.JobGetError) {

	logger.ERROR("[#server#] jobcache exception, job %s version %d event %s code %d %s", jobget.JobId, jobget.JobData.Version, event, jobgeterror.Code, jobgeterror.Error.Error())
	execat := time.Now()
	server.Notify.SendExecuteMessage(jobget.JobId, models.STATE_FAILED, jobgeterror.String(), execat, time.Time{})
	server.Notify.SendLog(jobget.JobId, "", workdir, models.STATE_FAILED, "", "", jobgeterror.String(), execat, 0.000000)
}

func (server *NodeServer) OnDriverExecuteHandlerFunc(state int, context *driver.DriverContext) {

	logger.INFO("[#server#] driver execute, job %s state %s", context.Job.JobId, models.GetStateString(state))
	server.Notify.SendExecuteMessage(context.Job.JobId, state, context.ExecErr, context.ExecAt, context.NextAt)
	//当状态为: STATE_STARTED, 忽略日志与发邮件.
	//当状态为: STATE_STOPED | STATE_FAILED, 记录日志，处理邮件通知.
	if state != models.STATE_STARTED {
		server.Notify.SendLog(context.Job.JobId, context.Job.Cmd, context.Job.WorkDir, state, context.StdOut, context.ErrOut, context.ExecErr, context.ExecAt, context.ExecTimes)
		server.Notify.SendMail(context.Job.JobId, context.Job.Name, context.Job.NotifySetting, context.Job.WorkDir, state, context.StdOut, context.ErrOut, context.ExecErr, context.ExecAt, context.ExecTimes)
	}
}

func (server *NodeServer) OnDriverSelectHandlerFunc(context *driver.DriverContext) {

	logger.INFO("[#server#] driver select, job %s", context.Job.JobId)
	server.Notify.SendSelectMessage(context.Job.JobId, context.NextAt)
}

func (server *NodeServer) OnDriverStopedHandlerFunc(state int, context *driver.DriverContext) {

	logger.INFO("[#server#] driver stoped, job %s", context.Job.JobId)
	server.Notify.SendExecuteMessage(context.Job.JobId, state, context.ExecErr, context.ExecAt, context.NextAt)
}
