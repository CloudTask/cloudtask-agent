package driver

import "github.com/cloudtask/libtools/gounits/logger"

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

func NewExecDriver(name string, cmd string, env []string) (*ExecDriver, error) {

	filePath, err := filepath.Abs(name + "/" + cmd)
	if err != nil {
		logger.ERROR("[#driver#] execdriver cmd file path error:%s", err)
		return nil, err
	}

	driver := &ExecDriver{Running: false, ExecTimes: ZERO_TICK}
	driver.Command = exec.Command("cmd", "/C", filePath)
	if err := driver.SetCommandPipe(); err != nil {
		logger.ERROR("[#driver#] execdriver setcommandpipe error:%s", err)
		return nil, err
	}
	driver.Command.Env = append(os.Environ(), env...)
	logger.INFO("[#driver#] execdriver create successed, %s", cmd)
	return driver, nil
}

func (driver *ExecDriver) Start(start chan<- bool) error {

	if driver.Command != nil {
		driver.Command.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		}
		logger.INFO("[#driver#] start execdriver")
		start_t := time.Now() //记录开始执行时间
		stdoutCh := make(chan []byte)
		erroutCh := make(chan []byte)
		driver.ReadCommandPipeBuffer(stdoutCh, erroutCh) //读取管道stdout、stderr输出数据
		defer func() {
			driver.StdOut.Buffer = <-stdoutCh
			driver.ErrOut.Buffer = <-erroutCh
			close(stdoutCh)
			close(erroutCh)
		}()
		if err := driver.Command.Start(); err != nil {
			start <- driver.Running
			logger.ERROR("[#driver#] start execdriver:%s", err)
			return err
		}
		driver.Running = true
		start <- driver.Running
		if err := driver.Command.Wait(); err != nil {
			driver.Running = false
			driver.ExecTimes = time.Now().Sub(start_t).Seconds() //计算执行时间差
			logger.ERROR("[#driver#] wait execdriver:%s", err)
			return err
		}
		driver.Running = false
		driver.ExecTimes = time.Now().Sub(start_t).Seconds() //计算执行时间差
		logger.INFO("[#driver#] execdriver over.")
		return nil
	}
	err := fmt.Errorf("start execdriver command invalid.")
	logger.ERROR("[#driver#] %s", err.Error())
	start <- driver.Running
	return err
}

func (driver *ExecDriver) Stop() error {

	logger.INFO("[#driver#] execdriver stop")
	if driver.Command != nil && driver.Command.Process != nil {
		sendCtrlBreak(driver.Command.Process.Pid) //发送退出消息
		afc := time.After(time.Second * 5)        //最多等待5s让任务进程退出
		done := false
	NEW_TICK_DURATION:
		ticker := time.NewTicker(time.Millisecond * 100)
		for !done {
			select {
			case <-ticker.C:
				ticker.Stop()
				if _, err := os.FindProcess(driver.Command.Process.Pid); err != nil {
					done = true //任务进程已退出
				}
				goto NEW_TICK_DURATION
			case <-afc:
				ticker.Stop()
				err := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprint(driver.Command.Process.Pid)).Run()
				if err != nil {
					logger.ERROR("[#driver#] execdriver kill:%s", err)
					return err
				}
				done = true
				goto NEW_TICK_DURATION
			}
		}
		logger.INFO("[#driver#] execdriver stop successed.")
		return nil
	}
	err := fmt.Errorf("stop execdriver command invalid.")
	logger.ERROR("[#driver#] %s", err.Error())
	return err
}

func sendCtrlBreak(pid int) error {

	dl, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		logger.ERROR("[#driver#] stop execdriver sendctrlbreak load kernel32 failed.")
		return err
	}

	proc, err := dl.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		logger.ERROR("[#driver#] stop execdriver sendctrlbreak getwin32 api failed.")
		return err
	}

	r, _, err := proc.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if r == 0 {
		logger.ERROR("[#driver#] execdriver sendctrlbreak failed.")
		return err
	}
	logger.INFO("[#driver#] execdriver sendctrlbreak successed.(%d)", pid)
	return nil
}
