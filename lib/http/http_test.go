package http

import (
	"testing"
)

func TestProxy(t *testing.T) {
	t.Skip()
	err := SetSocksProxy("127.0.0.1:1081")
	if err != nil {
		t.Errorf("set proxy err: %v", err)
	}
	resp, err := GetGlobalHttpClient().Get("http://127.0.0.1:8080/")
	if err != nil {
		t.Errorf("request failed, %v", err)
	}
	println(resp.Status)
}
