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
	"golang.org/x/term"
)

type Secrets map[string]string

func EncryptTokenInteractive() error {
	fmt.Println("⚠️  WARNING: If you forget this passphrase, your token cannot be recovered.")
	fmt.Println("Save your passphrase securely (e.g., password manager).")
	fmt.Println()

	fmt.Print("Enter token name (e.g. github, aws, stripe): ")
	tokenName, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	tokenName = strings.TrimSpace(tokenName)

secrets := Secrets{}
secretsFile := os.Getenv("HOME") + "/.secrets.json"

// Load existing secrets (if any)
fi, err := os.Stat(secretsFile)
if err == nil {
    if fi.Size() == 0 {
        secrets = Secrets{} // empty file, no secrets yet
    } else {
        f, err := os.Open(secretsFile)
        if err != nil {
            return fmt.Errorf("error opening secrets file: %w", err)
        }
        		
        defer func() {
			if cerr := f.Close(); cerr != nil {
				fmt.Println("Warning: failed to close file:", cerr)
			}
		}() // defer immediately after open

        if err := json.NewDecoder(f).Decode(&secrets); err != nil {
            return fmt.Errorf("error decoding secrets file: %w", err)
        }
    }
} else if !os.IsNotExist(err) {
    return fmt.Errorf("error checking secrets file: %w", err)
}

	// Check for existing token
	if _, exists := secrets[tokenName]; exists {
		fmt.Printf("Token %q already exists. Overwrite? (y/n): ", tokenName)
		confirm, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))
		if confirm != "y" && confirm != "yes" {
			fmt.Println("Aborted by user.")
			return nil
		}
	}

	fmt.Printf("Enter value for %s token: ", tokenName)
	tokenValue, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	tokenValue = strings.TrimSpace(tokenValue)

	// Ask for passphrase (with confirmation)
	fmt.Print("Enter passphrase to encrypt token: ")
	passBytes, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	passphrase := strings.TrimSpace(string(passBytes))

	fmt.Print("Confirm passphrase: ")
	confirmPassBytes, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	confirmPassphrase := strings.TrimSpace(string(confirmPassBytes))

	if passphrase != confirmPassphrase {
		return fmt.Errorf("passphrases do not match")
	}

	encrypted, err := encrypt(tokenValue, passphrase)
	if err != nil {
		return fmt.Errorf("encryption error: %w", err)
	}

	// Save token
	secrets[tokenName] = encrypted

	fw, err := os.Create(secretsFile)
	if err != nil {
		return fmt.Errorf("error creating secrets file: %w", err)
	}
	defer func() {
		if cerr := fw.Close(); cerr != nil {
			fmt.Println("Warning: failed to close file:", cerr)
		}
	}()

	encJSON, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON marshal error: %w", err)
	}

	if _, err := fw.Write(encJSON); err != nil {
		return fmt.Errorf("error writing secrets file: %w", err)
	}

	fmt.Printf("%s token encrypted and saved to ~/.secrets.json\n", tokenName)
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