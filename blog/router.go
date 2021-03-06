// +build appengine
package blog

import (
	"github.com/drone/routes"
	"net/http"
)

func init() {

	mux := routes.New()
	mux.Get("/blog/", ViewArticleHandler)
	mux.Get("/blog/:year/:month/:day/:title", ViewArticleHandler)
	mux.Get("/blog/tag/:tag", TagHandler)
	mux.Get("/blog/archive/:year/:month", ArchiveHandler)

	//mux.Get("/blog/comment/post", SaveCommentHandler)
	mux.Post("/blog/comment", SaveCommentHandler)

	mux.Get("/admin", AdminHandler)
	mux.Get("/admin/article/post", SaveArticleHandler)
	mux.Get("/admin/article/edit", EditArticleHandler)
	mux.Get("/admin/article/update", UpdateArticleHandler)
	mux.Get("/admin/article/delete", DeleteArticleHandler)
	mux.Get("/admin/article/preview", PreViewArticleHandler)
	mux.Get("/admin/comment", ListCommentHandler)
	mux.Del("/admin/comment", DeleteCommentHandler)
	//mux.Get("/admin/comment/delete", DeleteCommentHandler)

	mux.Get("/feed/atom", RssHandler)
	mux.Get("/feed", RssHandler)
	mux.Get("/rss.xml", RssHandler)

	mux.Get("/sitemap.xml", SitemapHandler)
	mux.Get("/sitemap", SitemapHandler)
	mux.Get("/release", ReleaseHandler)
	mux.Get("/", IndexHandler)

	http.Handle("/", mux)
}
