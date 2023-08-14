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

package config

import (
	"github.com/oceanbase/obagent/lib/crypto"
)

var (
	configCrypto crypto.Crypto
)

func InitCrypto(filename string, cryptoMethod crypto.CryptoMethod) (err error) {
	switch cryptoMethod {
	case crypto.AES:
		configCrypto, err = crypto.NewAESCrypto(filename)
	case crypto.PLAIN:
		configCrypto, err = &crypto.PlainCrypto{}, nil
	default:
		configCrypto, err = &crypto.PlainCrypto{}, nil
	}
	return
}
