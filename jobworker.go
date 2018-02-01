package main

import "github.com/cloudtask/cloudtask-agent/api"
import "github.com/cloudtask/cloudtask-agent/etc"
import "github.com/cloudtask/cloudtask-agent/server"
import "github.com/cloudtask/libtools/gounits/flocker"
import "github.com/cloudtask/libtools/gounits/logger"
import "github.com/cloudtask/libtools/gounits/rand"
import "github.com/cloudtask/libtools/gounits/system"

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
)

//JobWorker is exported
type JobWorker struct {
	RetryStartup bool
	Locker       *flocker.FileLocker
	NodeServer   *server.NodeServer
	APIServer    *api.Server
}

//AppCode is exported
var AppCode string

func init() {

	if appFile, err := exec.LookPath(os.Args[0]); err == nil {
		AppCode, _ = system.ReadFileMD5Code(appFile)
	}
}

//NewJobWorker is exported
func NewJobWorker() (*JobWorker, error) {

	var filePath string
	flag.StringVar(&filePath, "f", "./etc/jobworker.yaml", "jobworker etc file.")
	flag.Parse()
	if err := etc.New(filePath); err != nil {
		return nil, err
	}

	logConfigs := etc.LoggerConfigs()
	if logConfigs == nil {
		return nil, fmt.Errorf("logger configs invalid.")
	}
	logger.OPEN(logConfigs)

	key, err := rand.UUIDFile("./jobworker.key") //服务器唯一标识文件
	if err != nil {
		return nil, err
	}

	var fLocker *flocker.FileLocker
	if pidFile := etc.PidFile(); pidFile != "" {
		fLocker = flocker.NewFileLocker(pidFile, 0)
	}

	nodeServer, err := server.NewNodeServer(key)
	if err != nil {
		return nil, err
	}

	api.RegisterStore("AppCode", AppCode)
	api.RegisterStore("SystemConfig", etc.SystemConfig)
	api.RegisterStore("Cache", nodeServer.Cache)
	api.RegisterStore("Driver", nodeServer.Driver)
	api.RegisterStore("NodeKey", nodeServer.Key)
	api.RegisterStore("NodeData", nodeServer.Data)
	apiServer := api.NewServer(etc.SystemConfig.API.Hosts, etc.SystemConfig.API.EnableCors, nil)

	return &JobWorker{
		RetryStartup: etc.RetryStartup(),
		Locker:       fLocker,
		NodeServer:   nodeServer,
		APIServer:    apiServer,
	}, nil
}

//Startup is exported
func (worker *JobWorker) Startup() error {

	var err error
	for {
		if err != nil {
			if worker.RetryStartup == false {
				return err
			}
			time.Sleep(time.Second * 10) //retry, after sleep 10 seconds.
		}

		worker.Locker.Unlock()
		if err = worker.Locker.Lock(); err != nil {
			logger.ERROR("[#main#] pidfile lock error, %s", err)
			continue
		}

		if err = worker.NodeServer.Startup(); err != nil {
			logger.ERROR("[#main#] start server failure.")
			continue
		}
		break
	}

	go func() {
		logger.INFO("[#main#] API listener: %s", worker.APIServer.ListenHosts())
		if err := worker.APIServer.Startup(); err != nil {
			logger.ERROR("[#main#] API startup error, %s", err.Error())
		}
	}()
	logger.INFO("[#main#] jobworker started.")
	logger.INFO("[#main#] runtime %s, key:%s", worker.NodeServer.Runtime, worker.NodeServer.Key)
	return nil
}

//Stop is exported
func (worker *JobWorker) Stop() error {

	worker.Locker.Unlock()
	if err := worker.NodeServer.Stop(); err != nil {
		logger.ERROR("[#main#] stop server failure.")
		return err
	}
	logger.INFO("[#main#] jobworker stoped.")
	logger.CLOSE()
	return nil
}
