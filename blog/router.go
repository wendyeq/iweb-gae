// +build appengine
package blog

import (
	"net/http"
)

func init() {
	//http.HandleFunc("/static/", StaticHandler)
	//http.HandleFunc("/ueditor/", StaticHandler)

	http.HandleFunc("/release", ReleaseHandler)
	//http.HandleFunc("/login", LoginHandler)
	//http.HandleFunc("/logout", LogoutHandler)

	http.HandleFunc("/blog/", ViewArticleHandler)
	http.HandleFunc("/blog/tag/", TagHandler)
	http.HandleFunc("/blog/archive/", ArchiveHandler)

	http.HandleFunc("/blog/comment/post", SaveCommentHandler)

	http.HandleFunc("/admin/", AdminHandler)
	http.HandleFunc("/admin/article/post", SaveArticleHandler)
	http.HandleFunc("/admin/article/edit", EditArticleHandler)
	http.HandleFunc("/admin/article/update", UpdateArticleHandler)
	http.HandleFunc("/admin/article/delete", DeleteArticleHandler)
	http.HandleFunc("/admin/article/preview", PreViewArticleHandler)

	http.HandleFunc("/admin/comment/list", ListCommentHandler)
	http.HandleFunc("/admin/comment/delete", DeleteCommentHandler)
	http.HandleFunc("/feed", RssHandler)
	http.HandleFunc("/feed/atom", RssHandler)
	http.HandleFunc("/rss.xml", RssHandler)
	http.HandleFunc("/sitemap", SitemapHandler)
	http.HandleFunc("/sitemap.xml", SitemapHandler)
	http.HandleFunc("/", IndexHandler)
}
