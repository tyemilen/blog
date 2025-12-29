package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"

	"dokinar.ik/blog/server"
	"dokinar.ik/blog/services"
	_ "github.com/mattn/go-sqlite3"
)

func help() {
	fmt.Println("commands:\n\tserve <port>\n\tnew <article>\n\tedit <slug>")
}

func newArticle(db *sql.DB, title string) {
	tmp, err := os.CreateTemp("", fmt.Sprintf("%s-n-*.md", title))

	if err != nil {
		panic(err)
	}

	defer os.Remove(tmp.Name())

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	cmd := exec.Command(editor, tmp.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		panic(err)
	}
	if err := tmp.Close(); err != nil {
		log.Fatal("Something went wrong:", err, "\n", "Article path:", tmp.Name())
	}

	content, err := os.ReadFile(tmp.Name())

	if err != nil {
		log.Fatal("Something went wrong:", err, "\n", "Article path:", tmp.Name())
	}

	slug, err := services.CreateArticle(db, title, string(content))

	if err != nil {
		log.Fatal("Something went wrong:", err, "\n", "Article path:", tmp.Name())
	}

	fmt.Println(slug)
}

func editArticle(db *sql.DB, slug string) {
	article, err := services.FindArticle(db, slug)
	if err != nil {
		log.Fatal("Failed to load article:", err)
	}

	tmp, err := os.CreateTemp("", fmt.Sprintf("%s-e-*.md", article.Title))
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(article.Content); err != nil {
		log.Fatal(err)
	}

	if err := tmp.Close(); err != nil {
		log.Fatal(err)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	cmd := exec.Command(editor, tmp.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	content, err := os.ReadFile(tmp.Name())
	if err != nil {
		log.Fatal(err)
	}

	if err := services.UpdateArticle(db, slug, string(content)); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Article updated:", slug)
}

func main() {
	db, err := sql.Open("sqlite3", "./database.db")

	if err != nil {
		log.Fatal("Error while opening db damn: ", err.Error())
	}

	if len(os.Args) < 3 {
		help()
		return
	}

	command := os.Args[1]

	switch command {
	case "serve":
		server.Serve(fmt.Sprintf("0.0.0.0:%s", os.Args[2]), db)
	case "new":
		newArticle(db, os.Args[2])
	case "edit":
		editArticle(db, os.Args[2])
	default:
		help()
	}
}
