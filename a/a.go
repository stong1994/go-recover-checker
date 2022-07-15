package data

import "fmt"

type Service struct{}

func (s Service) hi() {
	go s.ExecTask()
}

func (s Service) ExecTask() {
	for {
		if err := Do(func() {
			recoverWorld()
		}); err != nil {
		}
	}
}

func recoverWorld() {
	fmt.Println("world")
}
func Do(f func()) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	f()
	return err
}
