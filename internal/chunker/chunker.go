package chunker

import "strings"

type Chunk struct {
	ID   int
	Text string
}

// SplitText 固定长度分割文本，带overlap
func SplitText(text string, chunkSize int, overlap int) []Chunk {
	runes := []rune(text) //用rune支持中文,可以计算中文个数
	var chunks []Chunk
	id := 0

	for i := 0; i < len(runes); i += chunkSize - overlap {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunkText := strings.TrimSpace(string(runes[i:end]))
		if len(chunkText) > 0 {
			chunks = append(chunks, Chunk{ID: id, Text: chunkText})
			id++
		}
		if end == len(runes) {
			break
		}
	}
	return chunks
}
