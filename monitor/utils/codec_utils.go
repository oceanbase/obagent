package utils

import (
	"gopkg.in/yaml.v3"
)

// DecodeYaml decode yaml
func DecodeYaml(yamlStr string) (map[string]interface{}, error) {
	var resultMap map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlStr), &resultMap)
	return resultMap, err
}

// EncodeYaml encode yaml
func EncodeYaml(yamlMap map[string]interface{}) (string, error) {
	result, err := yaml.Marshal(yamlMap)
	return string(result), err
}
