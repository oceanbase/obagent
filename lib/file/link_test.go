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
