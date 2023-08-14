/*
 * Copyright (c) 2023 OceanBase
 * OBAgent is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package net

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func Helloworld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("helloworld"))
}

func TestConnectable(t *testing.T) {
	configStr := `
        timeout: 10s
        targets:
          t1: '127.0.0.1:9877'
    `
	var connectivityConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &connectivityConfigMap)

	connectivityInput := &ConnectivityInput{}
	connectivityInput.Init(context.Background(), connectivityConfigMap)

	http.HandleFunc("/", Helloworld)
	go http.ListenAndServe(":9877", nil)
	time.Sleep(time.Second * 1)

	metrics, _ := connectivityInput.CollectMsgs()
	require.Equal(t, 1, len(metrics))
	value, exists := metrics[0].GetField("value")
	v, ok := value.(float64)
	require.True(t, exists)
	require.True(t, ok)
	require.Equal(t, 1.0, v)

}

func TestNotConnectable(t *testing.T) {
	configStr := `
        timeout: 10s
        targets:
          t1: '127.0.0.1:28833'
    `
	var connectivityConfigMap map[string]interface{}
	_ = yaml.Unmarshal([]byte(configStr), &connectivityConfigMap)

	connectivityInput := &ConnectivityInput{}
	connectivityInput.Init(context.Background(), connectivityConfigMap)

	metrics, _ := connectivityInput.CollectMsgs()
	require.Equal(t, 1, len(metrics))
	value, exists := metrics[0].GetField("value")
	v, ok := value.(float64)
	require.True(t, exists)
	require.True(t, ok)
	require.Equal(t, 0.0, v)
}
