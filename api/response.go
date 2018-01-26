package api

import "github.com/cloudtask/common/models"

import (
	"errors"
)

var (
	ErrRequestSuccessed       = errors.New("request successed.")
	ErrRequestAccepted        = errors.New("request accepted.")
	ErrRequestResolveInvaild  = errors.New("request resolve invaild.")
	ErrRequestNotFound        = errors.New("request resource not found.")
	ErrRequestServerException = errors.New("request server exception.")
	ErrRequestAllocNotFound   = errors.New("request resource not found in cache alloc.")
)

//HandleResponse is exportyed
type HandleResponse interface {
	SetContent(content string)
	SetData(data interface{})
}

//ResponseImpl is exported
type ResponseImpl struct {
	HandleResponse `json:"-,omitempty"`
	Content        string      `json:"content"`
	Data           interface{} `json:"data,omitempty"`
}

//SetContent is exported
func (response *ResponseImpl) SetContent(content string) {
	response.Content = content
}

//SetData is exported
func (response *ResponseImpl) SetData(data interface{}) {
	response.Data = data
}

//GetJobsBaseResponse is exported
type GetJobsBaseResponse struct {
	JobBase []*models.JobBase `json:"jobbase"`
}

//GetJobBaseResponse is exported
type GetJobBaseResponse struct {
	JobBase *models.JobBase `json:"jobbase"`
}
