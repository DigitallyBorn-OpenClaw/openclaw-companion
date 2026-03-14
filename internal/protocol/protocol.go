package protocol

import "encoding/json"

const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32000
)

type Request struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	ID     string      `json:"id,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Error  *Error      `json:"error,omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func Success(id string, result interface{}) Response {
	return Response{ID: id, Result: result}
}

func Failure(id string, code int, message string, details interface{}) Response {
	return Response{
		ID: id,
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
