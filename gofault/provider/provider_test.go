package provider

import (
	"testing"

	"github.com/gofault/gofault/ioc"
)

func TestProviderFunc(t *testing.T) {
	val := "not-called"
	pf := ProviderFunc(func() any {