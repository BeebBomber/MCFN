package validator

import (
	"os/exec"
)

func ValidateNixSyntax(path string) error {
	return exec.Command("nix-instantiate", "--parse", path).Run()
}
