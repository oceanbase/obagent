package crypto

type CryptoMethod string

const (
	AES   CryptoMethod = "aes"
	PLAIN CryptoMethod = "plain"
)

type Crypto interface {
	Encrypt(raw string) (string, error)
	Decrypt(raw string) (string, error)
}
