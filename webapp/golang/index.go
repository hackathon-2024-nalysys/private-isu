package main

import (
	"html/template"
	"log"
	"net/http"
)

var indexTemplate = template.Must(template.New("layout.html").Funcs(template.FuncMap{
	"imageURL": imageURL,
}).ParseFiles(
	getTemplPath("layout.html"),
	getTemplPath("index.html"),
	getTemplPath("posts.html"),
	getTemplPath("post.html"),
))

func getIndex(w http.ResponseWriter, r *http.Request) {
	me := getSessionUser(r)

	results := []Post{}

	err := db.Select(&results, "SELECT `id`, `user_id`, `body`, `mime`, `created_at` FROM `posts` ORDER BY `created_at` DESC")
	if err != nil {
		log.Print(err)
		return
	}

	posts, err := makePosts(results, getCSRFToken(r), false)
	if err != nil {
		log.Print(err)
		return
	}

	indexTemplate.Execute(w, struct {
		Posts     []Post
		Me        User
		CSRFToken string
		Flash     string
	}{posts, me, getCSRFToken(r), getFlash(w, r, "notice")})
}