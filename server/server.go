package server

import "github.com/cloudtask/cloudtask-agent/api"
import "github.com/cloudtask/cloudtask-agent/cache"
import "github.com/cloudtask/cloudtask-agent/driver"
import "github.com/cloudtask/cloudtask-agent/etc"
import "github.com/cloudtask/cloudtask-agent/notify"
import "github.com/cloudtask/libtools/gounits/logger"
import "github.com/cloudtask/libtools/gzkwrapper"
import "github.com/cloudtask/common/models"

import (
	"time"
)

const (
	//joballoc refresh loop interval
	refreshAllocInterval = 15 * time.Second
	//driver dispatch loop interval
	dispatchDriverInterval = 1 * time.Second
)

//NodeServer is exported
type NodeServer struct {
	gzkwrapper.INodeNotifyHandler
	cache.ICacheHandler
	driver.IDriverHandler
	Key        string
	Runtime    string
	ConfigPath string
	AllocPath  string
	Data       *gzkwrapper.NodeData
	Worker     *gzkwrapper.Worker
	Cache      *cache.Cache
	Driver     *driver.Driver
	Notify     *notify.NotifySender
	stopCh     chan struct{}
}

//NewNodeServer is exported
func NewNodeServer(key string) (*NodeServer, error) {

	clusterArgs := etc.ClusterArgs()
	server := &NodeServer{
		Key:        key,
		Runtime:    clusterArgs.Location,
		ConfigPath: clusterArgs.Root + "/ServerConfigs",
		AllocPath:  clusterArgs.Root + "/JOBS-" + clusterArgs.Location,
		stopCh:     make(chan struct{}),
	}

	worker, err := gzkwrapper.NewWorker(key, clusterArgs, server)
	if err != nil {
		return nil, err
	}

	server.Worker = worker
	server.Data = worker.Data
	cacheArgs := etc.CacheArgs()
	server.Cache = cache.NewCache(cacheArgs, etc.ServerConfig, server)
	server.Driver = driver.NewDirver(cacheArgs.SaveDirectory, server)
	server.Notify = notify.NewNotifySender(clusterArgs.Location, key, worker.Data.IpAddr, etc.ServerConfig)
	return server, nil
}

//Startup is exported
func (server *NodeServer) Startup() error {

	var err error
	defer func() {
		if err != nil {
			server.nodeUnRegister()
			return
		}
		//start cache alloc monitor loop.
		go server.monitorCacheAllocLoop()
		//start driver dispatch loop.
		go server.dispatchDriverLoop()
	}()

	if err = server.nodeRegister(); err != nil {
		logger.ERROR("[#server#] server register to cluster error, %s", err)
		server.Worker.Close()
		return err
	}

	if err = server.initServerConfig(); err != nil {
		logger.ERROR("[#server#] init server config error, %s", err)
		return err
	}

	if err = server.openCache(); err != nil {
		logger.ERROR("[#server#] server open cache error, %s", err)
		return err
	}
	return nil
}

//Stop is exported
func (server *NodeServer) Stop() error {

	close(server.stopCh)
	server.closeServerConfig()
	server.Driver.Clear()
	server.closeCache()
	if err := server.nodeUnRegister(); err != nil {
		logger.ERROR("[#server] unregister to cluster error, %s", err.Error())
		return err
	}
	return nil
}

//nodeRegister is exported
func (server *NodeServer) nodeRegister() error {

	if err := server.Worker.Open(); err != nil {
		return err
	}

	attach := models.AttachEncode(&models.AttachData{
		JobMaxCount: etc.CacheArgs().MaxJobs,
	})
	return server.Worker.Signin(attach)
}

//nodeUnRegister is exported
func (server *NodeServer) nodeUnRegister() error {

	if err := server.Worker.Signout(); err != nil {
		return err
	}
	return server.Worker.Close()
}

//initServerConfig is exported
//initialize server congfigs and watching zk config path.
func (server *NodeServer) initServerConfig() error {

	//watch server config path.
	if err := server.Worker.WatchOpen(server.ConfigPath, server.OnSeverConfigsWatchHandlerFunc); err != nil {
		return err
	}

	//read config data.
	data, err := server.Worker.Get(server.ConfigPath)
	if err != nil {
		return err
	}
	//save data to etc serverConfig.
	return server.RefreshServerConfig(data)
}

//closeServerConfig is exported
func (server *NodeServer) closeServerConfig() {

	server.Worker.WatchClose(server.ConfigPath)
}

//RefreshServerConfig is exported
//save serverConfig, re-set to references objects.
func (server *NodeServer) RefreshServerConfig(data []byte) error {

	if err := etc.SaveServerConfig(data); err != nil {
		return err
	}

	api.RegisterStore("ServerConfig", etc.ServerConfig)
	server.Cache.SetServerConfig(etc.ServerConfig)
	server.Notify.SetServerConfig(etc.ServerConfig)
	return nil
}

//openCache is exported
//init cache & jobs alloc.
func (server *NodeServer) openCache() error {

	logger.INFO("[#server] server initialize......")
	server.Cache.LoadJobs()

	//init alloc path.
	if err := server.makeAllocPath(); err != nil {
		return err
	}

	//watch alloc path.
	if err := server.Worker.WatchOpen(server.AllocPath, server.OnCacheAllocWatchHandlerFunc); err != nil {
		return err
	}

	time.Sleep(server.Worker.Pulse) //wait jobserver realloced...
	//initialize cache alloc.
	version, err := server.RefreshCacheAlloc()
	if err != nil {
		return err
	}

	logger.INFO("[#server] init cache alloc %s, version is %d", server.AllocPath, version)
	logger.INFO("[#server] start cache dump cleaner.")
	server.Cache.StartDumpCleaner() //start cache dump cleaner
	return nil
}

//closeCache is exported
//clear cache & stop cache.
func (server *NodeServer) closeCache() {

	server.Cache.Clear()
	logger.INFO("[#server] clear cache.")
	server.Worker.WatchClose(server.AllocPath)
	server.Cache.StopDumpCleaner() //close cache dump cleaner
	logger.INFO("[#server] stop cache dump cleaner.")
}

//makeAllocPath is exported
func (server *NodeServer) makeAllocPath() error {

	ret, err := server.Worker.Exists(server.AllocPath)
	if err != nil {
		return err
	}

	if !ret {
		data, err := server.Cache.MakeAllocBuffer()
		if err != nil {
			return err
		}
		return server.Worker.Create(server.AllocPath, data)
	}
	return nil
}

//RefreshCacheAlloc is exported
func (server *NodeServer) RefreshCacheAlloc() (int, error) {

	//read alloc data.
	data, err := server.Worker.Get(server.AllocPath)
	if err != nil {
		return -1, err
	}

	//set alloc data to cache.
	version, err := server.Cache.SetAllocBuffer(server.Key, data)
	if err != nil {
		return -1, err
	}
	return version, nil
}

//monitorCacheAllocLoop is exported
func (server *NodeServer) monitorCacheAllocLoop() {

	for {
		runTicker := time.NewTicker(refreshAllocInterval)
		select {
		case <-runTicker.C:
			{
				runTicker.Stop()
				originVersion := server.Cache.GetAllocVersion()
				version, err := server.RefreshCacheAlloc()
				if err != nil {
					logger.ERROR("[#server#] monitor jobs alloc %s error, %s", server.AllocPath, err)
					continue
				}
				if originVersion < version {
					logger.INFO("[#server] monitor jobs alloc %s changed, version is %d", server.AllocPath, version)
				}
			}
		case <-server.stopCh:
			{
				runTicker.Stop()
				logger.INFO("[#server] monitor jobs alloc loop exited.")
				return
			}
		}
	}
}

//disposeDriver is exported
func (server *NodeServer) disposeDriver(event cache.CacheEvent, jobbase *models.JobBase) {

	logger.INFO("[#server#] dispose driver: %s jobid: %s", event, jobbase.JobId)
	switch event {
	case cache.CACHE_EVENT_JOBSET:
		{
			logger.INFO("[#server#] driver set %s.", jobbase.JobId)
			server.Driver.Set(jobbase)
		}
	case cache.CACHE_EVENT_JOBREMOVE:
		{
			logger.INFO("[#server#] driver remove %s.", jobbase.JobId)
			server.Driver.Remove(jobbase.JobId)
		}
	}
}

//dispatchDriverLoop is exported
func (server *NodeServer) dispatchDriverLoop() {

	for {
		driverTicker := time.NewTicker(dispatchDriverInterval)
		select {
		case <-driverTicker.C:
			{
				driverTicker.Stop()
				server.Driver.Dispatch()
			}
		case <-server.stopCh:
			{
				driverTicker.Stop()
				logger.INFO("[#server] dispatch driver loop exited.")
				return
			}
		}
	}
}
