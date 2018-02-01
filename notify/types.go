package notify

type NotifyType int

const (
	NOTIFY_MESSAGE NotifyType = iota + 1
	NOTIFY_LOG
)

func (notifyType NotifyType) String() string {

	switch notifyType {
	case NOTIFY_MESSAGE:
		return "NOTIFY_MESSAGE"
	case NOTIFY_LOG:
		return "NOTIFY_LOG"
	}
	return ""
}

//NotifyEntry is exported
type NotifyEntry struct {
	NotifyType
	MsgID   string
	To      string
	Subject string
	Data    interface{}
}
