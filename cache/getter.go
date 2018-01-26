package cache

import "github.com/cloudtask/libtools/gounits/compress/tarlib"
import "github.com/cloudtask/libtools/gounits/utils"
import "github.com/cloudtask/libtools/gounits/httpx"
import "github.com/cloudtask/libtools/gounits/logger"
import "github.com/cloudtask/libtools/gounits/system"
import "github.com/cloudtask/common/models"

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

/*
GetState is exported
拉取Job文件包状态类型定义
*/
type GetState int

/*
状态枚举定义
*/
const (
	GET_WAITING GetState = iota + 1 //等待下载
	GET_DOING                       //正在下载
)

/*
获取状态类型名称
*/
func (state GetState) String() string {

	switch state {
	case GET_WAITING:
		return "GET_WAITING"
	case GET_DOING:
		return "GET_DOING"
	}
	return ""
}

/*
ErrorCode is exported
错误值类型定义
*/
type ErrorCode int

/*
错误值枚举定义
*/
const (
	ERROR_GETJOBBASE  = -1000 //获取任务信息失败
	ERROR_READJOBBASE = -1001 //读取任务信息失败
	ERROR_PULLJOBFILE = -1002 //拉取任务文件失败
	ERROR_DECOMPRESS  = -1003 //解压任务文件失败
	ERROR_MAKECMDFILE = -1004 //上传任务文件为空，根据Cmd创建脚本文件失败
)

/*
JobGetError is exported
错误类型定义
*/
type JobGetError struct {
	Code  ErrorCode //错误编码
	Error error     //错误描述
}

func (jobGetError *JobGetError) String() string {

	if jobGetError.Error != nil {
		return jobGetError.Error.Error()
	}
	return ""
}

/*
JobGet is exported
待获取的Job状态信息
*/
type JobGet struct {
	JobId   string          //任务编号
	JobData *models.JobData //任务分配数据
	JobBase *models.JobBase //任务基础信息
	State   GetState        //获取状态
}

/*
JobGetter is exported
Job信息获取器
1、主要负责Job基础信息获取和Job文件包下载
2、定时恢复失败Job信息或文件的拉取
3、下载Job文件包成功负责解压生成目录结构，写入job.json信息文件
*/
type JobGetter struct {
	sync.RWMutex                      //互斥锁对象
	Root         string               //缓存根目录
	Recovery     time.Duration        //恢复拉取间隔
	serverConfig *models.ServerConfig //服务器配置信息
	exec         bool                 //是否在执行恢复拉取
	quit         chan bool            //退出下载
	gets         map[string]*JobGet   //下载任务集合
	handler      IJobGetterHandler    //回调句柄
	client       *httpx.HttpClient    //网络调用客户端
}

//NewJobGetter is exported
func NewJobGetter(args *CacheArgs, serverConfig *models.ServerConfig, handler IJobGetterHandler) *JobGetter {

	var recovery time.Duration
	dur, err := time.ParseDuration(args.PullRecovery)
	if err != nil {
		logger.WARN("[#cache#] pullrecovery parse duration err, %s, use default 60s.", err.Error())
		recovery = time.Duration(60) * time.Second
	} else {
		recovery = dur
	}

	client := httpx.NewClient().
		SetTransport(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			DisableKeepAlives:     false,
			MaxIdleConns:          25,
			MaxIdleConnsPerHost:   25,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout,
			ExpectContinueTimeout: http.DefaultTransport.(*http.Transport).ExpectContinueTimeout,
		})

	return &JobGetter{
		exec:         false,
		Root:         args.SaveDirectory,
		Recovery:     recovery,
		serverConfig: serverConfig,
		quit:         make(chan bool, 1),
		gets:         make(map[string]*JobGet, 0),
		handler:      handler,
		client:       client,
	}
}

//SetServerConfig is exported
//setting serverConfig
func (getter *JobGetter) SetServerConfig(serverConfig *models.ServerConfig) {

	getter.serverConfig = serverConfig
}

//Get is exported
//get a jobbase and jobfile
func (getter *JobGetter) Get(jobdata *models.JobData) {

	logger.INFO("[#cache#] getter get %s", jobdata.JobId)
	getter.Remove(jobdata.JobId)
	getter.Lock()
	defer getter.Unlock()
	jobbase, jobgeterror := getter.tryGetJobBase(jobdata)
	if jobgeterror != nil {
		jobget := &JobGet{JobId: jobdata.JobId, JobData: jobdata, JobBase: nil, State: GET_WAITING}
		go func() {
			getter.handler.OnJobGetterExceptionHandlerFunc("", jobget, jobgeterror)
		}()
		getter.gets[jobdata.JobId] = jobget
		getter.execute()
		return
	}
	workdir := getter.Root + "/" + jobbase.JobId + "/" + jobbase.FileCode
	if jobgeterror := getter.tryGetJobFile(workdir, jobbase); jobgeterror != nil {
		jobget := &JobGet{JobId: jobdata.JobId, JobData: jobdata, JobBase: jobbase, State: GET_WAITING}
		go func() {
			getter.handler.OnJobGetterExceptionHandlerFunc(workdir, jobget, jobgeterror)
		}()
		if jobgeterror.Code == ERROR_DECOMPRESS { //如果解压失败，不加入恢复gets尝试下载
			logger.ERROR("[#cache#] getter error %d, %s", jobgeterror.Code, jobgeterror.Error.Error())
			return
		}
		getter.gets[jobdata.JobId] = jobget
		getter.execute()
	} else {
		getter.save(jobbase) //写job.json
	}
	go func() { //=携程回调避免上层锁
		getter.handler.OnJobGetterHandlerFunc(workdir, jobbase)
	}()
}

func (getter *JobGetter) Remove(jobid string) {

	logger.INFO("[#cache#] getter remove %s", jobid)
	getter.Lock()
	if ret := utils.Contains(jobid, getter.gets); ret {
		delete(getter.gets, jobid)
	}
	getter.Unlock()
}

func (getter *JobGetter) Quit() {

	logger.INFO("[#cache#] cahce getter call quit.")
	getter.Lock()
	if getter.exec {
		getter.quit <- true
	}
	for jobid := range getter.gets {
		delete(getter.gets, jobid)
	}
	close(getter.quit)
	getter.Unlock()
	logger.INFO("[#cache#] cahce getter quited....")
}

func (getter *JobGetter) Check(jobbase *models.JobBase) bool {

	if strings.TrimSpace(jobbase.FileName) != "" {
		jobfile := getter.Root + "/jobs/" + jobbase.FileName //job文件缺失
		if ret := system.FileExist(jobfile); !ret {
			return false
		}
	}

	jobroot := getter.Root + "/" + jobbase.JobId
	if ret, _ := system.PathExists(jobroot + "/" + jobbase.FileCode); !ret {
		return false
	}

	if ret := system.FileExist(jobroot + "/job.json"); !ret {
		return false
	}
	return true
}

func (getter *JobGetter) Load() []*models.JobBase {

	logger.INFO("[#cache#] getter load jobs...")
	jobs := []*models.JobBase{}
	if err := system.MakeDirectory(getter.Root + "/jobs"); err != nil {
		logger.WARN("[#cache#] getter make root err, %s", err.Error())
		return jobs
	}

	fis, err := ioutil.ReadDir(getter.Root)
	if err != nil {
		logger.Error("[#cache#] getter read root err, %s", err.Error())
		return jobs
	}

	for _, fic := range fis {
		if fic.IsDir() {
			fpath := getter.Root + "/" + fic.Name() + "/job.json"
			if ret := system.FileExist(fpath); ret {
				fd, err := os.OpenFile(fpath, os.O_RDONLY, 0777)
				if err != nil {
					logger.ERROR("[#cache#] getter open job.json err, %s, %s", fic.Name(), err.Error())
					continue
				}
				jobbase := &models.JobBase{}
				if err := json.NewDecoder(fd).Decode(jobbase); err != nil {
					logger.ERROR("[#cache#] getter read job.json err, %s, %s", fic.Name(), err.Error())
					fd.Close()
					continue
				}
				if getter.Check(jobbase) {
					jobs = append(jobs, jobbase)
				}
				fd.Close()
			}
		}
	}
	return jobs
}

func (getter *JobGetter) save(jobbase *models.JobBase) error {

	logger.INFO("[#cache#] getter save job %s", jobbase.JobId)
	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(jobbase)
	if err != nil {
		logger.ERROR("[#cache#] getter save job.json encode err, %s, %s", jobbase.JobId, err.Error())
		return err
	}

	jobroot := getter.Root + "/" + jobbase.JobId
	err = ioutil.WriteFile(jobroot+"/job.json", buf.Bytes(), 0777)
	if err != nil {
		logger.ERROR("[#cache#] getter save job.json write err, %s, %s", jobbase.JobId, err.Error())
	}
	return err
}

func (getter *JobGetter) execute() {

	count := len(getter.gets)
	if count > 0 && !getter.exec {
		getter.exec = true
		logger.INFO("[#cache#] getter start execute %d....", count)
		go func() {
		NEW_TICK_DURATION:
			ticker := time.NewTicker(getter.Recovery)
			for {
				select {
				case <-getter.quit:
					{
						ticker.Stop()
						getter.exec = false
						logger.INFO("[#cache#] getter quit execute....")
						return
					}
				case <-ticker.C:
					{
						ticker.Stop()
						getter.Lock()
						getter.doGet()
						if len(getter.gets) == 0 {
							getter.exec = false
							getter.Unlock()
							logger.INFO("[#cache#] getter quit execute....")
							return
						}
						getter.Unlock()
						goto NEW_TICK_DURATION
					}
				}
			}
		}()
	}
}

func (getter *JobGetter) doGet() {

	for jobid, jobget := range getter.gets {
		iscallback := false
		if jobget.JobBase == nil {
			jobbase, jobgeterror := getter.tryGetJobBase(jobget.JobData)
			if jobgeterror != nil {
				getter.handler.OnJobGetterExceptionHandlerFunc("", jobget, jobgeterror)
				continue
			}
			jobget.JobBase = jobbase
			iscallback = true
		}

		var workdir string
		if jobget.JobBase != nil {
			workdir = getter.Root + "/" + jobget.JobBase.JobId + "/" + jobget.JobBase.FileCode
		}

		if jobget.State == GET_WAITING {
			jobget.State = GET_DOING
			if jobgeterror := getter.tryGetJobFile(workdir, jobget.JobBase); jobgeterror != nil {
				jobget.State = GET_WAITING
				getter.handler.OnJobGetterExceptionHandlerFunc(workdir, jobget, jobgeterror)
				if jobgeterror.Code == ERROR_DECOMPRESS {
					delete(getter.gets, jobid) //拉取成功解压失败，从恢复gets删除不再尝试下载.
					continue
				}
			} else {
				if jobget.JobBase != nil {
					delete(getter.gets, jobid)  //拉取文件,拉取成功删除jobget
					getter.save(jobget.JobBase) //写job.json
				}
			}
		}

		if iscallback {
			getter.handler.OnJobGetterHandlerFunc(workdir, jobget.JobBase)
		}
	}
}

func (getter *JobGetter) tryGetJobBase(jobdata *models.JobData) (*models.JobBase, *JobGetError) {

	logger.INFO("[#cache#] getter try getjobbase, %s", jobdata.JobId)
	resp, err := getter.client.Get(context.Background(), getter.serverConfig.CloudDataAPI+"/sys_jobs/"+jobdata.JobId, nil, nil)
	if err != nil {
		return nil, &JobGetError{Code: ERROR_GETJOBBASE, Error: fmt.Errorf("jobgetter getjobbase http error:%s", err.Error())}
	}

	defer resp.Close()
	statuscode := resp.StatusCode()
	if statuscode != http.StatusOK {
		return nil, &JobGetError{Code: ERROR_GETJOBBASE, Error: fmt.Errorf("jobgetter getjobbase http status code:%d", statuscode)}
	}

	job := &models.Job{}
	if err := resp.JSON(job); err != nil {
		return nil, &JobGetError{Code: ERROR_GETJOBBASE, Error: fmt.Errorf("jobgetter getjobbase decode data error:%s", err.Error())}
	}
	jobbase := parseJobBase(job)
	jobbase.Version = jobdata.Version
	return jobbase, nil
}

func (getter *JobGetter) tryGetJobFile(jobdirectory string, jobbase *models.JobBase) *JobGetError {

	if strings.TrimSpace(jobbase.FileName) != "" {
		return getter.pullJobFile(jobdirectory, jobbase)
	}
	return getter.makeJobCommandFile(jobdirectory, jobbase)
}

func (getter *JobGetter) pullJobFile(jobdirectory string, jobbase *models.JobBase) *JobGetError {

	logger.INFO("[#cache#] getter pull jobfile %s", jobbase.FileName)
	jobroot := getter.Root + "/" + jobbase.JobId                            //job所在根目录
	jobfile := getter.Root + "/jobs/" + jobbase.FileName                    //job文件下载到本地的路径
	remoteurl := getter.serverConfig.FileServerAPI + "/" + jobbase.FileName //job文件远程下载路径
	if ret := system.FileExist(jobfile); !ret {
		if err := getter.client.GetFile(context.Background(), jobfile, remoteurl, nil, nil); err != nil {
			err = errors.New("getter pull jobfile error " + remoteurl + ", " + err.Error())
			return &JobGetError{Code: ERROR_PULLJOBFILE, Error: err}
		}
	}

	if ret, _ := system.PathExists(jobdirectory); !ret {
		tempdir := jobroot + "/temp"
		if err := tarlib.TarAutoDeCompress(jobfile, tempdir); err != nil {
			os.RemoveAll(tempdir)
			os.RemoveAll(jobroot)
			os.Remove(jobfile)
			err = errors.New("getter decompress jobfile error " + remoteurl + ", " + err.Error())
			return &JobGetError{Code: ERROR_DECOMPRESS, Error: err}
		}
		system.DirectoryCopy(tempdir, jobdirectory)
		os.RemoveAll(tempdir)
	}
	return nil
}

func (getter *JobGetter) makeJobCommandFile(jobdirectory string, jobbase *models.JobBase) *JobGetError {

	if ret, _ := system.PathExists(jobdirectory); !ret {
		if err := system.MakeDirectory(jobdirectory); err != nil {
			return &JobGetError{Code: ERROR_MAKECMDFILE, Error: err}
		}
	}

	cmd, err := createCommandFile(jobdirectory, jobbase.Cmd)
	if err != nil {
		return &JobGetError{Code: ERROR_MAKECMDFILE, Error: err}
	}
	jobbase.Cmd = cmd
	return nil
}

func parseJobBase(job *models.Job) *models.JobBase {

	//根据文件名计算filecode(md5)
	encoder := md5.New()
	encoder.Write([]byte(job.FileName))
	fileCode := hex.EncodeToString(encoder.Sum(nil))

	jobbase := &models.JobBase{
		JobId:         job.JobId,
		JobName:       job.Name,
		FileName:      job.FileName,
		FileCode:      fileCode,
		Cmd:           job.Cmd,
		Env:           job.Env,
		Timeout:       job.Timeout,
		Version:       0,
		Schedule:      job.Schedule,
		NotifySetting: job.NotifySetting,
	}

	if jobbase.Env == nil {
		jobbase.Env = []string{}
	}

	if jobbase.Schedule == nil {
		jobbase.Schedule = []*models.Schedule{}
	}

	if jobbase.NotifySetting == nil {
		jobbase.NotifySetting = &models.NotifySetting{}
	}
	return jobbase
}
