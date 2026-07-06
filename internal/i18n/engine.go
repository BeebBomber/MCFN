package i18n

import (
	"encoding/json"
	"os"
	"strings"
)

type Translation struct {
	Welcome        string `json:"welcome"`
	SelectMode     string `json:"select_mode"`
	HostnamePrompt string `json:"hostname_prompt"`
	DeployPrompt   string `json:"deploy_prompt"`
	GitTokenInst   string `json:"git_token_inst"`
}

var T Translation

func LoadLocale() {
	lang := os.Getenv("LANG")
	file := "en.json"
	if strings.HasPrefix(lang, "ru") {
		file = "ru.json"
	}
	data, _ := os.ReadFile("locales/" + file)
	json.Unmarshal(data, &T)
}
