package notify

type NotifyType int

const (
	NOTIFY_MESSAGE NotifyType = iota + 1
	NOTIFY_LOG
	NOTIFY_MAIL
)

func (notifyType NotifyType) String() string {

	switch notifyType {
	case NOTIFY_MESSAGE:
		return "NOTIFY_MESSAGE"
	case NOTIFY_LOG:
		return "NOTIFY_LOG"
	case NOTIFY_MAIL:
		return "NOTIFY_MAIL"
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

//MessageContent is exported
type MessageContent struct {
	MessageName string `json:"MessageName"` //消息名称
	Password    string `json:"Password"`    //消息密码
	MessageBody string `json:"MessageBody"` //消息内容
	ContentType string `json:"ContentType"` //访问类型
	CallbackURI string `json:"CallbackUri"` //回调地址
	InvokeType  string `json:"InvokeType"`  //调用方式
}

//Mail is exported
type Mail struct {
	From        string   `json:"From"`
	To          string   `json:"To"`
	Subject     string   `json:"Subject"`
	Body        string   `json:"Body"`
	ContentType string   `json:"ContentType"`
	MailType    string   `json:"MailType"`
	SMTPSetting struct{} `json:"SmtpSetting"`
}
