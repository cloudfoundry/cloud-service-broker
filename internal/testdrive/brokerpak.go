package testdrive

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Brokerpak string

func BuildBrokerpak(csbPath string, paths ...string) (Brokerpak, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	bpkDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, csbPath, "pak", "build", "--target", "current", filepath.Join(paths...))
	cmd.Dir = bpkDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error building brokerpak: %w\n\n%s", err, output)
	}

	return Brokerpak(bpkDir), nil
}

func (b Brokerpak) Cleanup() error {
	return os.RemoveAll(string(b))
}
