package articles

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Article struct {
	Title   string `json:"title"`
	Content string `json:"text"`
}

func (art *Article) ReturnWordsNonUnique() []string {
	return strings.Fields(art.Content)
}

func (art *Article) StemAndRemoveStopWords() error {
	content, err := stemAndRemoveStopWordsString(art.Content)
	if err != nil {
		slog.Error("Failed to stem and remove stop words", "error", err)
		return err
	} else {
		art.Content = content
		return nil
	}
}

func LoadArticleFromJson(filePath string) (*Article, error) {
	file, err := os.Open(filePath)
	if err != nil {
		slog.Error("Failed to open file", "error", err)
		return nil, err
	}
	byteValue, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Error reading file", "error", err)
		return nil, err
	}
	defer file.Close()
	var art *Article
	art = new(Article)
	if err := json.Unmarshal(byteValue, art); err != nil {
		return nil, err
	}
	return art, nil
}
