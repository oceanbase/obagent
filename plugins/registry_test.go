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

package plugins

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterInput(t *testing.T) {
	InputManager := GetInputManager()
	delete(InputManager.Registry, "test")
	InputManager.Register("test", func() Input {
		return nil
	})
	_, found := InputManager.Registry["test"]
	require.True(t, found)
}

func TestRegisterProcessor(t *testing.T) {
	ProcessorManager := GetProcessorManager()
	delete(ProcessorManager.Registry, "test")
	ProcessorManager.Register("test", func() Processor {
		return nil
	})
	_, found := ProcessorManager.Registry["test"]
	require.True(t, found)
}

func TestRegisterOutput(t *testing.T) {
	OutputManager := GetOutputManager()
	delete(OutputManager.Registry, "test")
	OutputManager.Register("test", func() Output {
		return nil
	})
	_, found := OutputManager.Registry["test"]
	require.True(t, found)
}

func TestRegisterExporter(t *testing.T) {
	exporterManager := GetExporterManager()
	delete(exporterManager.Registry, "test")
	exporterManager.Register("test", func() Exporter {
		return nil
	})
	_, found := exporterManager.Registry["test"]
	require.True(t, found)
}

func TestRegisterInputAgain(t *testing.T) {
	defer func() { recover() }()
	InputManager := GetInputManager()
	delete(InputManager.Registry, "test")
	for i := 0; i < 2; i++ {
		InputManager.Register("test", func() Input {
			return nil
		})
	}
	t.Errorf("Register input should panic")
}

func TestRegisterProcessorAgain(t *testing.T) {
	defer func() { recover() }()
	ProcessorManager := GetProcessorManager()
	delete(ProcessorManager.Registry, "test")
	for i := 0; i < 2; i++ {
		ProcessorManager.Register("test", func() Processor {
			return nil
		})
	}
	t.Errorf("Register processor should panic")
}

func TestRegisterOutputAgain(t *testing.T) {
	defer func() { recover() }()
	OutputManager := GetOutputManager()
	delete(OutputManager.Registry, "test")
	for i := 0; i < 2; i++ {
		OutputManager.Register("test", func() Output {
			return nil
		})
	}
	t.Errorf("Register output should panic")
}

func TestRegisterExporterAgain(t *testing.T) {
	defer func() { recover() }()
	exporterManager := GetExporterManager()
	delete(exporterManager.Registry, "test")
	for i := 0; i < 2; i++ {
		exporterManager.Register("test", func() Exporter {
			return nil
		})
	}
	t.Errorf("Register exporter should panic")
}

func TestGetInput(t *testing.T) {
	InputManager := GetInputManager()
	delete(InputManager.Registry, "test")
	InputManager.Register("test", func() Input {
		return nil
	})
	_, err := InputManager.GetPlugin("test")
	require.True(t, err == nil)
}

func TestGetProcessor(t *testing.T) {
	ProcessorManager := GetProcessorManager()
	delete(ProcessorManager.Registry, "test")
	ProcessorManager.Register("test", func() Processor {
		return nil
	})
	_, err := ProcessorManager.GetPlugin("test")
	require.True(t, err == nil)
}

func TestGetOutput(t *testing.T) {
	OutputManager := GetOutputManager()
	delete(OutputManager.Registry, "test")
	OutputManager.Register("test", func() Output {
		return nil
	})
	_, err := OutputManager.GetPlugin("test")
	require.True(t, err == nil)
}

func TestGetExporter(t *testing.T) {
	exporterManager := GetExporterManager()
	delete(exporterManager.Registry, "test")
	exporterManager.Register("test", func() Exporter {
		return nil
	})
	_, err := exporterManager.GetPlugin("test")
	require.True(t, err == nil)
}

func TestGetInputNotExists(t *testing.T) {
	InputManager := GetInputManager()
	delete(InputManager.Registry, "test")
	_, err := InputManager.GetPlugin("test")
	require.True(t, err != nil)
}

func TestGetProcessorNotExists(t *testing.T) {
	ProcessorManager := GetProcessorManager()
	delete(ProcessorManager.Registry, "test")
	_, err := ProcessorManager.GetPlugin("test")
	require.True(t, err != nil)
}

func TestGetOutputNotExists(t *testing.T) {
	OutputManager := GetOutputManager()
	delete(OutputManager.Registry, "test")
	_, err := OutputManager.GetPlugin("test")
	require.True(t, err != nil)
}

func TestGetExporterNotExists(t *testing.T) {
	exporterManager := GetExporterManager()
	delete(exporterManager.Registry, "test")
	_, err := exporterManager.GetPlugin("test")
	require.True(t, err != nil)
}
