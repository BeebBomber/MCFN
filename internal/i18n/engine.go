package i18n

import (
	"encoding/json"
	"os"
	"strings"
)

type Translation struct {
	Welcome          string `json:"welcome"`
	SelectMode       string `json:"select_mode"`
	HostnamePrompt   string `json:"hostname_prompt"`
	DeployPrompt     string `json:"deploy_prompt"`
	GitTokenInst     string `json:"git_token_inst"`
	PresetMsg        string `json:"preset_msg"`
	ValidationStart  string `json:"validation_start"`
	ErrorSyntax      string `json:"error_syntax"`
	GitRepoPrompt    string `json:"git_repo_prompt"`
	GitTokenPrompt   string `json:"git_token_prompt"`
}

var T Translation

func LoadLocale() {
	lang := os.Getenv("LANG")
	file := "locales/en.json"
	if strings.HasPrefix(lang, "ru") {
		file = "locales/ru.json"
	}
	data, _ := os.ReadFile(file)
	json.Unmarshal(data, &T)
}
