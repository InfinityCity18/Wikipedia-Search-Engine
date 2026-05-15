package dict

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/articles"
)

const batchSize = 30000

func CreateDict(path string) (map[string]bool, error) {
	dict := make(map[string]bool)
	files, err := os.ReadDir(path)
	if err != nil {
		slog.Error("Failed to read directory", "error", err)
		return nil, err
	}
	length := len(files)
	for i := 0; i < length; i += batchSize {
		fmt.Println(i)
		end := i + batchSize
		if end > length {
			end = length
		}
		window := files[i:end]
		var wg sync.WaitGroup
		ch := make(chan []string, batchSize)
		for _, file := range window {
			wg.Add(1)
			wg.Go(func() {
				art, err := articles.LoadArticleFromJson(path + "/" + file.Name())
				if err != nil {
					slog.Error("Error in goroutine", "error", err)
					return
				}
				if err := art.StemAndRemoveStopWords(); err != nil {
					slog.Error("Error in goroutine", "error", err)
				}
				ch <- art.ReturnWordsNonUnique()
				wg.Done()
			})
		}
		go func() {
			wg.Wait()
			close(ch)
		}()
		for words := range ch {
			for _, word := range words {
				dict[word] = true
			}
		}
	}
	return dict, nil
}
