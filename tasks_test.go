package muta

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewTasker(t *testing.T) {
	Convey("Should initialize the Tasks map", t, func() {
		tr := NewTasker()
		So(tr.Tasks, ShouldNotBeNil)
	})
}

func TestTaskerTask(t *testing.T) {
	Convey("Should add a task", t, func() {
		ta := NewTasker()
		err := ta.Task("a", []string{}, func() {})
		So(err, ShouldBeNil)
		So(len(ta.Tasks), ShouldEqual, 1)
	})

	Convey("Should not allow replacing tasks", t, func() {
		ta := NewTasker()
		err := ta.Task("a", []string{}, func() {})
		So(err, ShouldBeNil)
		err = ta.Task("a", []string{}, func() {})
		So(err, ShouldNotBeNil)
	})

	Convey("Should allow zero dependencies", t, func() {
		ta := NewTasker()
		err := ta.Task("a", func() {})
		So(err, ShouldBeNil)
	})

	Convey("Should allow many dependencies", t, func() {
		ta := NewTasker()
		err := ta.Task("a", "b", "c", func() {})
		So(err, ShouldBeNil)
		ds := ta.Tasks["a"].Dependencies
		So(ds, ShouldResemble, []string{"b", "c"})
	})

	Convey("Should allow no functions", t, func() {
		ta := NewTasker()
		err := ta.Task("a")
		So(err, ShouldBeNil)
	})

	Convey("Should concatenate dependency arrays", t, func() {
		ta := NewTasker()
		err := ta.Task("a", "b", []string{"c", "d"}, "e")
		So(err, ShouldBeNil)
		d := ta.Tasks["a"].Dependencies
		So(d, ShouldResemble, []string{"b", "c", "d", "e"})
	})
}

func TestTaskerRunTask(t *testing.T) {
	Convey("Should run a task", t, func() {
		ran := false
		ta := NewTasker()
		ta.Task("a", func() {
			ran = true
		})
		err := ta.RunTask("a")
		So(err, ShouldBeNil)
		So(ran, ShouldBeTrue)
	})

	Convey("Should return an error for a non-existent task", t, func() {
		ta := NewTasker()
		ta.Task("a", func() {})
		err := ta.RunTask("b")
		So(err, ShouldNotBeNil)
	})

	Convey("Should run task dependencies", t, func() {
		deps := []string{"b", "c"}
		called := []string{}
		ta := NewTasker()
		ta.Task("a", deps, func() {
			called = append(called, "a")
		})
		ta.Task("b", func() {
			called = append(called, "b")
		})
		ta.Task("c", func() {
			called = append(called, "c")
		})
		err := ta.RunTask("a")
		So(err, ShouldBeNil)
		So(called, ShouldContain, "b")
		So(called, ShouldContain, "c")
	})

	Convey("Should run dependencies even without a func", t, func() {
		called := false
		ta := NewTasker()
		ta.Task("a", "b")
		ta.Task("b", func() {
			called = true
		})
		err := ta.RunTask("a")
		So(err, ShouldBeNil)
		So(called, ShouldBeTrue)
	})

	Convey("Should error on circular dependencies like", t, func() {
		Convey("a[a]", nil)
		Convey("a[b], b[a]", nil)
		Convey("a[b], b[c], c[a]", nil)
	})
}
