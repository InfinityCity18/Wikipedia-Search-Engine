package webserver

import (
	"log"
	"log/slog"
	"net/http"
	"strings"
	"text/template"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/database"
)

const resultsTemplate = `
{{range .}}
<div class="article-result">
	<h3>{{.Title}} <span class="similarity">(Similarity: {{printf "%.2f" .Sim}})</span></h3>
	<p>{{.Content}}</p>
</div>
{{else}}
<p style="color: #666; font-style: italic;">No articles found matching your query.</p>
{{end}}`

func StartWebserver(db *database.Database) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) { searchHandler(w, r, db) })

	http.HandleFunc("/gopher.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "gopher.png")
	})

	slog.Info("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func searchHandler(w http.ResponseWriter, r *http.Request, db *database.Database) {
	query := r.URL.Query().Get("q")

	if len(strings.TrimSpace(query)) == 0 {
		return
	}

	arts, err := db.ProcessQuery(query)
	if err != nil {
		http.Error(w, "Error getting query", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("results").Parse(resultsTemplate)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, arts)
}
