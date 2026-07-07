package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
)

type RepoInfo struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

type RepoResponse struct {
	CloneURL string `json:"clone_url"`
	HTMLURL  string `json:"html_url"`
}

func CreatePrivateRepo(token, repoName string) (string, error) {
	body, _ := json.Marshal(RepoInfo{Name: repoName, Private: true})

	req, err := http.NewRequest("POST", "https://api.github.com/user/repos", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса к GitHub API: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		return "", fmt.Errorf("GitHub API вернул %d: %s", resp.StatusCode, string(data))
	}

	var repo RepoResponse
	if err := json.Unmarshal(data, &repo); err != nil {
		return "", err
	}
	return repo.HTMLURL, nil
}

func PushToRepo(token, repoName, outDir string) (string, error) {
	repoURL, err := CreatePrivateRepo(token, repoName)
	if err != nil {
		return "", err
	}

	username, err := getGitHubUsername(token)
	if err != nil {
		return "", err
	}

	authURL := fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", username, token, username, repoName)

	cmds := [][]string{
		{"git", "-C", outDir, "init"},
		{"git", "-C", outDir, "add", "."},
		{"git", "-C", outDir, "commit", "-m", "feat: initial NixOS configuration via MCFN"},
		{"git", "-C", outDir, "branch", "-M", "main"},
		{"git", "-C", outDir, "remote", "add", "origin", authURL},
		{"git", "-C", outDir, "push", "-u", "origin", "main"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("ошибка '%s': %s", filepath.Base(args[0])+" "+args[len(args)-1], string(out))
		}
	}

	return repoURL, nil
}

func getGitHubUsername(token string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var user struct {
		Login string `json:"login"`
	}
	json.NewDecoder(resp.Body).Decode(&user)
	if user.Login == "" {
		return "", fmt.Errorf("не удалось получить имя пользователя GitHub (проверьте токен)")
	}
	return user.Login, nil
}
