package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResponderKeepsHTTPStatusAndBusinessCodeIndependent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	err := NewBusinessError(CodeRegisterUserAlreadyExists, "user already exists", ErrConflict)
	NewResponder("test").Error(context, err)

	var response Response
	if decodeErr := json.Unmarshal(recorder.Body.Bytes(), &response); decodeErr != nil {
		t.Fatalf("decode response: %v", decodeErr)
	}
	if recorder.Code != http.StatusConflict {
		t.Fatalf("HTTP status = %d, want %d", recorder.Code, http.StatusConflict)
	}
	if response.Code != CodeRegisterUserAlreadyExists {
		t.Fatalf("response code = %d, want %d", response.Code, CodeRegisterUserAlreadyExists)
	}
	if response.Message != "resource already exists: user already exists" {
		t.Fatalf("message = %q, want %q", response.Message, "resource already exists: user already exists")
	}
}

func TestResponderPreservesKnownErrorContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	err := fmt.Errorf("%w: username is required", ErrInvalidInput)
	NewResponder("test").Error(context, err)

	var response Response
	if decodeErr := json.Unmarshal(recorder.Body.Bytes(), &response); decodeErr != nil {
		t.Fatalf("decode response: %v", decodeErr)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("HTTP status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if response.Message != "invalid input: username is required" {
		t.Fatalf("message = %q, want %q", response.Message, "invalid input: username is required")
	}
}
