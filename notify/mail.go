package notify

import "github.com/cloudtask/libtools/gounits/logger"
import "github.com/cloudtask/libtools/gounits/rand"
import "github.com/cloudtask/common/models"

import (
	"fmt"
	"time"
)

//SendMail is exported
func (sender *NotifySender) SendMail(jobid string, name string, notifysetting *models.NotifySetting,
	workdir string, state int, stdout string, errout string, execerr string, execat time.Time, exectimes float64) {

	msgid := rand.UUID(true)
	logger.INFO("[#notify#] mail %s job %s, state %d execat %s exectimes %.0f", msgid[:8], jobid, state, execat.Format("2006-01-02 15:04:05"), exectimes)
	if notifysetting == nil {
		return
	}

	var (
		prefix    string
		issucceed bool
		notify    *models.Notify
	)

	switch state {
	case models.STATE_STOPED:
		prefix = OKSubjectPrefix
		notify = &notifysetting.Succeed
		issucceed = true
	case models.STATE_FAILED:
		prefix = FAILSubjectPrefix
		notify = &notifysetting.Failed
		issucceed = false
	}

	var execatstring string
	if !execat.IsZero() {
		execatstring = execat.String()
	}

	if notify != nil {
		if !notify.Enabled {
			logger.INFO("[#notify#] mail %s enabled is false, return.", jobid)
			return
		}

		data := map[string]interface{}{
			"isSucceed": issucceed,
			"stdout":    stdout,
			"errout":    errout,
			"jobName":   name,
			"directory": workdir,
			"location":  sender.Runtime,
			"server":    sender.IPAddr,
			"execat":    execatstring,
			"execerr":   execerr,
			"duration":  fmt.Sprintf("%.2f", exectimes) + "sec",
			"content":   notify.Content,
		}

		entry := &NotifyEntry{
			NotifyType: NOTIFY_MAIL,
			MsgID:      msgid,
			To:         notify.To,
			Subject:    prefix + notify.Subject,
			Data:       data,
		}
		sender.syncQueue.Push(entry)
	}
}
