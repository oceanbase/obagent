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

package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// ConvertToFloat64 convert to float64
func ConvertToFloat64(v interface{}) (float64, bool) {
	var result float64
	convertOk := false
	switch v.(type) {
	case int:
		f, ok := v.(int)
		if ok {
			result = float64(f)
			convertOk = true
		}
	case int64:
		f, ok := v.(int64)
		if ok {
			result = float64(f)
			convertOk = true
		}
	case float64:
		f, ok := v.(float64)
		if ok {
			result = f
			convertOk = true
		}
	case string:
		s, ok := v.(string)
		if ok {
			f, err := strconv.ParseFloat(s, 64)
			if err == nil {
				result = f
				convertOk = true
			}
		}
	case []byte:
		bt, ok := v.([]byte)
		if ok {
			s := string(bt)
			f, err := strconv.ParseFloat(s, 64)
			if err == nil {
				result = f
				convertOk = true
			}
		}
	}
	return result, convertOk
}

// ConvertToBool convert to bool
func ConvertToBool(v interface{}) (bool, bool) {
	var result bool
	convertOk := false
	switch v.(type) {
	case int:
		i, ok := v.(int)
		if ok {
			result = (i > 0)
			convertOk = true
		}
	case int64:
		i, ok := v.(int64)
		if ok {
			result = (i > 0)
			convertOk = true
		}
	case float64:
		f, ok := v.(float64)
		if ok {
			result = (f > 0)
			convertOk = true
		}
	case bool:
		b, ok := v.(bool)
		if ok {
			result = b
			convertOk = true
		}
	case string:
		s, ok := v.(string)
		if ok {
			b, err := strconv.ParseBool(s)
			if err == nil {
				result = b
				convertOk = true
			}
		}
	case []byte:
		bt, ok := v.([]byte)
		if ok {
			s := string(bt)
			b, err := strconv.ParseBool(s)
			if err == nil {
				result = b
				convertOk = true
			}
		}
	}
	return result, convertOk
}

// ConvertToString convert to string
func ConvertToString(v interface{}) (string, bool) {
	result, ok := v.([]byte)
	if ok {
		return string(result), true
	} else {
		return fmt.Sprintf("%v", v), true
	}
}

func ConvertToLower(m map[string]string) {
	for k, v := range m {
		newK := strings.ToLower(k)
		newV := strings.ToLower(v)
		delete(m, k)
		m[newK] = newV
	}
}
