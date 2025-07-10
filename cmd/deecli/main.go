package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"io"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/deeragoo/deecli/internal/update"
	"github.com/deeragoo/deecli/version"

	"github.com/deeragoo/deecli/decryptonite"
	"github.com/deeragoo/deecli/encryptonite"
	"golang.org/x/term"
)

// Version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of deecli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deecli version:", version.Version)
	},
}

func getWorkflowID(token, repo, workflowFilename string) (int64, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/workflows", repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("GitHub API error listing workflows: %s", string(body))
	}

	var data struct {
		Workflows []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
			Path string `json:"path"`
		} `json:"workflows"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	// Try to find workflow by exact path (filename)
	for _, wf := range data.Workflows {
		if strings.HasSuffix(wf.Path, workflowFilename) {
			return wf.ID, nil
		}
	}

	// Not found
	return 0, fmt.Errorf("workflow file %q not found in repo %s", workflowFilename, repo)
}

func triggerGitHubWorkflow(token, repo string, workflowID int64, ref string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/workflows/%d/dispatches", repo, workflowID)
	payload := map[string]interface{}{
		"ref": ref,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("✅ Workflow triggered successfully!")
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("GitHub API error: %s", body)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "deecli",
		Short: "deecli is an all-in-one developer shortcut CLI",
	}

	// AWS S3 list command
	awsListCmd := &cobra.Command{
		Use:   "aws-s3-ls",
		Short: "Shortcut: List AWS S3 buckets",
		Run: func(cmd *cobra.Command, args []string) {
			out, err := exec.Command("aws", "s3", "ls").CombinedOutput()
			if err != nil {
				fmt.Println("Error running aws s3 ls:", err)
				return
			}
			fmt.Println(string(out))
		},
	}

	// Docker ps command
	dockerPsCmd := &cobra.Command{
		Use:   "docker-ps",
		Short: "Shortcut: List running Docker containers",
		Run: func(cmd *cobra.Command, args []string) {
			out, err := exec.Command("docker", "ps").CombinedOutput()
			if err != nil {
				fmt.Println("Error running docker ps:", err)
				return
			}
			fmt.Println(string(out))
		},
	}

	// Update command
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update the CLI to the latest version",
		Run: func(cmd *cobra.Command, args []string) {
			update.UpdateSelf(cmd, args, version.Version)
		},
	}

	// git-init-push command
	gitInitPushCmd := &cobra.Command{
		Use:   "git-init-push [remote-url]",
		Short: "Initialize git repo, commit all, set remote origin, and push",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			remote := args[0]

			cmds := [][]string{
				{"git", "init"},
				{"git", "add", "."},
				{"git", "commit", "-m", "Initial commit"},
				{"git", "remote", "add", "origin", remote},
				{"git", "push", "-u", "origin", "main"},
			}

			for _, c := range cmds {
				out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
				if err != nil {
					fmt.Printf("Error running %v: %v\nOutput: %s\n", c, err, out)
					return
				}
				fmt.Printf("Ran: %v\nOutput: %s\n", c, out)
			}
		},
	}

	// git-tag-push command
	gitTagPushCmd := &cobra.Command{
		Use:   "git-tag-push [tag]",
		Short: "Create a git tag and push to origin",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			tag := args[0]

			cmds := [][]string{
				{"git", "tag", tag},
				{"git", "push", "origin", tag},
			}

			for _, c := range cmds {
				out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
				if err != nil {
					fmt.Printf("Error running %v: %v\nOutput: %s\n", c, err, out)
					return
				}
				fmt.Printf("Ran: %v\nOutput: %s\n", c, out)
			}
		},
	}

	// github-create-repo command
	githubCreateRepoCmd := &cobra.Command{
		Use:   "github-create-repo [repo-name]",
		Short: "Create a GitHub repository using GitHub API",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			token := os.Getenv("GH_TOKEN")
			if token == "" {
				var err error
				token, err = decryptonite.GetTokenFromSecrets()
				if err != nil {
					fmt.Println("Error getting GitHub token:", err)
					return
				}
				fmt.Println("Using decrypted GitHub token from ~/.secrets.json")
			}

			repoName := args[0]
			private, _ := cmd.Flags().GetBool("private")

			username, err := getGitHubUsername(token)
			if err != nil {
				fmt.Println("Failed to get GitHub username:", err)
				return
			}

			fmt.Printf("You are authenticated as GitHub user: %s\n", username)
			fmt.Printf("You are about to create repository:\n  Name: %s\n  Private: %t\n", repoName, private)
			fmt.Print("Do you want to proceed? (y/n): ")

			reader := bufio.NewReader(os.Stdin)
			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(strings.ToLower(confirm))
			if confirm != "y" && confirm != "yes" {
				fmt.Println("Aborted by user.")
				return
			}

			err = createGitHubRepo(token, repoName, private)
			if err != nil {
				fmt.Println("Error creating repo:", err)
			} else {
				fmt.Printf("GitHub repo '%s' created successfully under user '%s'.\n", repoName, username)
			}
		},
	}
	githubCreateRepoCmd.Flags().BoolP("private", "p", false, "Create a private repository")

	// Encrypt token command
	encryptTokenCmd := &cobra.Command{
		Use:   "encrypt-token",
		Short: "Interactively encrypt a token and save to ~/.secrets.json",
		Run: func(cmd *cobra.Command, args []string) {
			err := encryptonite.EncryptTokenInteractive()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
		},
	}

	// decrypt-token command
	decryptTokenCmd := &cobra.Command{
		Use:   "decrypt-token",
		Short: "Decrypt and display a token from ~/.secrets.json",
		Run: func(cmd *cobra.Command, args []string) {
			token, err := decryptonite.GetTokenFromSecrets()
			if err != nil {
				fmt.Println("Error decrypting token:", err)
				return
			}
			fmt.Println("Decrypted token:", token)
			fmt.Println(token)
		},
	}
	
	
// delete-token command
var deleteTokenCmd = &cobra.Command{
	Use:   "delete-token",
	Short: "Delete a token from ~/.secrets.json after verifying passphrase",
	Run: func(cmd *cobra.Command, args []string) {
		secretsFile := os.Getenv("HOME") + "/.secrets.json"
		secrets := encryptonite.Secrets{}

		// Load secrets
		f, err := os.Open(secretsFile)
		if err != nil {
			fmt.Println("Error opening secrets file:", err)
			return
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				fmt.Println("Warning: failed to close file:", cerr)
			}
		}()

		if err := json.NewDecoder(f).Decode(&secrets); err != nil {
			fmt.Println("Error decoding secrets file:", err)
			return
		}

		// Get token name
		fmt.Print("Enter token name to delete: ")
		tokenName, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		tokenName = strings.TrimSpace(tokenName)

		encryptedToken, exists := secrets[tokenName]
		if !exists {
			fmt.Printf("Token %q not found.\n", tokenName)
			return
		}

		// Confirm deletion
		fmt.Printf("Are you sure you want to delete token %q? (y/n): ", tokenName)
		confirm, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))
		if confirm != "y" && confirm != "yes" {
			fmt.Println("Aborted by user.")
			return
		}

		// Ask for passphrase to verify
		fmt.Print("Enter passphrase for token to confirm deletion: ")
		passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			fmt.Println("Error reading passphrase:", err)
			return
		}
		passphrase := strings.TrimSpace(string(passBytes))

		// Attempt to decrypt to verify passphrase
		_, err = decryptonite.Decrypt(encryptedToken, passphrase)
		if err != nil {
			fmt.Println("Passphrase incorrect or decryption failed. Aborting deletion.")
			return
		}

		// Delete the token
		delete(secrets, tokenName)

		// Save updated secrets
		fw, err := os.Create(secretsFile)
		if err != nil {
			fmt.Println("Error writing secrets file:", err)
			return
		}
		defer func() {
			if cerr := fw.Close(); cerr != nil {
				fmt.Println("Warning: failed to close file:", cerr)
			}
		}()

		encJSON, err := json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			fmt.Println("Error encoding secrets:", err)
			return
		}

		if _, err := fw.Write(encJSON); err != nil {
			fmt.Println("Error saving secrets file:", err)
			return
		}

		fmt.Printf("Token %q deleted successfully.\n", tokenName)
	},
}

var githubRunWorkflowCmd = &cobra.Command{
	Use:   "github-run-workflow <repo> <workflow-file>",
	Short: "Trigger a GitHub Actions workflow via workflow_dispatch",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		repo := args[0]         // Format: "owner/repo"
		workflowFilename := args[1] // e.g. "ci.yml" or "build.yml"

		ref, _ := cmd.Flags().GetString("ref")
		if ref == "" {
			ref = "main"
		}

		token := os.Getenv("GH_TOKEN")
		if token == "" {
			var err error
			token, err = decryptonite.GetTokenByName("github_token") // pass token name explicitly
			if err != nil {
				fmt.Println("Error getting GitHub token:", err)
				return
			}
		}

		fmt.Printf("Fetching workflow ID for %q in repo %q...\n", workflowFilename, repo)
		workflowID, err := getWorkflowID(token, repo, workflowFilename)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Printf("Triggering workflow ID %d on repo %q (ref: %s)...\n", workflowID, repo, ref)
		err = triggerGitHubWorkflow(token, repo, workflowID, ref)
		if err != nil {
			fmt.Println("Error:", err)
		}
	},
}

githubRunWorkflowCmd.Flags().String("ref", "main", "Git branch or tag to run the workflow on")
githubRunWorkflowCmd.Flags().String("token", "github_token", "Token name in ~/.secrets.json")
	rootCmd.AddCommand(
		awsListCmd,
		dockerPsCmd,
		updateCmd,
		gitInitPushCmd,
		gitTagPushCmd,
		githubCreateRepoCmd,
		encryptTokenCmd,
		decryptTokenCmd,
		versionCmd,
		deleteTokenCmd,
		githubRunWorkflowCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// getGitHubUsername fetches the GitHub username for the provided token
func getGitHubUsername(token string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Warning: failed to close response body:", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(resp.Body); err != nil {
			fmt.Println("Warning: failed to read response body:", err)
		}
		return "", fmt.Errorf("GitHub API error: %s", buf.String())
	}

	var userData struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		return "", err
	}
	return userData.Login, nil
}

// createGitHubRepo calls GitHub API to create a new repo
func createGitHubRepo(token, repoName string, private bool) error {
	url := "https://api.github.com/user/repos"

	payload := map[string]interface{}{
		"name":    repoName,
		"private": private,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Warning: failed to close response body:", err)
		}
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		fmt.Println("Warning: failed to read error response body:", err)
	}
	return fmt.Errorf("GitHub API error: %s", buf.String())
}
