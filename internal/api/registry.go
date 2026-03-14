package api

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"sync"

	"github.com/ricky/oc-companion/internal/protocol"
)

type Handler func(context.Context, json.RawMessage) (interface{}, *protocol.Error)

type Method struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Usage       string                 `json:"usage"`
	Params      map[string]interface{} `json:"params,omitempty"`
	Handler     Handler                `json:"-"`
}

type Registry struct {
	mu      sync.RWMutex
	methods map[string]Method
}

func NewRegistry() *Registry {
	registry := &Registry{methods: make(map[string]Method)}
	registry.registerSystemMethods()

	return registry
}

func (r *Registry) Register(method Method) error {
	if method.Name == "" {
		return errors.New("method name is required")
	}

	if method.Handler == nil {
		return errors.New("method handler is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.methods[method.Name]; exists {
		return errors.New("method already registered")
	}

	r.methods[method.Name] = method

	return nil
}

func (r *Registry) Dispatch(ctx context.Context, request protocol.Request) protocol.Response {
	if request.Method == "" {
		return protocol.Failure(request.ID, protocol.CodeInvalidRequest, "method is required", nil)
	}

	r.mu.RLock()
	method, exists := r.methods[request.Method]
	r.mu.RUnlock()
	if !exists {
		return protocol.Failure(request.ID, protocol.CodeMethodNotFound, "method not found", map[string]string{"method": request.Method})
	}

	result, callErr := method.Handler(ctx, request.Params)
	if callErr != nil {
		return protocol.Failure(request.ID, callErr.Code, callErr.Message, callErr.Details)
	}

	return protocol.Success(request.ID, result)
}

func (r *Registry) Discover() []Method {
	r.mu.RLock()
	defer r.mu.RUnlock()

	methods := make([]Method, 0, len(r.methods))
	for _, method := range r.methods {
		methods = append(methods, Method{
			Name:        method.Name,
			Description: method.Description,
			Usage:       method.Usage,
			Params:      method.Params,
		})
	}

	sort.Slice(methods, func(i int, j int) bool {
		return methods[i].Name < methods[j].Name
	})

	return methods
}

func (r *Registry) registerSystemMethods() {
	_ = r.Register(Method{
		Name:        "system.ping",
		Description: "Returns service liveness status.",
		Usage:       `{"id":"1","method":"system.ping"}`,
		Handler: func(context.Context, json.RawMessage) (interface{}, *protocol.Error) {
			return map[string]string{"status": "ok"}, nil
		},
	})

	_ = r.Register(Method{
		Name:        "system.discover",
		Description: "Returns available tools and usage metadata.",
		Usage:       `{"id":"1","method":"system.discover"}`,
		Handler: func(_ context.Context, _ json.RawMessage) (interface{}, *protocol.Error) {
			return map[string]interface{}{
				"service": "oc-companion",
				"protocol": map[string]string{
					"transport": "unix-domain-socket",
					"encoding":  "json",
					"framing":   "json-stream",
				},
				"methods": r.Discover(),
			}, nil
		},
	})
}
