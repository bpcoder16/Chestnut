package operations

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

func ReadOrCreate[T any](filePath string, defaultValue T) (value T, err error) {
	var result T

	// 判断文件是否存在
	if _, err = os.Stat(filePath); err == nil {
		// 文件存在，读取内容
		var data []byte
		data, err = os.ReadFile(filePath)
		if err != nil {
			return result, fmt.Errorf("failed to read file: %w", err)
		}
		if err = json.Unmarshal(data, &result); err != nil {
			return result, fmt.Errorf("failed to unmarshal content: %w", err)
		}
		return result, nil
	} else if os.IsNotExist(err) {
		// 文件不存在，创建并写入默认值
		if err = WriteFile(filePath, defaultValue); err != nil {
			return result, err
		}
		return defaultValue, nil
	} else {
		return result, fmt.Errorf("failed to stat file: %w", err)
	}
}

// WriteFile 泛型写文件函数
func WriteFile[T any](filePath string, value T) (err error) {
	var data []byte
	data, err = json.MarshalIndent(value, "", "  ") // JSON 格式保存
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	if err = os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func ListFilesSorted(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() { // 只取文件
			files = append(files, entry.Name())
		}
	}

	// 按文件名排序
	sort.Strings(files)
	return files, nil
}

// ReadFile 读取文件内容为字符串
func ReadFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(data), nil
}
