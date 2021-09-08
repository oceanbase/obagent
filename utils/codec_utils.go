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

package utils

import (
	"gopkg.in/yaml.v3"
)

//DecodeYaml decode yaml
func DecodeYaml(yamlStr string) (map[string]interface{}, error) {
	var resultMap map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlStr), &resultMap)
	return resultMap, err
}

//EncodeYaml encode yaml
func EncodeYaml(yamlMap map[string]interface{}) (string, error) {
	result, err := yaml.Marshal(yamlMap)
	return string(result), err
}
