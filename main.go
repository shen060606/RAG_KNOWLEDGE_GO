package main

import (
	"fmt"

	"os"

	"github.com/shen060606/rag_koowledge_go/internal/rag"
	"github.com/shen060606/rag_koowledge_go/internal/store"
	"github.com/shen060606/rag_koowledge_go/uploads"
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
		rag.ImportDoc(vs, content)
		fmt.Printf("[RIGHT] 已导入%s\n", path)
		return nil
	})
	if err != nil {
		fmt.Printf("遍历目录失败: %v\n", err)
	}
	//3 交互式提问
	fmt.Println("=====RAG Knowledge System===")
	fmt.Println("请输入您的问题(输入exit退出)：")

	for {
		fmt.Print("\n>")
		var question string
		fmt.Scanln(&question)
		if question == "exit" {
			os.Exit(0)
		}

		//调用回答
		_, err := rag.Ask(vs, question)
		if err != nil {
			fmt.Println("抱歉，无法回答您的问题。")
			continue
		}

	}

	fmt.Println("===RAG Knowledge System结束===")

}
