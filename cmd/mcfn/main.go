package main

import (
	"fmt"
	"mcfn/internal/i18n"
	"mcfn/internal/deploy"
	"mcfn/internal/generator"
)

func main() {
	i18n.LoadLocale() // Автоопределение языка
	fmt.Println(i18n.T.Welcome)

	// Логика выбора Wizard/Pro
	// Логика выбора Community Presets
	// Логика вызова Generator
	// Логика вызова Deploy (если выбран удаленный хост)
}
