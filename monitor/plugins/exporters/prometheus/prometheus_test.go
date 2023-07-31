package prometheus

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrometheus(t *testing.T) {
	config := `{"formatType":"fmtText"}`
	sourceConfig := make(map[string]interface{})
	err := json.Unmarshal([]byte(config), &sourceConfig)
	require.Equal(t, nil, err)
	p := &Prometheus{}
	err = p.Init(context.Background(), sourceConfig)
	require.Equal(t, nil, err)
}
