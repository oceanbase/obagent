// All codes interacting with file system should be put here.
// Invoke file system operations via afero's Fs interface.
// This is convenient for mock testing.

package file

import (
	"github.com/spf13/afero"
)

// OsFs is a Fs implementation that uses functions provided by the os package.
// When in unit testing, replace this with afero.MemMapFs.
var Fs afero.Fs = afero.OsFs{}
