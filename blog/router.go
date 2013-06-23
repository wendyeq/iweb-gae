// +build appengine
package blog

import (
	"github.com/drone/routes"
	"net/http"
)

func init() {

	mux := routes.New()
	mux.Get("/blog/:id", ViewArticleHandler)

	// url like /blog/2013/05/08/golang
	mux.Get("/blog/:year([0-9]{4})/:month([0-9]{2})/:day([0-9]{2})/:title", ViewArticleHandler)
	mux.Get("/blog/tag/:tag", TagHandler)
	mux.Get("/blog/archive/:year([0-9]{4})/:month([0-9]{2})", ArchiveHandler)

	mux.Post("/blog/comment", SaveCommentHandler)

	mux.Get("/admin", AdminHandler)
	mux.Post("/admin/article/post", SaveArticleHandler)
	mux.Get("/admin/article/edit", EditArticleHandler)
	mux.Post("/admin/article/update", UpdateArticleHandler)
	mux.Post("/admin/article/delete", DeleteArticleHandler)
	mux.Post("/admin/article/preview", PreViewArticleHandler)

	mux.Get("/admin/comment", ListCommentHandler)
	mux.Post("/admin/comment/delete", DeleteCommentHandler)

	mux.Get("/feed/atom", RssHandler)
	mux.Get("/feed", RssHandler)
	mux.Get("/rss.xml", RssHandler)

	mux.Get("/sitemap.xml", SitemapHandler)
	mux.Get("/sitemap", SitemapHandler)
	mux.Get("/release", ReleaseHandler)
	mux.Get("/", IndexHandler)

	http.Handle("/", mux)
}
