package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestBaseController_JSON(t *testing.T) {
	ctrl := BaseController{}
	w := httptest.NewRecorder()
	err := ctrl.JSON(w, http.StatusOK, map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("JSON failed: %v", err)
	}
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["key"] != "value" {
		t.Fatalf("expected 'value', got %s", resp["key"])
	}
}
