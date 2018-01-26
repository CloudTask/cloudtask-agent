package driver

type JobState int

const (
	JOB_RUNNING JobState = iota + 1 //任务被调度状态
	JOB_WAITING                     //任务等待调度状态
)

func (state JobState) String() string {

	switch state {
	case JOB_RUNNING:
		return "JOB_RUNNING"
	case JOB_WAITING:
		return "JOB_WAITING"
	}
	return ""
}

type ExitState int

const (
	EXIT_NORMAL   ExitState = iota + 1 //无强制退出状态
	EXIT_STOP                          //强制停止退出(Action Stop 命令)
	EXIT_DEADLINE                      //超时强制退出(Execute Timeout)
)

func (state ExitState) String() string {

	switch state {
	case EXIT_NORMAL:
		return "EXIT_NORMAL"
	case EXIT_STOP:
		return "EXIT_STOP"
	case EXIT_DEADLINE:
		return "EXIT_DEADLINE"
	}
	return ""
}
