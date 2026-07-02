package obsidian

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppendMarkdown(t *testing.T) {
	// 使用临时目录
	tmpDir := t.TempDir()

	testContent := "# 测试标题\n\n这是测试内容。"

	// 测试追加到新文件
	relativePath := "test/test_file.md"
	err := AppendMarkdown(tmpDir, relativePath, testContent)
	if err != nil {
		t.Fatalf("AppendMarkdown 失败: %v", err)
	}

	// 验证文件存在
	fullPath := filepath.Join(tmpDir, relativePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Error("文件不存在")
	}

	// 验证文件内容
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	content := string(data)
	if content != testContent {
		t.Errorf("文件内容不匹配: got %q, want %q", content, testContent)
	}

	// 测试追加到已存在文件
	appendContent := "\n\n追加内容。"
	err = AppendMarkdown(tmpDir, relativePath, appendContent)
	if err != nil {
		t.Fatalf("追加失败: %v", err)
	}

	// 验证追加后内容
	data, err = os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	content = string(data)
	expectedContent := testContent + appendContent
	if content != expectedContent {
		t.Errorf("追加后内容不匹配: got %q, want %q", content, expectedContent)
	}
}

func TestAppendMarkdownIfMissingSkipsExistingDocID(t *testing.T) {
	tmpDir := t.TempDir()
	relativePath := "test/raw.md"
	first := "# RAW-ACCOUNT-SAFETY-20260702-001｜安全边际\n\n第一次内容\n"
	duplicate := "# RAW-ACCOUNT-SAFETY-20260702-001｜安全边际\n\n第二次内容\n"

	appended, err := AppendMarkdownIfMissing(tmpDir, relativePath, first, "RAW-ACCOUNT-SAFETY-20260702-001")
	if err != nil {
		t.Fatalf("AppendMarkdownIfMissing 首次写入失败: %v", err)
	}
	if !appended {
		t.Fatal("首次写入应返回 appended=true")
	}

	appended, err = AppendMarkdownIfMissing(tmpDir, relativePath, duplicate, "RAW-ACCOUNT-SAFETY-20260702-001")
	if err != nil {
		t.Fatalf("AppendMarkdownIfMissing 重复写入失败: %v", err)
	}
	if appended {
		t.Fatal("重复 docID 不应再次追加")
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, relativePath))
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if string(data) != first {
		t.Fatalf("重复写入后内容变化: %q", string(data))
	}
}

func TestEnsureFileExists(t *testing.T) {
	// 使用临时目录
	tmpDir := t.TempDir()

	// 测试创建新文件
	relativePath := "test/new_file.md"
	err := EnsureFileExists(tmpDir, relativePath)
	if err != nil {
		t.Fatalf("EnsureFileExists 失败: %v", err)
	}

	// 验证文件存在
	fullPath := filepath.Join(tmpDir, relativePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Error("文件不存在")
	}

	// 测试已存在文件
	err = EnsureFileExists(tmpDir, relativePath)
	if err != nil {
		t.Fatalf("已存在文件处理失败: %v", err)
	}

	// 验证文件未被覆盖
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if len(data) != 0 {
		t.Error("文件应保持为空")
	}
}
