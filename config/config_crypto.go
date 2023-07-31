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
