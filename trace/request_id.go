package trace

import (
	"context"
	"net/http"
)

const (
	RequestIdHeader = "X-Vcap-Request-Id"
)

func ContextWithRequestId(req *http.Request) context.Context {
	return context.WithValue(req.Context(), RequestIdHeader, req.Header.Get(RequestIdHeader))
}

func RequestIdFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(RequestIdHeader).(string); ok {
		return val
	}

	return ""
}
