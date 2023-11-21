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
