package main

import (
	"fmt"

	"github.com/leeola/muta"
)

func Hello() {
	fmt.Println("Hello")
}

func Readme() {
	fmt.Println(`
Nice, you ran Muta! Don't forget that you can get a task list
by running the following:

    $ muta -h
`)
}

func main() {
	// Add the "hello" task, with a func() handler
	muta.Task("hello", Hello)

	// Add the "world" task, with the "hello" dependency
	muta.Task("world", "hello", func() {
		fmt.Println("World")
	})

	// Add the Readme task
	muta.Task("readme", Readme)

	// Add the optional default task
	muta.Task("default", "world", "readme")

	// Run Te() or Start() to start Muta
	muta.Te()
}
