/*
 * Copyright (c) 2023 OceanBase
 * OBAgent is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oceanbase/obagent/tests/testutil"
)

func TestFileImpl_SymbolicLinkExists(t *testing.T) {
	f := FileImpl{}
	exists, err := f.SymbolicLinkExists("/home/admin/oceanbase/cgroup")
	assert.NoError(t, err)
	assert.False(t, exists)

	tmpDir1, err := ioutil.TempDir(testutil.TempDir, "a")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir1)
	tmpDir2, err := ioutil.TempDir(testutil.TempDir, "b")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir2)
	linkPath := filepath.Join(tmpDir2, "test")
	err = f.CreateSymbolicLink(tmpDir1, linkPath)
	assert.NoError(t, err)
	exists, err = f.SymbolicLinkExists(linkPath)
	assert.NoError(t, err)
	assert.True(t, exists)

}
