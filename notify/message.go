package notify

import "github.com/cloudtask/libtools/gounits/logger"
import "github.com/cloudtask/libtools/gounits/rand"
import "github.com/cloudtask/common/models"

import (
	"time"
)

//SendExecuteMessage is exported
func (sender *NotifySender) SendExecuteMessage(jobid string, state int, execerr string, execat time.Time, nextat time.Time) {

	msgid := rand.UUID(true)
	logger.INFO("[#notify#] message %s job %s, execute state %d execat %s nextat %s", msgid[:8], jobid, state, execat.Format("2006-01-02 15:04:05"), nextat.Format("2006-01-02 15:04:05"))
	jobExecute := &models.JobExecute{
		MsgHeader: models.MsgHeader{
			MsgName: models.MsgJobExecute,
			MsgId:   msgid,
		},
		JobId:     jobid,
		Location:  sender.Runtime,
		Key:       sender.Key,
		IPAddr:    sender.IPAddr,
		State:     state,
		ExecErr:   execerr,
		ExecAt:    execat,
		NextAt:    nextat,
		Timestamp: time.Now().UnixNano(),
	}

	entry := &NotifyEntry{
		NotifyType: NOTIFY_MESSAGE,
		MsgID:      msgid,
		Data:       jobExecute,
	}
	sender.syncQueue.Push(entry)
}

//SendSelectMessage is exported
func (sender *NotifySender) SendSelectMessage(jobid string, nextat time.Time) {

	msgid := rand.UUID(true)
	logger.INFO("[#notify#] message %s job %s, select nextat %s", msgid[:8], jobid, nextat.Format("2006-01-02 15:04:05"))
	jobSelect := &models.JobSelect{
		MsgHeader: models.MsgHeader{
			MsgName: models.MsgJobSelect,
			MsgId:   msgid,
		},
		JobId:     jobid,
		Location:  sender.Runtime,
		NextAt:    nextat,
		Timestamp: time.Now().UnixNano(),
	}

	entry := &NotifyEntry{
		NotifyType: NOTIFY_MESSAGE,
		MsgID:      msgid,
		Data:       jobSelect,
	}
	sender.syncQueue.Push(entry)
}
