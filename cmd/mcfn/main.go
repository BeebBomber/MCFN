package main

import (
	"bufio"
	"fmt"
	"mcfn/internal/git"
	"mcfn/internal/i18n"
	"mcfn/internal/generator"
	"mcfn/internal/validator"
	"os"
	"strings"
)

func main() {
	i18n.LoadLocale()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(i18n.T.Welcome)
	fmt.Println("-------------------------------------------")

	// 1. Выбор режима (пока для красоты, но переменная работает)
	fmt.Print(i18n.T.SelectMode)
	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(mode)

	// 2. Сбор данных
	fmt.Print(i18n.T.HostnamePrompt)
	hostname, _ := reader.ReadString('\n')
	hostname = strings.TrimSpace(hostname)

	fmt.Print("Введите имя пользователя: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	// 3. РЕАЛЬНАЯ ГЕНЕРАЦИЯ ФАЙЛОВ
	fmt.Println("⚙️ Генерируем конфигурационные файлы...")
	err := generator.CreateConfigs(hostname, username)
	if err != nil {
		fmt.Printf("❌ Ошибка генерации: %v\n", err)
		return
	}
	fmt.Println("✅ Файлы flake.nix и configuration.nix созданы.")

	// 4. ВАЛИДАЦИЯ (Проверка того, что мы только что создали)
	fmt.Println(i18n.T.ValidationStart)
	if err := validator.ValidateNixSyntax("flake.nix"); err != nil {
		fmt.Printf("%s %v\n", i18n.T.ErrorSyntax, err)
	} else {
		fmt.Println("✅ Синтаксис Nix в порядке.")
	}

	// 5. GITHUB SYNC
	fmt.Print("Синхронизировать с GitHub? (y/n): ")
	syncChoice, _ := reader.ReadString('\n')
	if strings.TrimSpace(syncChoice) == "y" {
		fmt.Print("Введите репозиторий (username/repo): ")
		repo, _ := reader.ReadString('\n')
		fmt.Print("Введите GitHub Token: ")
		token, _ := reader.ReadString('\n')

		fmt.Println("🚀 Начинаем синхронизацию...")
		err := git.Sync(strings.TrimSpace(repo), strings.TrimSpace(token))
		if err != nil {
			fmt.Printf("❌ Ошибка Git: %v\n", err)
		} else {
			fmt.Println("🥳 Конфигурация успешно отправлена на GitHub!")
		}
	}

	fmt.Println("\nГотово. Теперь вы можете запустить 'sudo nixos-rebuild switch --flake .'")
}
