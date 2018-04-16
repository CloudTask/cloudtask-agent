package driver

import "github.com/cloudtask/libtools/gounits/logger"

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func NewExecDriver(name string, cmd string, env []string) (*ExecDriver, error) {

	driver := &ExecDriver{Running: false, ExecTimes: ZERO_TICK}
	driver.Command = exec.Command("/bin/bash", "-c", "cd "+name+" && "+cmd)
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
		proc := getProcess(driver.Command.Process.Pid)
		if proc == nil {
			return fmt.Errorf("stop job getProcess invalid.")
		}
		sendCtrlBreak(proc)                //发送退出消息
		afc := time.After(time.Second * 5) //最多等待5s让任务进程退出
		done := false
	NEW_TICK_DURATION:
		ticker := time.NewTicker(time.Millisecond * 100)
		for !done {
			select {
			case <-ticker.C:
				ticker.Stop()
				if _, err := exec.Command("ls", "/proc/"+strconv.Itoa(proc.Pid)).Output(); err != nil {
					done = true //任务进程已退出
				}
				goto NEW_TICK_DURATION
			case <-afc:
				ticker.Stop()
				if err := proc.Signal(os.Kill); err != nil {
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

func sendCtrlBreak(proc *os.Process) error {

	if err := proc.Signal(os.Interrupt); err != nil {
		logger.ERROR("[#driver#] execdriver sendctrlbreak failed.")
		return err
	}
	logger.INFO("[#driver#] execdriver sendctrlbreak successed.(%d)", proc.Pid)
	return nil
}

func getProcess(ppid int) *os.Process {

	buf, err := exec.Command("/bin/sh", "-c", "ps --ppid "+strconv.Itoa(ppid)+" | awk '{print $1}'").Output()
	if err != nil {
		return nil
	}

	var pid int = 0
	r := bufio.NewReader(bytes.NewReader(buf))
	for {
		data, _, err := r.ReadLine()
		if err != nil {
			return nil
		}
		p := strings.TrimSpace(string(data))
		if p == "PID" {
			continue
		}
		pid, _ = strconv.Atoi(p)
		break
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	return proc
}
