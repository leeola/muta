//
//
//
package muta

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/leeola/muta/logging"
)

type Handler func()
type ErrorHandler func() error
type ContextHandler func(Ctx *interface{}) error
type StreamHandler func() (*Stream, error)

var DefaultTasker *Tasker = NewTasker()

func Task(name string, args ...interface{}) error {
	return DefaultTasker.Task(name, args...)
}

func Run() {
	DefaultTasker.Run()
}

func NewTasker() *Tasker {
	return &Tasker{
		Tasks:  make(map[string]*TaskerTask),
		Logger: logging.DefaultLogger(),
	}
}

type Tasker struct {
	Tasks  map[string]*TaskerTask
	Logger *logging.Logger
}

type TaskerTask struct {
	Name           string
	Dependencies   []string
	Handler        Handler
	ErrorHandler   ErrorHandler
	StreamHandler  StreamHandler
	ContextHandler ContextHandler
}

func (tr *Tasker) Task(n string, args ...interface{}) error {
	if tr.Tasks[n] != nil {
		return errors.New("Task already exists")
	}

	ds := []string{}

	var (
		h  Handler
		er ErrorHandler
		sh StreamHandler
		ch ContextHandler
	)

	for _, arg := range args {
		v := reflect.ValueOf(arg)
		switch v.Type().String() {
		case "string":
			ds = append(ds, v.String())
		case "[]string":
			ds = append(ds, v.Interface().([]string)...)
		case "func()":
			h = v.Interface().(func())
			break
		case "func() error":
			er = v.Interface().(func() error)
			break
		case "func() (*muta.Stream, error)":
			sh = v.Interface().(func() (*Stream, error))
			break
		case "func(*interface {}) error":
			ch = v.Interface().(func(*interface{}) error)
			break
		default:
			return errors.New(fmt.Sprintf(
				"unsupported task argument type '%s'", v.Type().String(),
			))
		}
	}

	tr.Tasks[n] = &TaskerTask{
		Name:           n,
		Dependencies:   ds,
		Handler:        h,
		ErrorHandler:   er,
		StreamHandler:  sh,
		ContextHandler: ch,
	}

	return nil
}

func (tr *Tasker) Run() error {
	return tr.RunTask("default")
}

func (tr *Tasker) RunTask(tn string) (err error) {
	t := tr.Tasks[tn]
	if t == nil {
		return errors.New(fmt.Sprintf("Task \"%s\" does not exist.", tn))
	}

	tr.Logger.Info([]string{"Task"}, tn, "starting")
	defer func() {
		if err != nil {
			tr.Logger.Error([]string{"Task"}, tn,
				"returned an Error:", err)
		} else {
			tr.Logger.Info([]string{"Task"}, tn, "complete")
		}
	}()

	if t.Dependencies != nil {
		for _, d := range t.Dependencies {
			tr.RunTask(d)
		}
	}

	if t.Handler != nil {
		t.Handler()
	} else if t.ErrorHandler != nil {
		return t.ErrorHandler()
	} else if t.StreamHandler != nil {
		var s *Stream
		s, err = t.StreamHandler()
		if err != nil {
			return err
		}
		if s != nil {
			err = s.Start()
			if err != nil {
				return err
			}
		}
	} else if t.ContextHandler != nil {
		return errors.New("Not implemented")
	}

	return nil
}
