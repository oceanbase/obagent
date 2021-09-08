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
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/huandu/go-clone"
	"gopkg.in/yaml.v3"
)

var testYaml1 = `
a: aaa
b: bbb
c:
  x1: 111
  x2: 222
d: [ a b c ]
`

var testYaml2 = `
c:
  a: 333
  b: 444
  foo: 123
  foo2: '123'
  fooBar: true
  fooBar2: 'true'
`

type Test1 struct {
	A string
	B string
	C map[string]string
	D []string
}

type Test3 struct {
	A       string
	B       string
	Foo     int
	Foo2    int
	FooBar  bool
	FooBar2 bool
}

func TestParseStrut(t *testing.T) {
	s := Test1{}
	err := Decode(strings.NewReader(testYaml1), &s)
	if err != nil {
		t.Error(err.Error())
	}
	if s.A != "aaa" || s.B != "bbb" || s.C["x1"] != "111" || s.C["x2"] != "222" ||
		reflect.DeepEqual(s.D, []string{"a", "b", "c"}) {
		t.Error("parsed result wrong!")
	}
	err = Decode(strings.NewReader(testYaml2), &s)
	if err != nil {
		t.Error(err.Error())
	}
	if s.A != "aaa" || s.B != "bbb" ||
		s.C["x1"] != "111" || s.C["x2"] != "222" ||
		s.C["a"] != "333" || s.C["b"] != "444" ||
		reflect.DeepEqual(s.D, []string{"a", "b", "c"}) {
		t.Error("new value not appended")
	}
}

func TestToStructured(t *testing.T) {
	in := map[string]interface{}{
		"a":      "1",
		"b":      "2",
		"foo":    123,
		"foobar": true,
	}
	out := Test3{}
	var outIfc interface{} = &out
	out2, err := ToStructured(in, outIfc, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if out.A != "1" || out.B != "2" {
		t.Error("ToStructured wrong")
	}
	if out2 != &out {
		t.Error("point should be same ")
	}

	out3, err := ToStructured(in, out, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	test3 := out3.(Test3)
	fmt.Println(test3)
	if test3.A != "1" || test3.B != "2" || test3.Foo != 123 || !test3.FooBar {
		t.Error("ToStructured wrong")
	}
}

func TestWalk(t *testing.T) {
	node := yaml.Node{}
	in := map[string]interface{}{
		"str": "1",
		"list": []string{
			"abc",
		},
	}
	err := node.Encode(in)
	if err != nil {
		t.Error(err)
	}
	p := replaceValues(&node, func(s string) string {
		return s + "!"
	})
	var out map[string]interface{}
	node.Decode(&out)
	if out["str"] != "1!" {
		t.Error("bad")
	}
	if out["list"].([]interface{})[0] != "abc!" {
		t.Error("bad")
	}
	fmt.Printf("%+v\n", p)
}

func TestReplacedPath(t *testing.T) {
	type replaceArgs struct {
		object   interface{}
		replacer func(string) string
	}

	type bb struct {
		Bbb string `yaml:"b.b.b"`
	}
	type cc struct {
		Num map[int]int `yaml:"c.c.c"`
	}
	type dd struct {
		Ddd map[string]int `yaml:"d.d.d"`
	}
	type ff struct {
		fff string `yaml:"f.f.f"`
	}
	type fooStruct struct {
		Aa string `yaml:"a.a"`
		Bb bb     `yaml:"b.b"`
		Cc cc     `yaml:"c.c"`
		dd dd     `yaml:"d.d"`
		ee string `yaml:"e.e"`
		ff ff     `yaml:"f.f"`
	}

	kvs := map[string]string{
		"aa":  "a-a",
		"bbb": "b-b-b",
		"1":   "11",
		"2":   "22",
	}

	var (
		objectMap = map[string]interface{}{
			"a.a": "aa",
			"b.b": map[string]string{
				"b.b.b": "bbb",
			},
			"c.c": map[int]int{
				1: 1,
			},
			"d.d": map[string]int{
				"d.d.d": 2,
			},
			"e.e": "ee",
			"f.f": map[string]string{
				"f.f.f": "fff",
			},
		}

		foo = fooStruct{
			Aa: "aa",
			Bb: bb{
				Bbb: "bbb",
			},
			Cc: cc{
				Num: map[int]int{1: 1},
			},
			dd: dd{
				Ddd: map[string]int{"d.d.d": 2},
			},
			ee: "e.e",
			ff: ff{
				fff: "f.f.f",
			},
		}

		replacer = func(s string) string {
			if v, ex := kvs[s]; ex {
				return v
			}
			return s
		}
	)

	tests := []struct {
		name    string
		args    replaceArgs
		want    [][]string
		wantErr bool
	}{
		{
			name: "empty map replace path",
			args: replaceArgs{
				object:   map[string]string{},
				replacer: replacer,
			},
			want:    nil,
			wantErr: false,
		},

		{
			name: "map replace path with nil replacer",
			args: replaceArgs{
				object:   objectMap,
				replacer: nil,
			},
			want:    nil,
			wantErr: true,
		},

		{
			name: "map replace path",
			args: replaceArgs{
				object:   clone.Clone(objectMap),
				replacer: replacer,
			},
			want: [][]string{
				{"a.a", "aa"},
				{"b.b", "b.b.b", "bbb"},
				{"c.c", "1", "1"},
				{"d.d", "d.d.d", "2"},
			},
			wantErr: false,
		},

		{
			name: "*map replace path",
			args: replaceArgs{
				object:   &objectMap,
				replacer: replacer,
			},
			want: [][]string{
				{"a.a", "aa"},
				{"b.b", "b.b.b", "bbb"},
				{"c.c", "1", "1"},
				{"d.d", "d.d.d", "2"},
			},
			wantErr: false,
		},

		{
			name: "fooStruct replace path",
			args: replaceArgs{
				object:   clone.Clone(foo),
				replacer: replacer,
			},
			want: [][]string{
				{"a.a", "aa"},
				{"b.b", "b.b.b", "bbb"},
				{"c.c", "c.c.c", "1", "1"},
			},
			wantErr: false,
		},

		{
			name: "*fooStruct replace path",
			args: replaceArgs{
				object:   clone.Clone(&foo),
				replacer: replacer,
			},
			want: [][]string{
				{"a.a", "aa"},
				{"b.b", "b.b.b", "bbb"},
				{"c.c", "c.c.c", "1", "1"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplacedPath(tt.args.object, tt.args.replacer)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplacedPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReplacedPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
