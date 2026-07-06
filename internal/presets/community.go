package presets

type Preset struct {
	Name string
	Desc string
}

func GetPresets() []Preset {
	return []Preset{
		{"Gaming", "Optimized for Steam/Wine"},
		{"Developer", "Docker, VSCode, Node.js"},
	}
}
