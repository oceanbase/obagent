package common

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/http"
	"github.com/oceanbase/obagent/lib/system"
	"github.com/oceanbase/obagent/lib/trace"
	"github.com/oceanbase/obagent/stat"
)

var libSystem system.System = system.SystemImpl{}

const statusURI = "/api/v1/status"

const logQuerierURI = "/api/v1/log"

// Before handlers, extract HTTP headers, and log the API request.
func PreHandlers(maskBodyRoutes ...string) func(*gin.Context) {
	return func(c *gin.Context) {
		if c.Request.RequestURI == statusURI {
			c.Next()
			return
		}
		// Use traceId passed from OCP-Server for logging.
		traceId := trace.GetTraceId(c.Request)
		c.Set(TraceIdKey, traceId)

		// Store OCP-Server's ip address for logging.
		// c.ClientIP() may not be accurate if HTTP requests are forwarded by proxy server.
		ocpServerIp := c.Request.Header.Get(trace.OcpServerIpHeader)
		c.Set(OcpServerIpKey, ocpServerIp)

		ctx := NewContextWithTraceId(c)

		masked := false
		for _, it := range maskBodyRoutes {
			if strings.HasPrefix(c.Request.RequestURI, it) {
				masked = true
			}
		}
		if masked {
			log.WithContext(ctx).Infof("API request: [%v %v, client=%v, ocpServerIp=%v, traceId=%v]",
				c.Request.Method, c.Request.URL, c.ClientIP(), ocpServerIp, traceId)
		} else {
			body := readRequestBody(c)
			log.WithContext(ctx).Infof("API request: [%v %v, client=%v, ocpServerIp=%v, traceId=%v, body=%v]",
				c.Request.Method, c.Request.URL, c.ClientIP(), ocpServerIp, traceId, body)
		}

		c.Next()
	}
}

var emptyRe = regexp.MustCompile(`\s+`)

func readRequestBody(c *gin.Context) string {
	body, _ := ioutil.ReadAll(c.Request.Body)
	c.Request.Body = ioutil.NopCloser(bytes.NewReader(body))
	return emptyRe.ReplaceAllString(string(body), "")
}

// If c.BindJSON fails (e.g. validation error), content-type will be set to text/plain.
// So set content-type before handlers.
func SetContentType(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")

	c.Next()
}

func getResponseFromContext(c *gin.Context) http.OcpAgentResponse {
	ctx := NewContextWithTraceId(c)

	if len(c.Errors) > 0 {
		var subErrors []interface{}
		for _, e := range c.Errors {
			switch e.Type {
			case gin.ErrorTypeBind:
				validationErrors := e.Err.(validator.ValidationErrors)
				for _, fieldError := range validationErrors {
					subErrors = append(subErrors, http.NewApiFieldError(fieldError))
				}
			default:
				subErrors = append(subErrors, http.ApiUnknownError{Error: e.Err})
			}
		}
		return http.NewSubErrorsResponse(subErrors)
	}

	if r, ok := c.Get(OcpAgentResponseKey); ok {
		if resp, ok := r.(http.OcpAgentResponse); ok {
			return resp
		}
	}

	log.WithContext(ctx).Error("no response object found from gin context")
	return http.NewErrorResponse(errors.Occur(errors.ErrUnexpected, "cannot build response body"))
}

// After handlers, build the complete OcpAgentResponse object,
// log the API result, and send HTTP response.
func PostHandlers(excludeRoutes ...string) func(*gin.Context) {
	localIpAddress, _ := libSystem.GetLocalIpAddress()
	return func(c *gin.Context) {
		for _, it := range excludeRoutes {
			if strings.HasPrefix(c.Request.RequestURI, it) {
				c.Next()
				return
			}
		}

		startTime := time.Now()

		c.Next()

		ctx := NewContextWithTraceId(c)
		resp := getResponseFromContext(c)

		duration := time.Now().Sub(startTime)
		resp.Duration = int(duration / time.Millisecond)

		ocpServerIp, _ := c.Get(OcpServerIpKey)
		if v, ok := c.Get(TraceIdKey); ok {
			if traceId, ok := v.(string); ok {
				resp.TraceId = traceId
			}
		}

		resp.Server = localIpAddress
		if resp.Successful {
			if c.Request.RequestURI != statusURI {
				if strings.HasPrefix(c.Request.RequestURI, logQuerierURI) {
					log.WithContext(ctx).Infof("API response OK: [%v %v, client=%v, ocpServerIp=%v, traceId=%v, duration=%v, status=%v]",
						c.Request.Method, c.Request.URL, c.ClientIP(), ocpServerIp, resp.TraceId, duration, resp.Status)
				} else {
					log.WithContext(ctx).Infof("API response OK: [%v %v, client=%v, ocpServerIp=%v, traceId=%v, duration=%v, status=%v, data=%+v]",
						c.Request.Method, c.Request.URL, c.ClientIP(), ocpServerIp, resp.TraceId, duration, resp.Status, resp.Data)
				}
			} else {
				log.WithContext(ctx).Debugf("API response OK: [%v %v, client=%v, ocpServerIp=%v, traceId=%v, duration=%v, status=%v, data=%+v]",
					c.Request.Method, c.Request.URL, c.ClientIP(), ocpServerIp, resp.TraceId, duration, resp.Status, resp.Data)
			}
		} else {
			log.WithContext(ctx).Infof("API response error: [%v %v, client=%v, ocpServerIp=%v, traceId=%v, duration=%v, status=%v, error=%v]",
				c.Request.Method, c.Request.URL, c.ClientIP(), ocpServerIp, resp.TraceId, duration, resp.Status, resp.Error.String())
		}
		c.JSON(resp.Status, resp)
	}
}

func MonitorAgentPostHandler(c *gin.Context) {
	startTime := time.Now()

	c.Next()

	duration := time.Now().Sub(startTime)
	serverIp := c.Request.Header.Get(trace.OcpServerIpHeader)
	ctx := NewContextWithTraceId(c)

	fields := log.Fields{
		"url":         c.Request.URL,
		"duration":    duration,
		"status":      c.Writer.Status(),
		"ocpServerIp": serverIp,
		"client":      c.ClientIP(),
	}
	if duration < 100*time.Millisecond {
		fields[logrus.FieldKeyLevel] = logrus.DebugLevel
	}

	log.WithContext(ctx).WithFields(fields).Info("request end")
}

func HttpStatMiddleware(c *gin.Context) {
	startTime := time.Now()

	// run other middleware
	c.Next()

	stat.HttpRequestMillisecondsSummary.With(prom.Labels{
		stat.HttpMethod:  c.Request.Method,
		stat.HttpStatus:  fmt.Sprintf("%d", c.Writer.Status()),
		stat.HttpApiPath: c.Request.URL.Path,
	}).Observe(float64(time.Now().Sub(startTime)) / float64(time.Millisecond))
}

func Recovery(c *gin.Context, err interface{}) {
	// log err to file
	log.WithContext(NewContextWithTraceId(c)).Errorf("request context %+v, err:%+v", c, err)
}

func IgnoreFaviconHandler(c *gin.Context) {
	if c.Request.URL.Path == "/favicon.ico" {
		c.Abort()
	}
}
