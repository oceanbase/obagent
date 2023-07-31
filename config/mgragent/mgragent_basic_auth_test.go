package mgragent

import (
	"bytes"
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/oceanbase/obagent/config"
)

var (
	basicAuthConfigGroup = &config.ConfigPropertiesGroup{
		Configs: []*config.ConfigProperty{
			{
				Key:         "agent.http.basic.auth.username",
				Value:       "mgragent",
				ValueType:   config.ValueString,
				Description: "mgragent basic auth username",
				Unit:        "",
			},
			{
				Key:         "agent.http.basic.auth.password",
				Value:       "root@123",
				ValueType:   config.ValueString,
				Description: "mgragent basic auth password",
				Unit:        "",
			},
		},
	}

	basicAuthModuleConfig = config.ModuleConfig{
		Module:   "mgragent.basic.auth",
		Process:  "mgragent",
		Disabled: false,
		Config: &config.BasicAuthConfig{
			Auth:     "basic",
			Username: "agent.http.basic.auth.username",
			Password: "agent.http.basic.auth.password",
		},
	}
)

func TestBasicAuthConfigPropertyExample(t *testing.T) {
	w := bytes.NewBuffer(make([]byte, 0, 10))
	err := yaml.NewEncoder(w).Encode(basicAuthConfigGroup)
	if err != nil {
		t.Failed()
	}
	fmt.Printf("%s\n", w.Bytes())
}

func TestBasicAuthModuleConfigExample(t *testing.T) {
	w := bytes.NewBuffer(make([]byte, 0, 10))
	err := yaml.NewEncoder(w).Encode(basicAuthModuleConfig)
	if err != nil {
		t.Failed()
	}
	fmt.Printf("%s\n", w.Bytes())
}

func TestModuleConfigTemplates(t *testing.T) {
	templates := &config.ModuleConfigGroup{
		Modules: []config.ModuleConfig{basicAuthModuleConfig},
	}

	w := bytes.NewBuffer(make([]byte, 0, 10))
	err := yaml.NewEncoder(w).Encode(templates)
	if err != nil {
		t.Failed()
	}
	fmt.Printf("%s\n", w.Bytes())
}
