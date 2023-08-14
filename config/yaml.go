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
	"io"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type ExpandFunc func(string) string

func Decode(r io.Reader, defaultConfig interface{}) error {
	return yaml.NewDecoder(r).Decode(defaultConfig)
}

func ToStructured(src interface{}, dst interface{}, replacer ExpandFunc) (interface{}, error) {
	node := yaml.Node{}
	err := node.Encode(src)
	if err != nil {
		return nil, err
	}
	if replacer != nil {
		replaceValues(&node, replacer)
	}

	v := reflect.ValueOf(dst)
	kind := v.Kind()
	if kind != reflect.Ptr {
		v1 := reflect.New(v.Type())
		v1.Elem().Set(v)
		err = node.Decode(v1.Interface())
		if err != nil {
			return nil, err
		}
		return v1.Elem().Interface(), nil
	} else {
		err = node.Decode(dst)
		if err != nil {
			return nil, err
		}
		return dst, nil
	}
}

func ReplacedPath(i interface{}, replacer ExpandFunc) ([][]string, error) {
	node := yaml.Node{}
	err := node.Encode(i)
	if err != nil {
		return nil, err
	}
	if replacer == nil {
		return nil, errors.Errorf("replacer is nil")
	}
	return replaceValues(&node, replacer), nil
}

func recursivelyReplaceValues(node *yaml.Node, replacer ExpandFunc, p []string, ret *[][]string) {
	switch node.Kind {
	case yaml.MappingNode:
		for i, n := range node.Content {
			if i%2 == 1 {
				recursivelyReplaceValues(n, replacer, append(p, node.Content[i-1].Value), ret)
			}
		}
	case yaml.SequenceNode:
		for i, n := range node.Content {
			recursivelyReplaceValues(n, replacer, append(p, strconv.Itoa(i)), ret)
		}
	case yaml.ScalarNode:
		oldValue := node.Value
		node.Value = replacer(node.Value)
		node.Tag = ""
		if oldValue != node.Value {
			*ret = append(*ret, append(p, oldValue))
		}
	default:
	}
}

func replaceValues(node *yaml.Node, replacer ExpandFunc) [][]string {
	var p [][]string
	recursivelyReplaceValues(node, replacer, []string{}, &p)
	return p
}

func ReplaceConfValues(conf interface{}, context map[string]string) (interface{}, error) {
	expander := NewExpanderWithKeyValues(
		DefaultExpanderPrefix,
		DefaultExpanderSuffix,
		context,
	)
	return ToStructured(conf, conf, expander.Replace)
}
