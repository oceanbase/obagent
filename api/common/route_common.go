package common

import (
	"net/http"

	"github.com/felixge/fgprof"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
)

func InitPprofRouter(r *gin.Engine) {
	pprof.Register(r, "debug/pprof")
	r.GET("/debug/fgprof", adapter.Wrap(func(_ http.Handler) http.Handler {
		return fgprof.Handler()
	}))
}
