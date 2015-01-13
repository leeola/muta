//
//
//
package muta

import "errors"

type TaskHandler func()

var DefaultTasker *Tasker = NewTasker()

func Task(name string, deps []string, h TaskHandler) error {
	return DefaultTasker.Task(name, deps, h)
}

func Run() {
	DefaultTasker.Run()
}

func NewTasker() *Tasker {
	return &Tasker{
		Tasks: make(map[string]*TaskerTask),
	}
}

type Tasker struct {
	Tasks map[string]*TaskerTask
}

func (tr *Tasker) Task(n string, ds []string, h TaskHandler) error {
	if tr.Tasks[n] != nil {
		return errors.New("Task already exists")
	}

	tr.Tasks[n] = &TaskerTask{
		Name:         n,
		Dependencies: ds,
		Handler:      h,
	}
	return nil
}

func (tr *Tasker) Run() {
}

func (tr *Tasker) RunTask(tn string) error {
	t := tr.Tasks[tn]
	if t.Dependencies != nil {
		for _, d := range t.Dependencies {
			tr.RunTask(d)
		}
	}
	t.Handler()
	return nil
}

type TaskerTask struct {
	Name         string
	Dependencies []string
	Handler      TaskHandler
}
