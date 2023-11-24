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
	return &dependentService{dep: dep}
}

func TestContainer_RegisterAndResolve(t *testing.T) {
	c := New()
	err := c.Register(newDummyService)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	svc, err := c.Resolve(&dummyService{})
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if svc.(*dummyService).msg != "hello" {
		t.Fatalf("expected 'hello', got %s", svc.(*dummyService).msg)
	}
}

func TestContainer_SingletonScope(t *testing.T) {
	c := New()
	c.Register(newDummyService)
	a, _ := c.Resolve(&dummyService{})