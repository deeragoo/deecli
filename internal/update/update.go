package update

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/blang/semver/v4"
	update "github.com/inconshreveable/go-update"
	"github.com/spf13/cobra"
)

const (
	owner = "deeragoo"
	repo  = "deecli"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func getLatestRelease() (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add Authorization header for private repos
	if token := os.Getenv("GH_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Warning: failed to close response body:", err)
		}
	}()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}

	if resp.StatusCode == 404 && os.Getenv("GH_TOKEN") == "" {
		fmt.Println("Repository may be private. Set GH_TOKEN to access private releases.")
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func findAssetURL(release *githubRelease) (string, error) {
	targetName := fmt.Sprintf("deecli-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		targetName += ".exe"
	}

	for _, asset := range release.Assets {
		if strings.EqualFold(asset.Name, targetName) {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", errors.New("no compatible binary found in latest release assets")
}

// UpdateSelf runs the self-update logic, comparing versions and downloading/applying update
func UpdateSelf(cmd *cobra.Command, args []string, currentVersion string) {
	fmt.Println("Checking for updates...")

	latestRelease, err := getLatestRelease()
	if err != nil {
		fmt.Println("Error fetching latest release:", err)
		return
	}

	cv, err := semver.ParseTolerant(strings.TrimPrefix(currentVersion, "v"))
	if err != nil {
		fmt.Println("Invalid current version format:", err)
		return
	}

	lv, err := semver.ParseTolerant(strings.TrimPrefix(latestRelease.TagName, "v"))
	if err != nil {
		fmt.Println("Invalid latest version format:", err)
		return
	}

	if !lv.GT(cv) {
		fmt.Printf("You are already running the latest version (%s).\n", currentVersion)
		return
	}

	fmt.Printf("New version available: %s → Downloading update...\n", latestRelease.TagName)

	assetURL, err := findAssetURL(latestRelease)
	if err != nil {
		fmt.Println("Error finding compatible binary:", err)
		return
	}

	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		fmt.Println("Failed to create download request:", err)
		return
	}

	if token := os.Getenv("GH_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to download update:", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Warning: failed to close response body:", err)
		}
	}()

	if resp.StatusCode != 200 {
		fmt.Println("Failed to download update: bad response", resp.Status)
		return
	}

	// Read the downloaded binary into memory
	binaryData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read update binary:", err)
		return
	}

	// Update the running binary (in-place)
	err = update.Apply(bytes.NewReader(binaryData), update.Options{
		TargetPath: os.Args[0],
	})
	if err != nil {
		fmt.Println("Failed to update running binary:", err)
		return
	}

	// Also update the binary in ~/.deecli/bin if it exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Failed to get user home directory:", err)
		return
	}

	deecliPath := []string{
		filepath.Join(homeDir, ".deecli", "bin", "deecli"),
		filepath.Join(homeDir, ".deecli", "deecli"),
	}

	if runtime.GOOS == "windows" {
		for i := range deecliPath {
			deecliPath[i] += ".exe"
		}
	}

	// Check if the file exists before overwriting
	for _, path := range deecliPath {
		if _, err := os.Stat(path); err == nil {
			err = update.Apply(bytes.NewReader(binaryData), update.Options{
				TargetPath: path,
			})
			if err != nil {
				fmt.Printf("Failed to update binary in %s: %v\n", path, err)
			} else {
				fmt.Printf("Updated binary in %s\n", path)
			}
		} else {
			fmt.Printf("No binary found at %s, skipping update there.\n", path)
		}
	}
	fmt.Println("Update applied successfully! Please restart the CLI.")
}
