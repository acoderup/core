package task

import (
	"errors"

	"github.com/acoderup/core/basic"
	"github.com/acoderup/core/logger"
)

var (
	TaskErr_CannotFindWorker  = errors.New("Cannot find fit worker.")
	TaskErr_TaskExecuteObject = errors.New("Task can only be executed executor")
)

type taskReqCommand struct {
	t Task
	n string
	g string
}

func (trc *taskReqCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()

	var err error
	var workerName string
	var worker *Worker
	if trc.g == "" {
		workerName, err = TaskExecutor.c.Get(trc.n)
		if err != nil {
			logger.Logger.Trace("taskReqCommand done error:", err)
			return err
		}
		worker = TaskExecutor.getWorker(workerName)
	} else {
		if wg, exist := TaskExecutor.getGroup(trc.g); wg != nil && exist {
			workerName, err = wg.c.Get(trc.n)
			if err != nil {
				logger.Logger.Trace("taskReqCommand done error:", err)
				return err
			}
			worker = wg.getWorker(workerName)
		} else {
			wg := TaskExecutor.AddGroup(trc.g)
			if wg != nil {
				workerName, err = wg.c.Get(trc.n)
				if err != nil {
					logger.Logger.Trace("taskReqCommand done error:", err)
					return err
				}
				worker = wg.getWorker(workerName)
			}
		}
	}
	if worker != nil {
		logger.Logger.Trace("task[", trc.n, "] dispatch-> worker[", workerName, "]")
		ste := SendTaskExe(worker.Object, trc.t)
		if ste == true {
			logger.Logger.Trace("SendTaskExe success.")
		} else {
			logger.Logger.Trace("SendTaskExe failed.")
		}
		return nil
	} else {
		logger.Logger.Tracef("[%v] worker is no found.", workerName)
		return TaskErr_CannotFindWorker
	}

}

func sendTaskReqToExecutor(t Task, name string, gname string) bool {
	if t == nil {
		logger.Logger.Trace("sendTaskReqToExecutor error,t is nil")
		return false
	}
	if t.getN() != nil && t.getS() == nil {
		logger.Logger.Error(name, " You must specify the source object task.")
		return false
	}
	return TaskExecutor.SendCommand(&taskReqCommand{t: t, n: name, g: gname}, true)
}

type fixTaskReqCommand struct {
	t Task
	n string
	g string
}

func (trc *fixTaskReqCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()

	var worker *Worker
	if trc.g == "" {
		worker = TaskExecutor.getFixWorker(trc.n)
		if worker == nil {
			worker = TaskExecutor.addFixWorker(trc.n)
		}
	} else {
		if wg, ok := TaskExecutor.getGroup(trc.g); ok && wg != nil {
			worker = wg.getFixWorker(trc.n)
			if worker == nil {
				worker = wg.addFixWorker(trc.n)
			}
		} else {
			wg := TaskExecutor.AddGroup(trc.g)
			if wg != nil {
				worker = wg.getFixWorker(trc.n)
				if worker == nil {
					worker = wg.addFixWorker(trc.n)
				}
			}
		}
	}

	if worker != nil {
		logger.Logger.Trace("task[", trc.n, "] dispatch-> worker[", trc.n, "]")
		ste := SendTaskExe(worker.Object, trc.t)
		if ste == true {
			logger.Logger.Trace("SendTaskExe success.")
		} else {
			logger.Logger.Trace("SendTaskExe failed.")
		}
		return nil
	} else {
		logger.Logger.Tracef("[%v] worker is no found.", trc.n)
		return TaskErr_CannotFindWorker
	}
}

func sendTaskReqToFixExecutor(t Task, name, gname string) bool {
	if t == nil {
		logger.Logger.Warn("sendTaskReqToExecutor error,t is nil")
		return false
	}
	if t.getN() != nil && t.getS() == nil {
		logger.Logger.Error(name, " You must specify the source object task.")
		return false
	}
	return TaskExecutor.SendCommand(&fixTaskReqCommand{t: t, n: name, g: gname}, true)
}

type broadcastTaskReqCommand struct {
	t Task
}

func (trc *broadcastTaskReqCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()

	trc.t.AddRefCnt(int32(len(TaskExecutor.workers)))
	for name, worker := range TaskExecutor.workers {
		//copy
		t := trc.t.clone(name)
		if t != nil {
			//logger.Logger.Trace("task[", t.name, "] dispatch-> worker[", name, "]")
			SendTaskExe(worker.Object, t)
		}
	}
	return nil
}

func sendTaskReqToAllExecutor(t Task) bool {
	if t == nil {
		logger.Logger.Warn("sendTaskReqToExecutor error,t is nil")
		return false
	}
	return TaskExecutor.SendCommand(&broadcastTaskReqCommand{t: t}, true)
}
