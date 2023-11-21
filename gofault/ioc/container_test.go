package ioc

import (
	"context"
	"testing"
)

type dummyService struct {
	msg string
}

func newDummyService() *dummyService {