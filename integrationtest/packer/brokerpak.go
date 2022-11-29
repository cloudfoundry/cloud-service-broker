// Package packer is a test helper to build brokerpaks for the integration tests
package packer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type BuildBrokerpakOption func(*buildBrokerpakCfg)

type buildBrokerpakCfg struct {
	extrafiles []func(string) error
	dir        string
}

func BuildBrokerpak(csbPath, sourcePath string, opts ...BuildBrokerpakOption) (string, error) {
	var cfg buildBrokerpakCfg
	for _, o := range opts {
		o(&cfg)
	}

	if cfg.dir == "" {
		bpkDir, err := os.MkdirTemp("", "")
		if err != nil {
			return "", err
		}
		cfg.dir = bpkDir
	}

	for _, cb := range cfg.extrafiles {
		if err := cb(cfg.dir); err != nil {
			return "", err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, csbPath, "pak", "build", "--target", "current", sourcePath)
	cmd.Dir = cfg.dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(cfg.dir)
		return "", fmt.Errorf("error building brokerpak: %w\n\n%s", err, output)
	}

	return cfg.dir, nil
}

func WithExtraFile(name, contents string) BuildBrokerpakOption {
	return func(cfg *buildBrokerpakCfg) {
		cfg.extrafiles = append(cfg.extrafiles, func(dir string) error {
			return os.WriteFile(filepath.Join(dir, name), []byte(contents), 0666)
		})
	}
}

func WithDirectory(dir string) BuildBrokerpakOption {
	return func(cfg *buildBrokerpakCfg) {
		cfg.dir = dir
	}
}
