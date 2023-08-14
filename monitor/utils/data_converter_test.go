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

package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertToFloat64FromInt(t *testing.T) {
	var i int
	i = 1
	result, ok := ConvertToFloat64(i)
	require.True(t, ok)
	require.Equal(t, 1.0, result)
}

func TestConvertToFloat64FromInt64(t *testing.T) {
	var i int64
	i = 1
	result, ok := ConvertToFloat64(i)
	require.True(t, ok)
	require.Equal(t, 1.0, result)
}

func TestConvertToFloat64FromFloat64(t *testing.T) {
	var f float64
	f = 1.0
	result, ok := ConvertToFloat64(f)
	require.True(t, ok)
	require.Equal(t, 1.0, result)
}

func TestConvertToFloat64FromStringSuccess(t *testing.T) {
	var s string
	s = "1.0"
	result, ok := ConvertToFloat64(s)
	require.True(t, ok)
	require.Equal(t, 1.0, result)
}

func TestConvertToFloat64FromStringFail(t *testing.T) {
	var s string
	s = "1.0.0"
	result, ok := ConvertToFloat64(s)
	require.True(t, !ok)
	require.Equal(t, 0.0, result)
}

func TestConvertToFloat64FromBytesSuccess(t *testing.T) {
	var s []byte
	s = []byte("1.0")
	result, ok := ConvertToFloat64(s)
	require.True(t, ok)
	require.Equal(t, 1.0, result)
}

func TestConvertToFloat64FromBytesFail(t *testing.T) {
	var s []byte
	s = []byte("1.0.0")
	result, ok := ConvertToFloat64(s)
	require.True(t, !ok)
	require.Equal(t, 0.0, result)
}

func TestConvertToBoolFromIntTrue(t *testing.T) {
	var i int
	i = 1
	result, ok := ConvertToBool(i)
	require.True(t, ok)
	require.True(t, result)
}

func TestConvertToBoolFromIntFalse(t *testing.T) {
	var i int
	i = -1
	result, ok := ConvertToBool(i)
	require.True(t, ok)
	require.True(t, !result)
}

func TestConvertToBoolFromInt64True(t *testing.T) {
	var i int64
	i = 1
	result, ok := ConvertToBool(i)
	require.True(t, ok)
	require.True(t, result)
}

func TestConvertToBoolFromInt64False(t *testing.T) {
	var i int64
	i = -1
	result, ok := ConvertToBool(i)
	require.True(t, ok)
	require.True(t, !result)
}

func TestConvertToBoolFromFloat64True(t *testing.T) {
	var f float64
	f = 1.0
	result, ok := ConvertToBool(f)
	require.True(t, ok)
	require.True(t, result)
}

func TestConvertToBoolFromFloat64False(t *testing.T) {
	var f float64
	f = -1.0
	result, ok := ConvertToBool(f)
	require.True(t, ok)
	require.True(t, !result)
}

func TestConvertToBoolFromBoolTrue(t *testing.T) {
	var b bool
	b = true
	result, ok := ConvertToBool(b)
	require.True(t, ok)
	require.True(t, result)
}

func TestConvertToBoolFromBoolFalse(t *testing.T) {
	var b bool
	b = false
	result, ok := ConvertToBool(b)
	require.True(t, ok)
	require.True(t, !result)
}

func TestConvertToBoolFromStringTrue(t *testing.T) {
	var s string
	s = "true"
	result, ok := ConvertToBool(s)
	require.True(t, ok)
	require.True(t, result)
}

func TestConvertToBoolFromStringFalse(t *testing.T) {
	var s string
	s = "false"
	result, ok := ConvertToBool(s)
	require.True(t, ok)
	require.True(t, !result)
}

func TestConvertToBoolFromStringFail(t *testing.T) {
	var s string
	s = "1.0.0"
	result, ok := ConvertToBool(s)
	require.True(t, !ok)
	require.True(t, !result)
}

func TestConvertToBoolFromBytesTrue(t *testing.T) {
	var s []byte
	s = []byte("t")
	result, ok := ConvertToBool(s)
	require.True(t, ok)
	require.True(t, result)
}

func TestConvertToBoolFromBytesFalse(t *testing.T) {
	var s []byte
	s = []byte("f")
	result, ok := ConvertToBool(s)
	require.True(t, ok)
	require.True(t, !result)
}

func TestConvertToBoolFromBytesFail(t *testing.T) {
	var s []byte
	s = []byte("1.0.0")
	result, ok := ConvertToBool(s)
	require.True(t, !ok)
	require.True(t, !result)
}

func TestConvertToStringFromBytes(t *testing.T) {
	var s []byte
	s = []byte("1.0.0")
	result, ok := ConvertToString(s)
	require.True(t, ok)
	require.Equal(t, "1.0.0", result)
}

func TestConvertToStringFromOther(t *testing.T) {
	var i int
	i = 1
	result, ok := ConvertToString(i)
	require.True(t, ok)
	require.Equal(t, "1", result)
}
