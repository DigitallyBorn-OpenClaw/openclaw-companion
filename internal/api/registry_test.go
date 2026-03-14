package api

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ricky/oc-companion/internal/protocol"
)

func TestDispatch_EchoesNumericID(t *testing.T) {
	registry := NewRegistry()
	request := protocol.Request{
		ID:     json.RawMessage(`123`),
		Method: "system.ping",
	}

	response := registry.Dispatch(context.Background(), request)
	if string(response.ID) != "123" {
		t.Fatalf("expected ID to be echoed, got %s", string(response.ID))
	}

	if response.Error != nil {
		t.Fatalf("expected no error, got %+v", response.Error)
	}
}

func TestDispatch_UnknownMethod(t *testing.T) {
	registry := NewRegistry()
	request := protocol.Request{ID: json.RawMessage(`"1"`), Method: "does.not.exist"}

	response := registry.Dispatch(context.Background(), request)
	if response.Error == nil {
		t.Fatalf("expected error for unknown method")
	}

	if response.Error.Code != protocol.CodeMethodNotFound {
		t.Fatalf("expected method not found code, got %d", response.Error.Code)
	}
}
