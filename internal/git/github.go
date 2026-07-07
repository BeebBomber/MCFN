package git

import (
	"fmt"
	"os/exec"
)

func Sync(repo, token string) error {
	// Формируем URL с токеном для авторизации
	authURL := fmt.Sprintf("https://%s@github.com/%s.git", token, repo)

	// Последовательность команд Git
	commands := [][]string{
		{"git", "init"},
		{"git", "add", "."},
		{"git", "commit", "-m", "MCFN: Initial NixOS configuration"},
		{"git", "branch", "-M", "main"},
		{"git", "remote", "add", "origin", authURL},
		{"git", "push", "-u", "origin", "main", "-f"}, // -f для первой принудительной синхронизации
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("ошибка при выполнении %v: %v", cmdArgs, err)
		}
	}
	return nil
}
