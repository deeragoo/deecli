package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/deeragoo/deecli/internal/update"
	"github.com/deeragoo/deecli/version"

	"github.com/deeragoo/deecli/decryptonite"
	"github.com/deeragoo/deecli/encryptonite"
)

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
				{"git", "push", "-u", "origin", "master"},
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
	Short: "Interactively encrypt a GitHub token and save to ~/.secrets.json",
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
	Short: "Decrypt and display the GitHub token from ~/.secrets.json",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := decryptonite.GetTokenFromSecrets()
		if err != nil {
			fmt.Println("Error decrypting GitHub token:", err)
			return
		}
		fmt.Println("Decrypted GitHub token:")
		fmt.Println(token)
	},
}



	rootCmd.AddCommand(
		awsListCmd,
		dockerPsCmd,
		updateCmd,
		gitInitPushCmd,
		gitTagPushCmd,
		githubCreateRepoCmd,
		encryptTokenCmd,
		decryptTokenCmd,
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
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
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
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return fmt.Errorf("GitHub API error: %s", buf.String())
}
