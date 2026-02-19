package ffmpeg

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	binaryPath string
	extractErr error
	once       sync.Once
)

func BinaryPath() (string, error) {
	once.Do(func() {
		if len(binary) == 0 {
			extractErr = fmt.Errorf("no ffmpeg binary for %s/%s", runtime.GOOS, runtime.GOARCH)
			return
		}
		binaryPath, extractErr = extract()
	})
	return binaryPath, extractErr
}

func extract() (string, error) {
	dir, err := os.MkdirTemp("", "goattic-ffmpeg-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name = "ffmpeg.exe"
	}

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, binary, 0755); err != nil {
		return "", fmt.Errorf("write ffmpeg binary: %w", err)
	}

	return path, nil
}
