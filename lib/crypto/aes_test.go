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
