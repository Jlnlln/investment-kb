package prompt

import (
	"fmt"
	"os"
)

// Load 加载 Prompt 文件
func Load(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取 Prompt 文件失败: %w", err)
	}
	return string(data), nil
}