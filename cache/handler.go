package cache

import "github.com/cloudtask/common/models"

type CacheEvent string

const (
	CACHE_EVENT_JOBERROR  CacheEvent = "CACHE_EVENT_JOBERROR"
	CACHE_EVENT_JOBSET    CacheEvent = "CACHE_EVENT_JOBSET"
	CACHE_EVENT_JOBREMOVE CacheEvent = "CACHE_EVENT_JOBREMOVE"
)

//ICacheHandler is exported
type ICacheHandler interface {
	OnJobCacheChangedHandlerFunc(event CacheEvent, jobbase *models.JobBase)
	OnJobCacheExceptionHandlerFunc(event CacheEvent, workdir string, jobget *JobGet, jobgeterror *JobGetError)
}

type JobCacheChangedHandlerFunc func(event CacheEvent, jobbase *models.JobBase)

func (fn JobCacheChangedHandlerFunc) OnJobCacheChangedHandlerFunc(event CacheEvent, jobbase *models.JobBase) {
	fn(event, jobbase)
}

type JobCacheExceptionHandlerFunc func(event CacheEvent, workdir string, jobget *JobGet, jobgeterror *JobGetError)

func (fn JobCacheExceptionHandlerFunc) OnJobCacheExceptionHandlerFunc(event CacheEvent, workdir string, jobget *JobGet, jobgeterror *JobGetError) {
	fn(event, workdir, jobget, jobgeterror)
}

//IJobGetterHandler is exported
type IJobGetterHandler interface {
	OnJobGetterExceptionHandlerFunc(workdir string, jobget *JobGet, jobgeterror *JobGetError)
	OnJobGetterHandlerFunc(workdir string, jobbase *models.JobBase)
}

type JobGetterExceptionHandlerFunc func(workdir string, jobget *JobGet, jobgeterror *JobGetError)

func (fn JobGetterExceptionHandlerFunc) OnJobGetterExceptionHandlerFunc(workdir string, jobget *JobGet, jobgeterror *JobGetError) {
	fn(workdir, jobget, jobgeterror)
}

type JobGetterHandlerFunc func(workdir string, jobbase *models.JobBase)

func (fn JobGetterHandlerFunc) OnJobGetterHandlerFunc(workdir string, jobbase *models.JobBase) {
	fn(workdir, jobbase)
}
