package main

import "github.com/kardianos/service"

type program struct{}

func (*program) Start(s service.Service) error {
	panic("unimplemented")
}

func (*program) Stop(s service.Service) error {
	panic("unimplemented")
}

var _ service.Interface = (*program)(nil)

var srv service.Service

func init() {
	var err error
	srv, err = service.New(&program{}, &service.Config{
		Name:        "XX Service",
		DisplayName: "XX Service",
		Description: "Service for XX",
	})

	if err != nil {
		panic(err)
	}
}
