package uploads

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

type FileType int

const (
	TypeUnknown FileType = iota //0  //Go 常量自增计数器，每出现一次 const 块自动从 0 开始，每行 + 1
	TypeTXT                     //1
	TypeMD                      //2
	TypePDF                     //3
)

// DectectType 通过文件后缀检测文件类型
func DetectType(path string) FileType {
	ext := strings.ToLower(filepath.Ext(path)) //提取后缀名并且转换为小写

	switch ext {
	case ".txt":
		return TypeTXT
	case ".md":
		return TypeMD
	case ".pdf":
		return TypePDF
	default:
		return TypeUnknown
	}

}

// ExtractText 从文件中提取文本
func ExtractText(path string, ft FileType) (string, error) {
	switch ft {
	case TypeMD, TypeTXT:
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case TypePDF:
		return extractPDF(path)
	default:
		return "", fmt.Errorf("不支持的文件类型: %v", filepath.Ext(path))
	}

}

// extraPDF 从 PDF 文件中提取文本
func extractPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close() // 关闭pdf

	var buf strings.Builder
	totalPages := r.NumPage()
	for i := 1; i <= totalPages; i++ {
		page := r.Page(i)                   // 获取第i页对象
		text, err := page.GetPlainText(nil) // 提取当前页纯文本，nil代表不用自定义渲染配置
		if err != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}
	return buf.String(), nil

}

// ProcessFile 处理文件,使用上面的 函数
func ProcessFile(path string) (string, error) {
	ft := DetectType(path)
	if ft == TypeUnknown {
		return "", fmt.Errorf("不支持的文件类型： %v", filepath.Ext(path))
	}

	return ExtractText(path, ft)
}

// WalkDir 遍历目录，对每个文件callback
func WalkDir(dir string, callback func(path string) error) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		return callback(path)
	})

}
