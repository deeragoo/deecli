package decryptonite

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/scrypt"
	"crypto/aes"
	"crypto/cipher"
)

type Secrets struct {
	GithubToken string `json:"github_token"`
}

// decrypt decrypts the token using the passphrase
func decrypt(encryptedB64, passphrase string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return "", err
	}

	if len(data) < 16+12 {
		return "", errors.New("encrypted data too short")
	}
	salt := data[:16]
	nonce := data[16 : 16+12]
	ciphertext := data[16+12:]

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

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GetTokenFromSecrets prompts for passphrase, decrypts and returns token
func GetTokenFromSecrets() (string, error) {
	f, err := os.Open(os.Getenv("HOME") + "/.secrets.json")
	if err != nil {
		return "", err
	}
	defer f.Close()

	var secrets Secrets
	err = json.NewDecoder(f).Decode(&secrets)
	if err != nil {
		return "", err
	}

	fmt.Print("Enter passphrase to decrypt GitHub token: ")
	passphrase, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	passphrase = strings.TrimSpace(passphrase)

	token, err := decrypt(secrets.GithubToken, passphrase)
	if err != nil {
		return "", err
	}

	return token, nil
}
