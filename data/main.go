package data

import "fmt"

func helloRecover() {
	defer func() {
		recover()
	}()
	fmt.Println("hello")
}

func helloNotRecover() {
	go recoverWorld()
	fmt.Println("hello")
}

// Hello
// this is a public function
func Hello() {
	go world()
	fmt.Println("Hello")
}

// hello
// this is a private function
func hello() {
	defer func() {}()
	fmt.Println("hello")
}

func helloWorld() {
	defer world()
	fmt.Println("hello")
}

func world() {
	fmt.Println("world")
}

func recoverWorld() {
	defer func() {
		_ = recover()
	}()
	fmt.Println("world")
}
