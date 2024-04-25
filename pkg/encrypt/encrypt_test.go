package encrypt

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	passphrase := "securepassphrase"
	e := New(passphrase)
	originalText := "Hello, World!"

	encryptedText, err := e.Encrypt(originalText)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if encryptedText == originalText {
		t.Fatalf("Encryption did not change the plaintext")
	}

	decryptedText, err := e.Decrypt(encryptedText)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decryptedText != originalText {
		t.Fatalf("Decryption failed: got %v, want %v", decryptedText, originalText)
	}
}

func TestEncryptEmptyText(t *testing.T) {
	passphrase := "securepassphrase"
	e := New(passphrase)
	originalText := ""

	encryptedText, err := e.Encrypt(originalText)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decryptedText, err := e.Decrypt(encryptedText)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decryptedText != originalText {
		t.Fatalf("Decryption failed: got %v, want %v", decryptedText, originalText)
	}
}

func TestEncryptDecryptWithDifferentPassphrases(t *testing.T) {
	e := New("securepassphrase")
	eWrong := New("wrongpassphrase")
	originalText := "Hello, World!"

	encryptedText, err := e.Encrypt(originalText)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = eWrong.Decrypt(encryptedText)
	if err == nil {
		t.Fatal("Decrypt should have failed with the wrong passphrase")
	}
}
