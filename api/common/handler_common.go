package common

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/obagent/lib/http"
	"github.com/oceanbase/obagent/log"
)

// keys stored in gin.Context
const (
	OcpAgentResponseKey = "ocpAgentResponse"
	TraceIdKey          = "traceId"
	OcpServerIpKey      = "ocpServerIp"
)

func NewContextWithTraceId(c *gin.Context) context.Context {
	traceId := ""
	if t, ok := c.Get(TraceIdKey); ok {
		if ts, ok := t.(string); ok {
			traceId = ts
		}
	}
	return context.WithValue(context.Background(), log.TraceIdKey{}, traceId)
}

func SendResponse(c *gin.Context, data interface{}, err error) {
	resp := http.BuildResponse(data, err)
	c.Set(OcpAgentResponseKey, resp)
}
