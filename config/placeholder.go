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

package config

import (
	"os"
	"strings"
)

const (
	DefaultExpanderPrefix = "${"
	DefaultExpanderSuffix = "}"
)

type Expander struct {
	keyValues map[string]string
	replacer  *strings.Replacer
	prefix    string
	suffix    string
}

func (r *Expander) Add(key, value string) {
	r.keyValues[key] = value
	r.replacer = nil
}

func (r *Expander) AddAll(kv map[string]string) {
	for k, v := range kv {
		r.Add(k, v)
	}
}

func (r *Expander) Copy() *Expander {
	if r == nil {
		return nil
	}
	return &Expander{
		prefix:    r.prefix,
		suffix:    r.suffix,
		keyValues: r.keyValues,
		replacer:  r.replacer,
	}
}

func (r *Expander) init() {
	var pairs []string
	for k, v := range r.keyValues {
		pairs = append(pairs, r.prefix+k+r.suffix, v)
	}
	r.replacer = strings.NewReplacer(pairs...)
}

func (r *Expander) Replace(s string) string {
	if r == nil {
		return s
	}
	if r.replacer == nil {
		r.init()
	}
	return r.replacer.Replace(s)
}

func (r *Expander) TrimPrefixSuffix(keys ...string) []string {
	ret := make([]string, 0, len(keys))
	for _, key := range keys {
		ret = append(ret, strings.NewReplacer(r.prefix, "", r.suffix, "").Replace(key))
	}
	return ret
}

func NewExpander(prefix, suffix string) *Expander {
	return NewExpanderWithKeyValues(prefix, suffix, map[string]string{})
}

func NewExpanderWithKeyValues(prefix, suffix string, skvs ...map[string]string) *Expander {
	conkvs := make(map[string]string, 128)
	for _, kvs := range skvs {
		for k, v := range kvs {
			conkvs[k] = v
		}
	}

	return &Expander{
		prefix:    prefix,
		suffix:    suffix,
		keyValues: conkvs,
	}
}

func EnvironKeyValues() map[string]string {
	kvs := make(map[string]string, 64)
	envs := os.Environ()
	for _, pairs := range envs {
		kv := strings.SplitN(pairs, "=", 2)
		kvs[kv[0]] = kv[1]
	}
	return kvs
}
