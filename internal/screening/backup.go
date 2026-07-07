package screening

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func Timestamp() string {
	return time.Now().Format("20060102_150405")
}

func BackupFiles(paths Paths, fullPaths []string, timestamp string) (string, error) {
	backupRoot, err := paths.BackupRoot(timestamp)
	if err != nil {
		return "", err
	}
	for _, src := range fullPaths {
		rel, err := filepath.Rel(paths.KBRoot, src)
		if err != nil {
			return "", err
		}
		dst := filepath.Join(backupRoot, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return "", err
		}
		if err := copyFile(src, dst); err != nil {
			return "", fmt.Errorf("备份失败: %s: %w", src, err)
		}
	}
	return backupRoot, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
