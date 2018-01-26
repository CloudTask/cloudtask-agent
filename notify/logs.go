package notify

import "github.com/cloudtask/libtools/gounits/logger"
import "github.com/cloudtask/common/models"

import (
	"time"
)

//SendLog is exported
func (sender *NotifySender) SendLog(msgid string, jobid string, command string, workdir string, state int,
	stdout string, errout string, execerr string, execat time.Time, exectimes float64) {

	logger.INFO("[#notify#] log %s job %s, state %d execat %s exectimes %.0f", msgid[:8], jobid, state, execat.Format("2006-01-02 15:04:05"), exectimes)
	jobLog := &models.JobLog{
		JobId:     jobid,
		MsgId:     msgid,
		Event:     execerr,
		Stat:      state,
		Location:  sender.Runtime,
		Command:   command,
		WorkDir:   workdir,
		IpAddr:    sender.IPAddr,
		StdOut:    stdout,
		ErrOut:    errout,
		ExecErr:   execerr,
		ExecAt:    execat,
		ExecTimes: exectimes,
		CreateAt:  time.Now().Unix() * 1000, //按毫秒单位存储
	}

	entry := &NotifyEntry{
		NotifyType: NOTIFY_LOG,
		MsgID:      msgid,
		Data:       jobLog,
	}
	sender.syncQueue.Push(entry)
}
