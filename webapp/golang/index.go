package main

import (
	"log"
	"net/http"

	"github.com/catatsuy/private-isu/webapp/golang/templates"
	"github.com/catatsuy/private-isu/webapp/golang/types"
	"github.com/valyala/quicktemplate"
)

func getIndex(w http.ResponseWriter, r *http.Request) {
	me := getSessionUser(r)

	results := []types.Post{}

	err := db.Select(&results, "SELECT p.`id`, `user_id`, `body`, `mime`, p.`created_at` FROM `posts` p INNER JOIN `users` u ON p.user_id = u.id WHERE u.del_flg = 0 ORDER BY `created_at` DESC LIMIT ?", postsPerPage)
	if err != nil {
		log.Print(err)
		return
	}

	posts, err := makePosts(results, getCSRFToken(r), false)
	if err != nil {
		log.Print(err)
		return
	}

	templates.WriteLayout(w, func(qw *quicktemplate.Writer) {
		templates.StreamContentPage(qw, getCSRFToken(r), getFlash(w, r, "notice"), posts)
	}, me)
}
