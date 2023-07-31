package crypto

type PlainCrypto struct{}

func (p *PlainCrypto) Encrypt(raw string) (string, error) {
	return raw, nil
}

func (p *PlainCrypto) Decrypt(raw string) (string, error) {
	return raw, nil
}
