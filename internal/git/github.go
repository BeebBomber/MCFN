package git

import (
	"fmt"
	"os/exec"
)

func Sync(repo, token string) {
	authURL := fmt.Sprintf("https://%s@github.com/%s.git", token, repo)
	exec.Command("git", "init").Run()
	exec.Command("git", "remote", "add", "origin", authURL).Run()
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "Initial NixOS config").Run()
	exec.Command("git", "push", "-u", "origin", "main").Run()
}
