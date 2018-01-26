package driver

import "github.com/cloudtask/common/models"

import (
	"fmt"
	"time"
)

type CoreHandler struct {
	ICoreHandler
}

type ExecCore struct {
	JobId      string           //任务编号
	WorkDir    string           //工作目录
	Exit       ExitState        //退出状态
	ExecAt     time.Time        //本次执行时间
	NextAt     time.Time        //下次执行时间
	Schedule   *models.Schedule //执行计划
	ExecDriver *ExecDriver      //执行驱动
	handler    ICoreHandler     //回调handler
}

func NewExecCore(jobid string, schedule *models.Schedule, handler ICoreHandler) *ExecCore {

	return &ExecCore{
		JobId:      jobid,
		Exit:       EXIT_NORMAL,
		ExecAt:     time.Time{},
		NextAt:     time.Time{},
		Schedule:   schedule,
		ExecDriver: nil,
		handler:    handler,
	}
}

func (core *ExecCore) GetExecTimes() float64 {

	if core.ExecDriver != nil {
		return core.ExecDriver.ExecTimes
	}
	return ZERO_TICK
}

func (core *ExecCore) GetExecDriverPipeBuffer() ([]byte, []byte) {

	if core.ExecDriver != nil {
		return core.ExecDriver.StdOut.Buffer, core.ExecDriver.ErrOut.Buffer
	}
	return nil, nil
}

func (core *ExecCore) Execute(seed time.Time, workdir string, cmd string, env []string) {

	if core.ExecDriver != nil {
		return
	}

	core.Exit = EXIT_NORMAL //退出状态复位
	core.WorkDir = workdir  //设置工作目录
	core.ExecAt = seed      //设置执行时间
	execdriver, err := NewExecDriver(workdir, cmd, env)
	if err != nil {
		go core.handler.OnCoreHandlerFunc(core, models.STATE_FAILED, fmt.Errorf("%s:%s", ErrExecuteException.Error(), err.Error()))
		return
	}

	core.ExecDriver = execdriver
	start := make(chan bool)
	go func() { //协程开启任务程序
		if err := core.ExecDriver.Start(start); err != nil { //start内部为start与wait，wait会阻塞，传入rc当start成功后可回调成功状态
			switch core.Exit {
			case EXIT_STOP: //通过stop命令退出，虽然强制关闭，但按流程退出.
				core.handler.OnCoreHandlerFunc(core, models.STATE_STOPED, nil)
			case EXIT_DEADLINE: //进程执行太久超时退出
				core.handler.OnCoreHandlerFunc(core, models.STATE_FAILED, ErrExecuteDeadline)
			default: //异常退出
				core.handler.OnCoreHandlerFunc(core, models.STATE_FAILED, fmt.Errorf("%s:%s", ErrExecuteException.Error(), err.Error()))
			}
		} else {
			core.handler.OnCoreHandlerFunc(core, models.STATE_STOPED, nil)
		}
		core.ExecDriver = nil
		core.Exit = EXIT_NORMAL
	}()
	ret := <-start
	if ret {
		go core.handler.OnCoreHandlerFunc(core, models.STATE_STARTED, nil)
	}
	close(start)
}

func (core *ExecCore) Close(state ExitState) error {

	if core.ExecDriver != nil {
		core.Exit = state
		if err := core.ExecDriver.Stop(); err != nil { //调用stop等待退出
			return fmt.Errorf("%s : %s", ErrExecuteTerminal.Error(), err.Error())
		}
	}
	return nil
}
