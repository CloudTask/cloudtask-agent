package cache

import "github.com/cloudtask/common/models"
import "github.com/cloudtask/libtools/gounits/logger"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
)

//JobStore is exported
type JobStore struct {
	sync.RWMutex                                   //互斥锁对象
	IJobGetterHandler                              //getter回调句柄
	configs           *CacheConfigs                //配置参数对象
	alloc             *models.JobsAlloc            //任务分配表
	getter            *JobGetter                   //任务信息获取器
	jobs              map[string]*models.JobBase   //任务信息本地缓存
	changedCallback   JobCacheChangedHandlerFunc   //任务改变回调
	exceptionCallback JobCacheExceptionHandlerFunc //任务异常回调
}

//NewJobStore is exported
//jobs & alloc cache
func NewJobStore(centerAPI string, configs *CacheConfigs,
	changedCallback JobCacheChangedHandlerFunc,
	exceptionCallback JobCacheExceptionHandlerFunc) *JobStore {

	alloc := &models.JobsAlloc{
		Version: 0,
		Jobs:    make([]*models.JobData, 0),
	}

	store := &JobStore{
		configs:           configs,
		alloc:             alloc,
		jobs:              make(map[string]*models.JobBase, 0),
		changedCallback:   changedCallback,
		exceptionCallback: exceptionCallback,
	}

	store.getter = NewJobGetter(centerAPI, configs, store)
	return store
}

//GetAllocVersion is exported
//return jobs alloc version.
func (store *JobStore) GetAllocVersion() int {

	store.RLock()
	version := store.alloc.Version
	store.RUnlock()
	return version
}

//LoadJobs is exported
//load cache root jobs directory data.
func (store *JobStore) LoadJobs() {

	store.Lock()
	result := store.getter.Load()
	for _, jobbase := range result {
		jobbase.Version = 0 //first load, set version to memory is zero, wait re-alloc.
		store.jobs[jobbase.JobId] = jobbase
		logger.INFO("[#cache#] read jobid:%s fcode:%s version:%d", jobbase.JobId, jobbase.FileCode, jobbase.Version)
	}
	logger.INFO("[#cache#] cache jobs count %d.", len(store.jobs))
	store.Unlock()
}

//GetJobsCount is exported
//return jobs count from cache alloc.
func (store *JobStore) GetJobsCount() int {

	store.RLock()
	defer store.RUnlock()
	return len(store.alloc.Jobs)
}

//GetJobs is exported
//return jobs from cache alloc.
func (store *JobStore) GetJobs() []*models.JobBase {

	jobs := []*models.JobBase{}
	store.RLock()
	for _, jobdata := range store.alloc.Jobs {
		if jobbase, ret := store.jobs[jobdata.JobId]; ret {
			jobs = append(jobs, jobbase)
		}
	}
	store.RUnlock()
	return jobs
}

//GetJob is exported
//return a job from cache alloc.
func (store *JobStore) GetJob(jobid string) *models.JobBase {

	store.RLock()
	defer store.RUnlock()
	for _, jobdata := range store.alloc.Jobs {
		if jobdata.JobId == jobid {
			if jobbase, ret := store.jobs[jobid]; ret {
				return jobbase
			}
		}
	}
	return nil
}

//ClearJobs is exported
//clear all cache jobs and alloc.
func (store *JobStore) ClearJobs() {

	store.Lock()
	store.alloc.Jobs = []*models.JobData{}
	store.alloc.Version = 0
	store.getter.Quit()
	for jobid := range store.jobs {
		delete(store.jobs, jobid)
	}
	store.Unlock()
}

//MakeAllocBuffer is exported
func (store *JobStore) MakeAllocBuffer() ([]byte, error) {

	alloc := &models.JobsAlloc{Version: 0, Jobs: make([]*models.JobData, 0)}
	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(alloc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//SetAllocBuffer is exported
//alloc changed, set alloc data.
func (store *JobStore) SetAllocBuffer(key string, data []byte) error {

	store.Lock()
	defer store.Unlock()
	tempalloc := &models.JobsAlloc{
		Version: 0,
		Jobs:    make([]*models.JobData, 0),
	}

	if err := models.JobsAllocDeCode(data, tempalloc); err != nil {
		return err
	}

	if tempalloc.Version == 0 || tempalloc.Version == store.alloc.Version {
		return nil
	}

	if tempalloc.Version < store.alloc.Version && store.alloc.Version != 0 {
		return fmt.Errorf("josalloc version invalid. tempalloc:%d storealloc:%d", tempalloc.Version, store.alloc.Version)
	}

	for i := len(tempalloc.Jobs) - 1; i >= 0; i-- {
		if tempalloc.Jobs[i].Key != key { //筛选非本实例的任务
			tempalloc.Jobs = append(tempalloc.Jobs[:i], tempalloc.Jobs[i+1:]...)
		}
	}

	var found bool
	for i := len(store.alloc.Jobs) - 1; i >= 0; i-- {
		found = false
		jobdata := store.alloc.Jobs[i]
		for j := 0; j < len(tempalloc.Jobs); j++ {
			if jobdata.JobId == tempalloc.Jobs[j].JobId {
				found = true
				break
			}
		}
		if !found { //找出被删除的任务
			store.alloc.Jobs = append(store.alloc.Jobs[:i], store.alloc.Jobs[i+1:]...)
			store.getter.Remove(jobdata.JobId) //从下载器中删除
			if jobbase, ret := store.jobs[jobdata.JobId]; ret {
				jobbase.Version = 0 //remove job, re-set job version to zero, wait re-alloc.
				logger.INFO("[#cache#] CACHE_EVENT_JOBREMOVE ###REMOVE %v", jobdata)
				go store.changedCallback(CACHE_EVENT_JOBREMOVE, jobbase)
			}
		}
	}

	for i := 0; i < len(tempalloc.Jobs); i++ {
		found = false
		jobdata := tempalloc.Jobs[i]
		for j := 0; j < len(store.alloc.Jobs); j++ { //找出版本号已改变的任务(任务更新)
			if jobdata.JobId == store.alloc.Jobs[j].JobId {
				if jobdata.Version > store.alloc.Jobs[j].Version {
					store.alloc.Jobs[j] = jobdata
					jobbase := store.tryGet(jobdata)
					if jobbase != nil {
						logger.INFO("[#cache#] CACHE_EVENT_JOBSET ###CHANGE %v", jobdata)
						go store.changedCallback(CACHE_EVENT_JOBSET, jobbase)
					}
				}
				found = true
				break
			}
		}
		if !found { //加入新添加的任务
			store.alloc.Jobs = append(store.alloc.Jobs, jobdata)
			jobbase := store.tryGet(jobdata)
			if jobbase != nil {
				logger.INFO("[#cache#] CACHE_EVENT_JOBSET ###CREATE %v", jobdata)
				go store.changedCallback(CACHE_EVENT_JOBSET, jobbase)
			}
		}
	}
	store.alloc = tempalloc
	return nil
}

func (store *JobStore) tryGet(jobdata *models.JobData) *models.JobBase {

	jobbase, ret := store.jobs[jobdata.JobId]
	check := true
	if jobbase != nil { //job文件缺失或目录不全，需要重新调用get
		check = store.getter.Check(jobbase)
	}
	if !ret || jobbase.Version != jobdata.Version || !check {
		store.getter.Get(jobdata) //等待回调触发jobbase
		return nil
	}
	return jobbase
}

//OnJobGetterExceptionHandlerFunc is exported
func (store *JobStore) OnJobGetterExceptionHandlerFunc(workdir string, jobget *JobGet, jobgeterror *JobGetError) {

	if jobget != nil && jobgeterror != nil {
		store.Lock()
		if jobget.JobBase != nil {
			store.jobs[jobget.JobId] = jobget.JobBase
		}
		store.Unlock()
		logger.ERROR("[#cache#] CACHE_EVENT_JOBERROR ###ERROR(%d) %s %s", jobgeterror.Code, jobget.JobId, jobgeterror.Error.Error())
		store.exceptionCallback(CACHE_EVENT_JOBERROR, workdir, jobget, jobgeterror)
	}
}

//OnJobGetterHandlerFunc is exported
func (store *JobStore) OnJobGetterHandlerFunc(workdir string, jobbase *models.JobBase) {

	if jobbase != nil {
		store.Lock()
		for _, jobdata := range store.alloc.Jobs {
			if jobdata.JobId == jobbase.JobId && jobdata.Version == jobbase.Version {
				bstr := "###CREATE"
				if _, ret := store.jobs[jobdata.JobId]; ret {
					bstr = "###CHANGE"
				}
				logger.INFO("[#cache#] CACHE_EVENT_JOBSET %s %v", bstr, jobdata)
				store.jobs[jobbase.JobId] = jobbase
				store.changedCallback(CACHE_EVENT_JOBSET, jobbase)
				break
			}
		}
		store.Unlock()
	}
}
