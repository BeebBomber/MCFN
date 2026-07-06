package main

import (
	"bufio"
	"fmt"
	"mcfn/internal/git"
	"mcfn/internal/i18n"
	"mcfn/internal/validator"
	"os"
	"strings"
)

func main() {
	i18n.LoadLocale()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(i18n.T.Welcome)
	fmt.Println("-------------------------------------------")

	// Ввод режима
	fmt.Print(i18n.T.SelectMode)
	modeInput, _ := reader.ReadString('\n')
	mode := strings.TrimSpace(modeInput)

	// ИСПОЛЬЗУЕМ переменную mode, чтобы компилятор не ругался
	switch mode {
	case "1":
		fmt.Println("Режим: Мастер (Wizard)")
	case "2":
		fmt.Println("Режим: Профи (Pro)")
	default:
		fmt.Println("Выбран режим по умолчанию")
	}

	fmt.Print(i18n.T.HostnamePrompt)
	hostnameInput, _ := reader.ReadString('\n')
	hostname := strings.TrimSpace(hostnameInput)

	// Пример использования hostname, чтобы тоже не было ошибки
	fmt.Printf("Конфигурируем хост: %s\n", hostname)

	fmt.Println(i18n.T.ValidationStart)
	// Важно: путь должен быть правильным, для теста проверим сам main.go
	if err := validator.ValidateNixSyntax("flake.nix"); err != nil {
		fmt.Printf("%s %v\n", i18n.T.ErrorSyntax, err)
	}

	fmt.Print("GitHub Sync? (y/n): ")
	syncInput, _ := reader.ReadString('\n')
	sync := strings.TrimSpace(syncInput)

	if sync == "y" {
		fmt.Println(i18n.T.GitTokenInst)
		fmt.Print(i18n.T.GitRepoPrompt)
		repoInput, _ := reader.ReadString('\n')
		repo := strings.TrimSpace(repoInput)

		fmt.Print(i18n.T.GitTokenPrompt)
		tokenInput, _ := reader.ReadString('\n')
		token := strings.TrimSpace(tokenInput)
		
		git.Sync(repo, token)
	}

	fmt.Println("Завершено.")
}
