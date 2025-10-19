package backend

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// FileController 文件操作控制器
type FileController struct{}

// NewFileController 创建文件控制器
func NewFileController() *FileController {
	return &FileController{}
}

// HandleFile 处理文件操作
func (fc *FileController) HandleFile(command string) string {
	// 创建文件
	if strings.Contains(command, "创建文件") {
		return fc.createFile(command)
	}

	// 写入内容
	if strings.Contains(command, "写入") {
		return fc.writeFile(command)
	}

	return "未识别的文件操作命令"
}

// createFile 创建文件
func (fc *FileController) createFile(command string) string {
	parts := strings.Split(command, "创建文件")
	if len(parts) > 1 {
		filename := strings.TrimSpace(parts[1])
		if filename != "" {
			cmd := exec.Command("touch", filename)
			err := cmd.Run()
			if err != nil {
				return fmt.Sprintf("创建文件失败: %v", err)
			}
			return fmt.Sprintf("✅ 已创建文件: %s", filename)
		}
	}
	return "请指定文件名"
}

// writeFile 写入文件
func (fc *FileController) writeFile(command string) string {
	// 简单实现：写入内容到文件
	// 格式："写入 <文件名> <内容>"
	parts := strings.Fields(command)
	if len(parts) >= 3 {
		filename := parts[1]
		content := strings.Join(parts[2:], " ")

		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			return fmt.Sprintf("写入文件失败: %v", err)
		}
		return fmt.Sprintf("✅ 已写入内容到: %s", filename)
	}
	return "请指定文件名和内容"
}

// ReadFile 读取文件内容
func (fc *FileController) ReadFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// DeleteFile 删除文件
func (fc *FileController) DeleteFile(filename string) error {
	return os.Remove(filename)
}

// FileExists 检查文件是否存在
func (fc *FileController) FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
