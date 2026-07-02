package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AppendMarkdown 将 Markdown 内容追加到 Obsidian 指定文件
func AppendMarkdown(vaultPath, relativePath, content string) error {
	fullPath := filepath.Join(vaultPath, relativePath)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 打开文件（不存在则创建，存在则追加）
	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 写入内容
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// AppendMarkdownIfMissing 在聚合 Markdown 中不存在指定标题 ID 时才追加内容。
// 返回值 appended 表示本次是否实际写入。
func AppendMarkdownIfMissing(vaultPath, relativePath, content, docID string) (bool, error) {
	fullPath := filepath.Join(vaultPath, relativePath)

	if docID != "" {
		data, err := os.ReadFile(fullPath)
		if err != nil && !os.IsNotExist(err) {
			return false, fmt.Errorf("读取文件失败: %w", err)
		}
		if err == nil && aggregateDocExists(string(data), docID) {
			return false, nil
		}
	}

	if err := AppendMarkdown(vaultPath, relativePath, content); err != nil {
		return false, err
	}
	return true, nil
}

func aggregateDocExists(content, docID string) bool {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "# "+docID || strings.HasPrefix(line, "# "+docID+"｜") {
			return true
		}
	}
	return false
}

// EnsureFileExists 确保文件存在（为空文件）
func EnsureFileExists(vaultPath, relativePath string) error {
	fullPath := filepath.Join(vaultPath, relativePath)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 如果文件不存在，创建空文件
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		file, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("创建文件失败: %w", err)
		}
		file.Close()
	}

	return nil
}
