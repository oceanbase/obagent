// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpander(t *testing.T) {
	expander := NewExpander("${", "}")
	expander.Add("a", "1")
	expander.AddAll(map[string]string{
		"b": "2",
		"c": "3",
	})
	expander.init()
	s := expander.Replace("${a} + ${b} = ${c}")
	if s != "1 + 2 = 3" {
		t.Error("expand wrong")
	}
}

func TestNilExpander(t *testing.T) {
	expander := (*Expander)(nil)
	result := expander.Replace("a")
	assert.Equal(t, "a", result)
}

func TestExpander_Copy(t *testing.T) {
	os.Setenv("EXPANDER_TEST_KEY", "EXPANDER_TEST_VALUE")
	expander := NewExpanderWithKeyValues(
		DefaultExpanderPrefix,
		DefaultExpanderSuffix,
		map[string]string{
			"1": "1",
		},
		EnvironKeyValues(),
	)
	assert.Equal(t, expander, expander.Copy())
	assert.Equal(t, expander.keyValues["EXPANDER_TEST_KEY"], "EXPANDER_TEST_VALUE")

	nilExpander := (*Expander)(nil)
	assert.Equal(t, nilExpander, nilExpander.Copy())
}

func TestExpander_TrimPrefixSuffix(t *testing.T) {
	expander := NewExpanderWithKeyValues(
		DefaultExpanderPrefix,
		DefaultExpanderSuffix,
		nil,
	)
	t.Run("trim with normal input", func(t *testing.T) {
		result := expander.TrimPrefixSuffix("${foo}", "${bar}")
		assert.Equal(t, []string{"foo", "bar"}, result)

		result2 := expander.TrimPrefixSuffix("${foo}")
		assert.Equal(t, []string{"foo"}, result2)
	})

	t.Run("trim with abnormal input", func(t *testing.T) {
		result := expander.TrimPrefixSuffix("${}", "${}")
		assert.Equal(t, []string{"", ""}, result)

		result2 := expander.TrimPrefixSuffix("${}")
		assert.Equal(t, []string{""}, result2)

		result3 := expander.TrimPrefixSuffix("${foo")
		assert.Equal(t, []string{"foo"}, result3)

		result4 := expander.TrimPrefixSuffix("foo}")
		assert.Equal(t, []string{"foo"}, result4)
	})
}
