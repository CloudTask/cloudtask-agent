package cache

import "github.com/cloudtask/common/models"

//CacheArgs is exported
type CacheArgs struct {
	MaxJobs       int
	SaveDirectory string
	AutoClean     bool
	CleanInterval string
	PullRecovery  string
}

//Cache is exported
type Cache struct {
	ServerConfig *models.ServerConfig
	dumpCleaner  *DumpCleaner
	jobStore     *JobStore
}

//NewCache is exported
func NewCache(args *CacheArgs, serverConfig *models.ServerConfig, handler ICacheHandler) *Cache {

	return &Cache{
		ServerConfig: serverConfig,
		dumpCleaner:  NewDumpCleaner(args),
		jobStore: NewJobStore(args, serverConfig,
			handler.OnJobCacheChangedHandlerFunc,
			handler.OnJobCacheExceptionHandlerFunc),
	}
}

//SetServerConfig is exported
//setting serverConfig
func (cache *Cache) SetServerConfig(serverConfig *models.ServerConfig) {

	cache.ServerConfig = serverConfig
	cache.jobStore.SetServerConfig(serverConfig)
}

//StartDumpCleaner is exported
func (cache *Cache) StartDumpCleaner() {

	cache.dumpCleaner.Start()
}

//StopDumpCleaner is exported
func (cache *Cache) StopDumpCleaner() {

	cache.dumpCleaner.Stop()
}

//LoadJobs is exported
//load local jobs
func (cache *Cache) LoadJobs() {

	cache.jobStore.LoadJobs()
}

//Clear is exported
//clear local all jobs & alloc
func (cache *Cache) Clear() {

	cache.jobStore.ClearJobs()
}

//MakeAllocBuffer is exported
func (cache *Cache) MakeAllocBuffer() ([]byte, error) {

	return cache.jobStore.MakeAllocBuffer()
}

//SetAllocBuffer is exported
//set jobs alloc
func (cache *Cache) SetAllocBuffer(key string, data []byte) (int, error) {

	if err := cache.jobStore.SetAllocBuffer(key, data); err != nil {
		return -1, err
	}
	version := cache.jobStore.GetAllocVersion()
	return version, nil
}

//GetAllocVersion is exported
//return jobsalloc version
func (cache *Cache) GetAllocVersion() int {

	return cache.jobStore.GetAllocVersion()
}

//GetJobsCount is exported
//return jobsalloc job count
func (cache *Cache) GetJobsCount() int {

	return cache.jobStore.GetJobsCount()
}

//GetJobs is exported
//return cache jobs
func (cache *Cache) GetJobs() []*models.JobBase {

	return cache.jobStore.GetJobs()
}

//GetJob is exported
//return a cache job
func (cache *Cache) GetJob(jobid string) *models.JobBase {

	return cache.jobStore.GetJob(jobid)
}
