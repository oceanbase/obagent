package trace

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"

	agentlog "github.com/oceanbase/obagent/log"
)

const (
	// trace id
	TraceIdHeader = "X-OCP-Trace-ID"
	// ip
	OcpServerIpHeader = "X-OCP-Server-IP"
)

func GetTraceId(request *http.Request) string {
	// If no traceId passed, generate one.
	traceId := request.Header.Get(TraceIdHeader)
	if traceId == "" {
		traceId = RandomTraceId()
	}
	return traceId
}

func RandomTraceId() string {
	n := 8
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}

func ContextWithRandomTraceId() context.Context {
	return context.WithValue(context.Background(), agentlog.TraceIdKey{}, RandomTraceId())
}

func ContextWithTraceId(req *http.Request) context.Context {
	return context.WithValue(context.Background(), agentlog.TraceIdKey{}, GetTraceId(req))
}
