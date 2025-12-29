package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"dokinar.ik/blog/server"
	"dokinar.ik/blog/services"
	_ "github.com/mattn/go-sqlite3"
)

func help() {
	fmt.Println("commands:\n\tserve <port>\n\tnew <article>\n\tedit <slug>\n\tdel <slug>")
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

func confirm(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		res, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		answer := strings.ToLower(strings.TrimSpace(res))

		if answer == "y" || answer == "yes" {
			return true
		} else if answer == "n" || answer == "no" {
			return false
		} else {
			fmt.Println("Invalid input")
		}
	}
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
	case "del":
		article, err := services.FindArticle(db, os.Args[2])

		if err != nil {
			log.Fatalln("Article not found")
		}

		fmt.Printf(
			"Article:\n\tid: %d\n\tslug: %s\n\ttitle: %s\n\tcontent: %s\n\tcreated: %d\n",
			article.ID, article.Slug, article.Title, article.Content[:56], article.Created_at)

		confirmdel := confirm("Delete?")

		if confirmdel {
			fmt.Println("Done")
		} else {
			log.Fatalln("Abort")
		}
	default:
		help()
	}
}
