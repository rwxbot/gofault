package ioc

import (
	"context"
	"sync"
	"testing"
)

type dummyService struct {
	msg string
}

func newDummyService() *dummyService {
	return &dummyService{msg: "hello"}
}

func (s *dummyService) SomeMethod() string {
	return s.msg
}

type dependentService struct {
	dep *dummyService
}

func newDependentService(dep *dummyService) *dependentService {