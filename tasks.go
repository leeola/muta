//
//
//
package muta

import (
	"errors"
	"fmt"
	"reflect"
)

type TaskHandler func()

var DefaultTasker *Tasker = NewTasker()

func Task(name string, args ...interface{}) error {
	return DefaultTasker.Task(name, args...)
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

func (tr *Tasker) Task(n string, args ...interface{}) error {
	ds := []string{}
	var h TaskHandler
	for _, arg := range args {
		v := reflect.ValueOf(arg)
		switch v.Type().String() {
		case "string":
			ds = append(ds, v.String())
		case "[]string":
			ds = append(ds, v.Interface().([]string)...)
		case "func()":
			h = v.Interface().(func())
			// Break on the first func found
			break
		default:
			return errors.New(fmt.Sprintf(
				"unsupported task argument type '%s'", v.Type().String(),
			))
		}
	}
	return tr.TaskStrict(n, ds, h)
}

func (tr *Tasker) TaskStrict(n string, ds []string, h TaskHandler) error {
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

func (tr *Tasker) Run() error {
	return tr.RunTask("default")
}

func (tr *Tasker) RunTask(tn string) error {
	t := tr.Tasks[tn]
	if t == nil {
		return errors.New(fmt.Sprintf("Task \"%s\" does not exist.", tn))
	}

	if t.Dependencies != nil {
		for _, d := range t.Dependencies {
			tr.RunTask(d)
		}
	}

	if t.Handler != nil {
		t.Handler()
	}
	return nil
}

type TaskerTask struct {
	Name         string
	Dependencies []string
	Handler      TaskHandler
}
