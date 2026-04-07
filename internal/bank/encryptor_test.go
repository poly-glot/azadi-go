package bank

import "testing"

func TestEncryptor_RoundTrip(t *testing.T) {
	enc, err := NewEncryptor("dev-only-key-change-in-prod-32ch")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{"account number", "12345678"},
		{"sort code", "20-30-40"},
		{"empty", ""},
		{"long string", "this is a much longer string to encrypt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := enc.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("encrypt: %v", err)
			}
			if encrypted == tt.plaintext && tt.plaintext != "" {
				t.Error("encrypted should differ from plaintext")
			}
			decrypted, err := enc.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("decrypt: %v", err)
			}
			if decrypted != tt.plaintext {
				t.Errorf("got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptor_DifferentCiphertexts(t *testing.T) {
	enc, _ := NewEncryptor("dev-only-key-change-in-prod-32ch")
	e1, _ := enc.Encrypt("12345678")
	e2, _ := enc.Encrypt("12345678")
	if e1 == e2 {
		t.Error("same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestEncryptor_ShortKey(t *testing.T) {
	_, err := NewEncryptor("short")
	if err == nil {
		t.Error("should fail with short key")
	}
}

func TestEncryptor_InvalidCiphertext(t *testing.T) {
	enc, _ := NewEncryptor("dev-only-key-change-in-prod-32ch")
	_, err := enc.Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Error("should fail with invalid ciphertext")
	}
}
