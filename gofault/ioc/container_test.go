package ioc

import (
	"context"
	"sync"
	"testing"
)

type dummyService struct {
	msg string
}