// +build appengine
package blog

import (
	"html/template"
	"time"
)

type ArticleMetaData struct {
	Id         string
	Author     string
	Title      string
	Permalink  string
	Tags       []string
	Summary    string
	Content    []byte
	PostTime   time.Time
	UpdateTime time.Time
	Count      int64
	Flag       int64 //1 draf, 2 public
}

type Tags struct {
	Tag       string
	ArticleId string
}

type Comment struct {
	Id        string
	ParentId  string
	ArticleId string
	Author    string
	Email     string
	Website   string
	Content   string
	PostTime  time.Time
	Flag      int64 //1 hidden, 2 public

}

type Article struct {
	MetaData ArticleMetaData
	Text     template.HTML //markdonw conver to html
	Tags     []Tags
	Comments []Comment
}

type IndexData struct {
	Articles []Article
	Tags     map[string]int
	Archives map[NewString]int
}

type NewString string
