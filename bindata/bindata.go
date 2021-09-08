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

package bindata

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// ModTime return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _assetsI18nErrorEnJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x96\x5d\x8e\xe3\x36\x0c\xc7\xdf\xe7\x14\xc4\x00\xc1\xbe\x6c\x75\x80\x79\xdc\x6e\x0b\x2c\x50\xa0\xe8\xc7\x1e\x40\x96\xe9\x44\xb0\x2c\x7a\x28\x29\x89\x51\xf4\xee\x0b\x49\x96\x62\x27\x76\x30\xfb\x32\x18\xfd\x49\xfe\x44\x91\x8c\xe4\xff\x5e\x00\x5e\x91\x59\x34\xb2\x15\x8c\xef\x01\x9d\x7f\x7d\x83\xd7\x2f\xb2\x85\x79\xf9\x06\x87\xf3\xeb\xe7\xe2\xa7\x8d\xc1\xa3\x34\x42\xf2\x31\x0c\x68\x93\xf3\xb7\xac\x41\xd1\xd6\x11\xc1\xe2\x75\x44\xe5\xb1\x8d\xbe\xdf\xeb\x0a\x90\x99\x78\xf6\x2d\xce\x2d\x5d\xac\x21\xd9\x8a\x4e\x1b\x8c\xfe\xbf\x4a\x6b\xc9\x43\xd1\x21\xea\xd0\x31\x0d\x10\xd8\xdc\xa5\x66\xcf\xd2\xe8\x56\xa8\x13\xaa\xde\x85\x21\xa5\x96\x35\xa8\x5a\xf5\xbe\xb0\xf6\x78\xbf\x4d\x12\xc1\x53\xde\xe6\x70\xfe\x0c\x8c\xd2\x91\x5d\x6f\xd4\x69\xfb\x90\x60\xd4\x72\x54\xb0\x2d\xf2\x6e\x6c\x4a\x24\x05\x0b\xbc\x6a\xe7\xdd\x82\x91\x6c\x70\x39\xa1\x3f\x21\x97\x14\x20\xbb\xdd\xd1\x2a\x8e\x51\x7a\x14\xad\x66\x54\x9e\x78\x5a\xd2\x92\x09\xaa\x69\x37\x25\xc6\x81\xce\xdb\x8c\x6c\xfa\x00\x43\x9d\xe8\x62\xb7\xd3\x88\x16\xe8\x88\x9f\x51\xee\x8e\xe3\xa6\xc1\x68\xdb\x3f\x1e\xc6\x4d\x43\x43\x46\x2b\x88\xe6\x3c\x06\x87\x73\xec\xd7\x8d\xb8\x06\x7a\xe9\x7a\x61\xc9\x8b\x8e\x82\x4d\x03\xf8\xaf\x74\x3d\xb8\x11\x95\xee\x34\xb6\xd0\x4c\xe0\xa9\x47\x0b\xa9\x89\xd1\x69\x0d\x78\x0f\xc8\x93\x18\xa5\xea\xe5\x31\xf5\xfb\xaf\x28\x80\xa3\xce\x5f\x24\x23\xcc\x16\xe8\xa4\x36\xd8\x6e\x17\x47\x5b\xe7\xa5\x31\x4b\xca\xb7\x2c\xfd\x1c\x27\xd8\x0d\xd2\xf7\x22\xfe\x1c\x0b\xaf\x9e\xa5\xf2\x4b\xd2\x6f\x59\xfa\x20\x67\x3d\xd0\x23\x93\x42\xe7\xf6\x66\x7a\x36\xe7\x59\x46\xab\x10\xa8\xdb\x9d\xa5\x23\xfa\xca\xd3\xb6\xa3\x05\xed\x88\xbe\xb2\xa2\xe9\x19\xc6\x79\x1a\x0b\x67\x81\x88\x72\x65\xec\x0e\xa2\x9b\x9c\xc7\x41\xb4\xda\xf5\x29\x9f\xe0\xe6\x22\x2d\x12\x89\x46\x48\x86\xcd\x34\x0a\x8a\x9a\xdb\x00\x30\xfe\x92\x0a\xb2\x1c\x81\x3f\xbf\x40\xd5\xe7\x52\xaf\x0f\x12\x01\x24\x1a\xb4\xea\x14\xe3\xbe\x12\x68\x82\xb4\xdc\x71\x77\xc8\x67\x8c\x05\x90\x9c\xee\xe7\x7f\xe2\x3f\x50\xf4\xcd\xa0\xdc\xc5\x1a\x2a\x55\x2c\x8f\x6e\xe6\x2b\x2e\xa5\x56\xe3\x6f\xc6\xbd\x74\x1b\x22\xef\x3c\xcb\x31\x3d\x25\x65\x11\x4f\xba\xb9\xb7\x41\x69\x63\x58\x2b\xbd\x4c\x57\x63\xee\x57\x94\x63\x4c\x94\xd3\x6d\xe8\x9e\x87\xc7\x0a\x3f\x46\xc7\x12\xef\x07\x73\xb0\x22\x8c\x47\x96\x2d\x0a\xa7\x58\x8f\xa9\x60\xbf\x27\xcf\x78\xaf\x70\xb0\x30\xdb\x21\xdb\xdf\x3e\x1d\xce\x9f\x16\xed\x6d\xa4\xea\xc3\x98\x8b\x5d\x16\xf2\x38\xbf\x8c\xb9\xf2\x59\x85\xa4\x6e\x66\x51\x19\x34\x6e\x20\x68\xfc\x30\x21\xb7\x71\x89\x10\x64\x8d\xb6\x8b\x36\xae\x50\xd9\xf8\x94\x98\xaa\xbb\x22\xde\xd5\x78\x9d\xdb\x6e\xa5\xb7\x78\xb1\xb3\x0f\xa0\xdc\xee\x25\xa1\x7e\x1c\x84\xa1\xd6\xa7\x3c\xbf\x5f\xc3\x50\xcb\x53\x5e\xcc\xcd\x60\x6a\x46\xa6\xeb\xf4\xf8\xab\x48\xf2\xce\x24\xe7\x10\xc6\xd1\x48\x85\x65\xdd\x2e\x26\xe5\xef\x6c\x82\x9b\xe9\x34\xcf\xc9\x53\x62\x9e\x96\x0d\xde\x2a\xa9\x44\x7b\x86\xc9\xc5\xbc\xeb\x47\x3d\xd1\x6e\x2b\x6a\x38\x8d\x93\x40\xaf\x84\x22\xeb\xd1\xce\xf7\x36\x8d\x13\xa0\x57\x50\xc4\xed\x72\xe6\x51\x48\x7f\x5b\x21\x0d\xa3\x6c\xa7\xf8\x6b\xb2\xda\x1e\x23\x26\x5b\x60\xb6\xc0\x6c\x81\x8b\xf6\x27\x18\x75\xfe\x54\xba\x65\xb4\xa2\xc5\xe7\xfa\x91\xa4\x5d\x7a\xa2\x8b\x61\x3b\x14\xaf\xda\x63\x2b\xde\x83\x56\xbd\x49\x5f\x21\xa9\xd2\x30\x33\xca\x43\x36\x2f\xb3\x37\x14\xef\xdb\xe1\x8c\x74\x5e\x0c\xe8\xe2\xdd\x2e\x9c\x0f\x4d\x24\xfd\x21\x9d\x87\x59\x84\x24\xbe\xfc\xff\xf2\x23\x00\x00\xff\xff\x72\x8b\x79\xab\x40\x0b\x00\x00")

func assetsI18nErrorEnJsonBytes() ([]byte, error) {
	return bindataRead(
		_assetsI18nErrorEnJson,
		"assets/i18n/error/en.json",
	)
}

func assetsI18nErrorEnJson() (*asset, error) {
	bytes, err := assetsI18nErrorEnJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/i18n/error/en.json", size: 2880, mode: os.FileMode(420), modTime: time.Unix(1630584784, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"assets/i18n/error/en.json": assetsI18nErrorEnJson,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"assets": {nil, map[string]*bintree{
		"i18n": {nil, map[string]*bintree{
			"error": {nil, map[string]*bintree{
				"en.json": {assetsI18nErrorEnJson, map[string]*bintree{}},
			}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
