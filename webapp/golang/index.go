package main

import (
	"log"
	"net/http"

	"github.com/catatsuy/private-isu/webapp/golang/templates"
	"github.com/catatsuy/private-isu/webapp/golang/types"
)

func getIndex(w http.ResponseWriter, r *http.Request) {
	me := getSessionUser(r)

	results := []types.Post{}

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

	templates.WriteLayout(w, func() string {
		return templates.ContentPage(getCSRFToken(r), getFlash(w, r, "notice"), posts)
	}, me)
}
