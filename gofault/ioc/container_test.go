package ioc

import (
	"context"
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
	return &dependentService{dep: dep}
}

func TestContainer_RegisterAndResolve(t *testing.T) {
	c := New()
	err := c.Register(newDummyService)
	if err != nil {