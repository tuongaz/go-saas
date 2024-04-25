package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

type Encryptor struct {
	passphrase string
}

func New(passphrase string) *Encryptor {
	return &Encryptor{passphrase: passphrase}
}

func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	data := []byte(plaintext)
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	key := pbkdf2.Key([]byte(e.passphrase), salt, 1000, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	finalCiphertext := append(salt, ciphertext...)
	return base64.StdEncoding.EncodeToString(finalCiphertext), nil
}

func (e *Encryptor) Decrypt(cipherText string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	salt := data[:16]
	data = data[16:]
	key := pbkdf2.Key([]byte(e.passphrase), salt, 1000, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

type Interface interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(cipherText string) (string, error)
}
