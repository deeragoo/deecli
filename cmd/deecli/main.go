package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/deeragoo/deecli/internal/update"
	"github.com/deeragoo/deecli/version"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "deecli",
		Short: "deecli is an all-in-one developer shortcut CLI",
	}

	// Existing commands...

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

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update the CLI to the latest version",
		Run: func(cmd *cobra.Command, args []string) {
			update.UpdateSelf(cmd, args, version.Version)
		},
	}

	// New command: git-init-push
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

	// New command: git-tag-push
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

	rootCmd.AddCommand(awsListCmd, dockerPsCmd, updateCmd, gitInitPushCmd, gitTagPushCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
