/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

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
