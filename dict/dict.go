package dict

import (
	"log/slog"
	"os"
	"sync"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/articles"
)

const batchSize = 300

type WordEntry struct {
	Doc_count    int
	Global_count int
}

func CreateArticlesList(path string) ([]*articles.Article, error) {
	articles_list := []*articles.Article{}

	files, err := os.ReadDir(path)
	if err != nil {
		slog.Error("Failed to read directory", "error", err)
		return nil, err
	}

	length := len(files)
	ch := make(chan *articles.Article, batchSize)
	var wg sync.WaitGroup
	for i := 0; i < length; i += batchSize {
		wg.Go(func() {
			end := min(i+batchSize, length)
			window := files[i:end]
			for _, file := range window {
				art, err := articles.LoadArticleFromJson(path + "/" + file.Name())
				if err != nil {
					slog.Error("Error in goroutine loading articles", "error", err)
				}
				if err = art.StemAndRemoveStopWords(); err != nil {
					slog.Error("Error in goroutine stemming words", "error", err)
				}
				if len(art.Content) != 0 {
					ch <- art
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	for art := range ch {
		articles_list = append(articles_list, art)
	}
	return articles_list, nil
}

func CreateDict(articles_list []*articles.Article) (map[string]*WordEntry, int, error) {
	dict := make(map[string]*WordEntry)
	length := len(articles_list)
	for i := 0; i < length; i += batchSize {
		end := min(i+batchSize, length)
		window := articles_list[i:end]
		var wg sync.WaitGroup
		ch := make(chan []string, batchSize)
		for _, article := range window {
			wg.Add(1)
			wg.Go(func() {
				ch <- article.ReturnWordsNonUnique()
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
	return dict, length, nil
}
