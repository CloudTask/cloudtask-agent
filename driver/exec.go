package driver

import "github.com/cloudtask/libtools/gounits/logger"

import (
	"errors"
	"io"
	"io/ioutil"
	"os/exec"
)

/*
执行错误定义
*/
var (
	//执行超时(超过job.ExecMaxSec阀值)
	ErrExecuteDeadline = errors.New("the job has been executed for too long and has exceeded the timeout threshold.")
	//执行异常
	ErrExecuteException = errors.New("job execute exception")
	//执行终止
	ErrExecuteTerminal = errors.New("job execute terminal error")
)

/*
ExecDriver 接口定义
*/
type IExecDriver interface {
	//启动任务
	Start(start chan<- bool) error
	//停止任务
	Stop() error
	//设置exec.cmd输入输出管道
	SetCommandPipe() error
	//读取exec.cmd管道数据
	ReadCommandPipeBuffer(stdoutCh chan<- []byte, erroutCh chan<- []byte)
}

/*
StdOutput 输出数据结构定义
*/
type StdOutput struct {
	Reader io.ReadCloser //输出读取对象(stdout或stderr)
	Buffer []byte        //输出数据
}

/*
ExecDriver 任务执行体
负责任务执行的生命期和状态.
*/
type ExecDriver struct {
	Running   bool      //执行状态
	Command   *exec.Cmd //执行对象
	ExecTimes float64   //总执行时长
	StdOut    StdOutput //标准输出
	ErrOut    StdOutput //错误输出
}

/*
SetCommandPipe 设置command对象管道
指定程序标准输出到StdOut.reader & ErrOut.reader
设置失败返回error.
*/
func (driver *ExecDriver) SetCommandPipe() error {

	logger.INFO("[#driver#] set command pipe.")
	//设置stdout管道输出
	stdout, err := driver.Command.StdoutPipe()
	if err != nil {
		logger.ERROR("[#driver#] set stdoutpipe error:%s", err)
		return err
	}
	driver.StdOut.Reader = stdout
	//设置errout管道输出
	stderr, err := driver.Command.StderrPipe()
	if err != nil {
		logger.ERROR("[#driver#] set stderrpipe error:%s", err)
		return err
	}
	driver.ErrOut.Reader = stderr
	return nil
}

/*
ReadCommandPipeBuffer 读取任务输出管道数据
StdOut.Buffer:标准输出
ErrOut.Buffer:错误输出
分别开启携程读取管道数据，避免读取阻塞.
若不开启携程，有种情况会出现读取stderr数据导致stdout读取阻塞.
*/
func (driver *ExecDriver) ReadCommandPipeBuffer(stdoutCh chan<- []byte, erroutCh chan<- []byte) {

	logger.INFO("[#driver#] read command pipe.")
	//携程读取stdout管道数据
	go func() {
		buf, err := ioutil.ReadAll(driver.StdOut.Reader)
		if err != nil {
			logger.ERROR("[#driver#] read stdoutpipe error:%s", err)
		}
		stdoutCh <- buf
	}()
	//携程读取errout管道数据
	go func() {
		buf, err := ioutil.ReadAll(driver.ErrOut.Reader)
		if err != nil {
			logger.ERROR("[#driver#] read stderrpipe error:%s", err)
		}
		erroutCh <- buf
	}()
}
