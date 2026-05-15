package dict

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/articles"
)

const batchSize = 30000

type WordEntry struct {
	Doc_count    int
	Global_count int
}

func CreateDict(path string) (map[string]*WordEntry, []*articles.Article, int, error) {
	dict := make(map[string]*WordEntry)
	articles_list := []*articles.Article{}
	files, err := os.ReadDir(path)
	if err != nil {
		slog.Error("Failed to read directory", "error", err)
		return nil, nil, 0, err
	}
	length := len(files)
	for i := 0; i < length; i += batchSize {
		fmt.Println(i)
		end := min(i + batchSize, length)
		window := files[i:end]
		var wg sync.WaitGroup
		ch := make(chan []string, batchSize)
		for _, file := range window {
			wg.Add(1)
			wg.Go(func() {
				art, err := articles.LoadArticleFromJson(path + "/" + file.Name())
				articles_list = append(articles_list, art)
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
			doc_bool := false
			for _, word := range words {
				key, ok := dict[word]
				if ok {
					key.Global_count++
					if !doc_bool {
						key.Doc_count++
						doc_bool = true
					}
				} else {
					dict[word] = &WordEntry{1, 1}
					doc_bool = true
				}
			}
		}
	}
	return dict, articles_list, length, nil
}
