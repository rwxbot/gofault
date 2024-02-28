package server

import (
	"net/http"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	handler := http.NewServeMux()