package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type Article struct {
	ID         int64
	Slug       string
	Title      string
	Content    string
	Languages  []string
	Created_at int64
}

func FindUsedLangs(node ast.Node) []string {
	seen := map[string]struct{}{}
	var langs []string

	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		if cb, ok := n.(*ast.CodeBlock); ok {
			if len(cb.Info) > 0 {
				lang := string(cb.Info)
				if _, exists := seen[lang]; !exists {
					seen[lang] = struct{}{}
					langs = append(langs, lang)
				}
			}
		}
		return ast.GoToNext
	})

	return langs
}

func CreateArticle(db *sql.DB, title string, content string) (string, error) {
	slug := strings.Join(strings.Split(strings.ToLower(title), " "), "-")

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(content))

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	parsed := markdown.Render(doc, renderer)

	langs, err := json.Marshal(FindUsedLangs(doc))

	if err != nil {
		fmt.Println("Warn: Could not detect any languages")
	}

	query := "INSERT INTO `articles` (slug, title, content, languages) VALUES(?, ?, ?, ?);"

	_, err = db.Exec(query, slug, title, string(parsed), string(langs))

	return slug, err
}

func FindArticle(db *sql.DB, slug string) (Article, error) {
	row := db.QueryRow("SELECT * FROM `articles` WHERE slug = ?", slug)

	var article Article
	var langsRaw string

	if err := row.Scan(&article.ID, &article.Slug, &article.Title, &article.Content, &langsRaw, &article.Created_at); err != nil {
		return article, err
	}

	json.Unmarshal([]byte(langsRaw), &article.Languages)

	return article, nil
}

func GetArticles(db *sql.DB) ([]Article, error) {
	rows, err := db.Query("SELECT * FROM `articles`;")

	if err != nil {
		return nil, err
	}

	var articles []Article

	for rows.Next() {
		var article Article
		var langsRaw string

		if err := rows.Scan(&article.ID, &article.Slug, &article.Title, &article.Content, &langsRaw, &article.Created_at); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(langsRaw), &article.Languages)

		articles = append(articles, article)
	}

	return articles, nil
}

func UpdateArticle(db *sql.DB, slug string, content string) error {
	result, err := db.Exec("UPDATE `articles` SET content = ? WHERE slug = ?", content, slug)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func DeleteArticle(db *sql.DB, slug string) error {
	result, err := db.Exec("DELETE FROM `articles` WHERE slug = ?", slug)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
