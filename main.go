package main

import (
	"fmt"

	"github.com/shen060606/rag_koowledge_go/internal/api"
	"github.com/shen060606/rag_koowledge_go/internal/rag"
	"github.com/shen060606/rag_koowledge_go/internal/store"
	"github.com/shen060606/rag_koowledge_go/internal/uploads"
)

func main() {
	//1 初始化向量存储器
	vs := store.NewVectorStore()

	// 2 导入知识库
	err := uploads.WalkDir("uploads", func(path string) error {
		content, err := uploads.ProcessFile(path)
		if err != nil {
			fmt.Printf("[WRONG] %s: %v\n", path, err)
			return nil
		}
		if err := rag.ImportDoc(vs, content); err != nil {
			fmt.Printf("[WRONG] 导入 %s 失败: %v\n", path, err)
			return nil
		}
		fmt.Printf("[RIGHT] 已导入%s\n", path)
		return nil
	})
	if err != nil {
		fmt.Printf("遍历目录失败: %v\n", err)
	}
	//3 web服务
	r := api.Setup(vs)
	r.Run(":8088")

}
