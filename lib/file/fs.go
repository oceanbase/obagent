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
