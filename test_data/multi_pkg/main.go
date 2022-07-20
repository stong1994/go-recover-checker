package main

import (
	"github.com/stong1994/go-recover-checker/test_data/multi_pkg/controller"
	"github.com/stong1994/go-recover-checker/test_data/service"
)

func main() {
	c := controller.Control{
		S: service.Service{},
	}
	c.Hello()
}
