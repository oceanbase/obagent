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

package mgragent

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/oceanbase/obagent/tests/testutil"
)

func TestValidateYaml(t *testing.T) {
	m := &Manager{}
	err := m.validateYaml(`
test:
  - a: 1
  - a: 2
`)
	if err != nil {
		t.Errorf("should success")
	}

	err = m.validateYaml(`
test~
xxxxx
12345$
`)
	if err == nil {
		t.Errorf("should fail")
	}

	err = m.validateYaml(``)
	if err != nil {
		t.Errorf("should success")
	}
}

func TestNewChangeModuleConfigs(t *testing.T) {
	testutil.MakeDirs()
	defer testutil.DelTestFiles()

	m := NewManager(ManagerConfig{
		ModuleConfigDir: testutil.ModuleConfigDir,
	})
	ctx := context.Background()
	req := &ModuleConfigChangeRequest{
		ModuleConfigChanges: []ModuleConfigChange{
			{
				Operation: ModuleConfigSet,
				FileName:  "new_file.yaml",
				Content: `
test:
    k1: v1
    k2: v2
`,
			},
		},
	}
	changed, err := m.ChangeModuleConfigs(ctx, req)
	if err != nil {
		t.Error(err)
	}
	if len(changed) != 1 && changed[0] != "new_file.yaml" {
		t.Errorf("changed files wrong")
	}
	content, err := ioutil.ReadFile(filepath.Join(m.config.ModuleConfigDir, "new_file.yaml"))
	if err != nil {
		t.Error(err)
	}
	if string(content) != req.ModuleConfigChanges[0].Content {
		t.Errorf("content wrong")
	}
	content, err = ioutil.ReadFile(filepath.Join(m.config.ModuleConfigDir, "new_file.yaml.bak"))
	if err != nil {
		t.Error(err)
	}
	if string(content) != "" {
		t.Errorf("content wrong")
	}

	changed, err = m.ChangeModuleConfigs(ctx, &ModuleConfigChangeRequest{
		ModuleConfigChanges: []ModuleConfigChange{
			{
				Operation: ModuleConfigRestore,
				FileName:  "new_file.yaml",
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
	if testutil.FileExists(filepath.Join(m.config.ModuleConfigDir, "new_file.yaml")) ||
		testutil.FileExists(filepath.Join(m.config.ModuleConfigDir, "new_file.yaml.bak")) {
		t.Errorf("restore should del files")
	}
}

func TestSetChangeModuleConfigs(t *testing.T) {
	testutil.MakeDirs()
	defer testutil.DelTestFiles()

	m := NewManager(ManagerConfig{
		ModuleConfigDir: testutil.ModuleConfigDir,
	})
	orig := `
test:
    k1: v0
    k2: v0
`
	_ = ioutil.WriteFile(filepath.Join(m.config.ModuleConfigDir, "test.yaml"), []byte(orig), 0644)

	ctx := context.Background()
	req := &ModuleConfigChangeRequest{
		ModuleConfigChanges: []ModuleConfigChange{
			{
				Operation: ModuleConfigSet,
				FileName:  "test.yaml",
				Content: `
test:
    k1: v1
    k2: v2
`,
			},
		},
	}
	changed, err := m.ChangeModuleConfigs(ctx, req)
	if err != nil {
		t.Error(err)
	}
	if len(changed) != 1 && changed[0] != "test.yaml" {
		t.Errorf("changed files wrong")
	}
	content, err := ioutil.ReadFile(filepath.Join(m.config.ModuleConfigDir, "test.yaml"))
	if err != nil {
		t.Error(err)
	}
	if string(content) != req.ModuleConfigChanges[0].Content {
		t.Errorf("content wrong")
	}
	content, err = ioutil.ReadFile(filepath.Join(m.config.ModuleConfigDir, "test.yaml.bak"))
	if err != nil {
		t.Error(err)
	}
	if string(content) != orig {
		t.Errorf("content wrong")
	}

	changed, err = m.ChangeModuleConfigs(ctx, &ModuleConfigChangeRequest{
		ModuleConfigChanges: []ModuleConfigChange{
			{
				Operation: ModuleConfigRestore,
				FileName:  "test.yaml",
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
	if len(changed) != 1 && changed[0] != "test.yaml" {
		t.Errorf("changed files wrong")
	}
	content, err = ioutil.ReadFile(filepath.Join(m.config.ModuleConfigDir, "test.yaml"))
	if err != nil {
		t.Error(err)
	}
	if string(content) != orig {
		t.Errorf("content wrong")
	}
	if testutil.FileExists(filepath.Join(m.config.ModuleConfigDir, "new_file.yaml.bak")) {
		t.Errorf("restore should del bak file")
	}
}

func TestDelChangeModuleConfigs(t *testing.T) {
	testutil.MakeDirs()
	defer testutil.DelTestFiles()

	m := NewManager(ManagerConfig{
		ModuleConfigDir: testutil.ModuleConfigDir,
	})
	orig := `
test:
    k1: v0
    k2: v0
`
	_ = ioutil.WriteFile(filepath.Join(m.config.ModuleConfigDir, "test.yaml"), []byte(orig), 0644)

	ctx := context.Background()
	req := &ModuleConfigChangeRequest{
		ModuleConfigChanges: []ModuleConfigChange{
			{
				Operation: ModuleConfigDel,
				FileName:  "test.yaml",
				Content:   "",
			},
		},
	}
	changed, err := m.ChangeModuleConfigs(ctx, req)
	if err != nil {
		t.Error(err)
	}
	if len(changed) != 1 && changed[0] != "test.yaml" {
		t.Errorf("changed files wrong")
	}
	if testutil.FileExists(filepath.Join(m.config.ModuleConfigDir, "new_file.yaml.bak")) {
		t.Errorf("file should be deleted")
	}

	content, err := ioutil.ReadFile(filepath.Join(m.config.ModuleConfigDir, "test.yaml.bak"))
	if err != nil {
		t.Error(err)
	}
	if string(content) != orig {
		t.Errorf("content wrong")
	}

	changed, err = m.ChangeModuleConfigs(ctx, &ModuleConfigChangeRequest{
		ModuleConfigChanges: []ModuleConfigChange{
			{
				Operation: ModuleConfigRestore,
				FileName:  "test.yaml",
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
	if len(changed) != 1 && changed[0] != "test.yaml" {
		t.Errorf("changed files wrong")
	}
	content, err = ioutil.ReadFile(filepath.Join(m.config.ModuleConfigDir, "test.yaml"))
	if err != nil {
		t.Error(err)
	}
	if string(content) != orig {
		t.Errorf("content wrong")
	}
	if testutil.FileExists(filepath.Join(m.config.ModuleConfigDir, "new_file.yaml.bak")) {
		t.Errorf("restore should del bak file")
	}
}
