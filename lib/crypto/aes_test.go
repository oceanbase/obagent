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

package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptAndDecrypt(t *testing.T) {
	aesCrypter, _ := NewAESCrypto("../../etc/.config_secret.key")
	raw := "root"
	encrypted, _ := aesCrypter.Encrypt(raw)
	decrypted, _ := aesCrypter.Decrypt(encrypted)
	require.Equal(t, decrypted, raw)
}
