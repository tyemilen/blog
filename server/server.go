package server

import (
	"database/sql"
	"fmt"
	"net/http"

	"dokinar.ik/blog/services"
	"dokinar.ik/blog/views/pages"
	"github.com/a-h/templ"
	_ "github.com/mattn/go-sqlite3"
)

func Serve(addr string, db *sql.DB) {
	server := http.NewServeMux()
	fs := http.FileServer(http.Dir("./public"))

	server.Handle("/assets/", http.StripPrefix("/assets/", fs))

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		articles, err := services.GetArticles(db)

		if err != nil {
			fmt.Printf("Something went wrong while getting articles: %s\n", err)
			return
		}

		templ.Handler(pages.Index(articles)).ServeHTTP(w, r)
	})

	server.HandleFunc("/article/{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		article, err := services.FindArticle(db, slug)

		if err != nil {
			return
		}

		templ.Handler(pages.Article(article)).ServeHTTP(w, r)
	})

	fmt.Println("Listening on", addr)

	http.ListenAndServe(addr, server)
}
