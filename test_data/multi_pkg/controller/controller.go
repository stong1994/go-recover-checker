package controller

import "github.com/stong1994/go-recover-checker/test_data/service"

type Control struct {
	S service.Service
}

func (c Control) Hello() {
	c.S.Hello()
}
