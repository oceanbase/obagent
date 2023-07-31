package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/config/mgragent"
	http2 "github.com/oceanbase/obagent/lib/http"
)

func Test_NewServer(t *testing.T) {
	Convey("time api", t, func() {
		server := NewServer(config.AgentVersion, mgragent.ServerConfig{})
		handler := func(w http.ResponseWriter, r *http.Request) {
			server.Router.ServeHTTP(w, r)
		}
		req := httptest.NewRequest("GET", "http://127.0.0.1:62888/api/v1/time", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		var successResponse http2.OcpAgentResponse
		_ = json.Unmarshal(body, &successResponse)

		So(resp.StatusCode, ShouldEqual, http.StatusOK)
		So(successResponse.Status, ShouldEqual, http.StatusOK)
		So(successResponse.Successful, ShouldEqual, true)
	})
}
