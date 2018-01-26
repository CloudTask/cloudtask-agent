package notify

import "github.com/cloudtask/libtools/gounits/container"
import "github.com/cloudtask/libtools/gounits/httpx"
import "github.com/cloudtask/libtools/gounits/logger"
import "github.com/cloudtask/common/models"

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"net"
	"net/http"
	"time"
)

//NotifySender is exported
type NotifySender struct {
	Runtime      string
	Key          string
	IPAddr       string
	ServerConfig *models.ServerConfig
	client       *httpx.HttpClient
	syncQueue    *container.SyncQueue
}

//NewNotifySender is exported
func NewNotifySender(runtime string, key string, ipaddr string, serverConfig *models.ServerConfig) *NotifySender {

	client := httpx.NewClient().
		SetTransport(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 90 * time.Second,
			}).DialContext,
			DisableKeepAlives:     false,
			MaxIdleConns:          50,
			MaxIdleConnsPerHost:   50,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout,
			ExpectContinueTimeout: http.DefaultTransport.(*http.Transport).ExpectContinueTimeout,
		})

	notifySender := &NotifySender{
		Runtime:      runtime,
		Key:          key,
		IPAddr:       ipaddr,
		ServerConfig: serverConfig,
		client:       client,
		syncQueue:    container.NewSyncQueue(),
	}
	go notifySender.doPopLoop()
	return notifySender
}

//SetServerConfig is exported
//setting serverConfig
func (sender *NotifySender) SetServerConfig(serverConfig *models.ServerConfig) {

	sender.ServerConfig = serverConfig
}

func (sender *NotifySender) doPopLoop() {

	for {
		value := sender.syncQueue.Pop()
		if value != nil {
			entry := value.(*NotifyEntry)
			switch entry.NotifyType {
			case NOTIFY_MESSAGE:
				go sendMessage(sender.client, sender.ServerConfig, entry.MsgID, entry.Data)
			case NOTIFY_LOG:
				go sendLog(sender.client, sender.ServerConfig, entry.MsgID, entry.Data)
			case NOTIFY_MAIL:
				go sendMail(sender.client, sender.ServerConfig, entry.MsgID, entry.To, entry.Subject, entry.Data)
			}
		}
	}
}

func sendMessage(client *httpx.HttpClient, serverConfig *models.ServerConfig, msgid string, data interface{}) {

	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		logger.ERROR("[#notify#] message request %s encode error, %s", msgid, err.Error())
		return
	}

	msgArgs := serverConfig.MessageServer
	msgContent := MessageContent{
		MessageName: msgArgs.Name,
		Password:    msgArgs.Password,
		MessageBody: buf.String(),
		ContentType: msgArgs.ContentType,
		CallbackURI: msgArgs.Callback,
		InvokeType:  msgArgs.Invoke,
	}

	resp, err := client.PostJSON(context.Background(), msgArgs.APIAddr, nil, msgContent, nil)
	if err != nil {
		logger.ERROR("[#notify#] message request %s error, %s", msgid, err.Error())
		return
	}

	defer resp.Close()
	statusCode := resp.StatusCode()
	if statusCode >= http.StatusBadRequest {
		logger.ERROR("[#notify#] message request %s failure, %d", msgid, statusCode)
	}
}

func sendLog(client *httpx.HttpClient, serverConfig *models.ServerConfig, msgid string, data interface{}) {

	resp, err := client.PostJSON(context.Background(), serverConfig.CloudDataAPI+"/logs?strict=true", nil, data, nil)
	if err != nil {
		logger.ERROR("[#notify#] logs request %s error, %s", msgid, err.Error())
		return
	}

	defer resp.Close()
	statusCode := resp.StatusCode()
	if statusCode >= http.StatusBadRequest {
		logger.ERROR("[#notify#] logs request %s failure, %d", msgid, statusCode)
	}
}

func sendMail(client *httpx.HttpClient, serverConfig *models.ServerConfig, msgid string, to string, subject string, data interface{}) {

	var doc bytes.Buffer
	p := MailTemplate
	t := template.New("")
	t, _ = t.Parse(p)
	t.Execute(&doc, data)
	html := doc.String()

	mail := map[string]interface{}{
		"From":        "cloudtask@newegg.com",
		"To":          to,
		"Subject":     subject,
		"Body":        html,
		"ContentType": "Html",
		"MailType":    "Smtp",
		"SmtpSetting": map[string]interface{}{},
	}

	resp, err := client.PostJSON(context.Background(), serverConfig.NotifyAPI, nil, mail, nil)
	if err != nil {
		logger.ERROR("[#notify#] mail request %s error, %s", msgid, err.Error())
		return
	}

	defer resp.Close()
	statusCode := resp.StatusCode()
	if statusCode >= http.StatusBadRequest {
		logger.ERROR("[#notify#] mail request %s failure, %d", msgid, statusCode)
	}
}
