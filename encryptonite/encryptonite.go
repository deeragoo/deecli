package encryptonite

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/scrypt"
)

type Secrets struct {
	GithubToken string `json:"github_token"`
}

// EncryptTokenInteractive prompts user and encrypts token then saves it
func EncryptTokenInteractive() error {
	fmt.Print("Enter your GitHub Personal Access Token: ")
	token, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	token = strings.TrimSpace(token)

	fmt.Print("Enter a passphrase to encrypt the token: ")
	passphrase, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	passphrase = strings.TrimSpace(passphrase)

	encrypted, err := encrypt(token, passphrase)
	if err != nil {
		return fmt.Errorf("encryption error: %w", err)
	}

	secrets := Secrets{
		GithubToken: encrypted,
	}

	f, err := os.Create(os.Getenv("HOME") + "/.secrets.json")
	if err != nil {
		return fmt.Errorf("error creating secrets file: %w", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("Warning: failed to close secrets file:", err)
		}
	}()

	encJSON, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON marshal error: %w", err)
	}

	if _, err := f.Write(encJSON); err != nil {
		return fmt.Errorf("error writing secrets file: %w", err)
	}

	fmt.Println("Token encrypted and saved to ~/.secrets.json")
	return nil
}

func encrypt(plaintext, passphrase string) (string, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	key, err := scrypt.Key([]byte(passphrase), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nil, nonce, []byte(plaintext), nil)

	var buffer bytes.Buffer
	buffer.Write(salt)
	buffer.Write(nonce)
	buffer.Write(ciphertext)

	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}
