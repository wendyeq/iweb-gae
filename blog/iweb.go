// +build appengine
package blog

import (
	"appengine"
	"appengine/urlfetch"
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
	"net/http"
	//"regexp"
	"strconv"
	"strings"
	textTemplate "text/template"
	"time"
)

var (
	themes, _ = GetContext().Args["themes"].(string)

	indexTPL = template.Must(template.ParseFiles(
		"templates/"+themes+"/index.html",
		"templates/"+themes+"/common/header.html",
		"templates/"+themes+"/common/sidebar.html",
		"templates/"+themes+"/common/footer.html"))

	adminTPL = template.Must(template.ParseFiles(
		"templates/"+themes+"/admin.html",
		"templates/"+themes+"/common/header.html",
		"templates/"+themes+"/common/sidebar.html",
		"templates/"+themes+"/common/footer.html"))

	editTPL = template.Must(template.ParseFiles(
		"templates/" + themes + "/edit.html"))

	viewTPL = template.Must(template.ParseFiles(
		"templates/"+themes+"/view.html",
		"templates/"+themes+"/common/header.html",
		"templates/"+themes+"/common/footer.html"))

	commentTPL = template.Must(template.ParseFiles(
		"templates/"+themes+"/comments.html",
		"templates/"+themes+"/common/header.html",
		"templates/"+themes+"/common/footer.html"))

	rssTPL = textTemplate.Must(textTemplate.ParseFiles(
		"templates/" + themes + "/rss.xml"))

	sitemapTPL = textTemplate.Must(textTemplate.ParseFiles(
		"templates/" + themes + "/sitemap.xml"))

	releaseTPL = template.Must(template.ParseFiles(
		"templates/"+themes+"/release.html",
		"templates/"+themes+"/common/header.html",
		"templates/"+themes+"/common/footer.html"))
)

func server404(w http.ResponseWriter, err error) {

}

func serveError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, "Internal Server Error "+err.Error())
}

func PreViewArticleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	args := ctx.Args
	articleMetaData := &ArticleMetaData{}
	articleMetaData.Author, _ = args["author"].(string)
	articleMetaData.Title = strings.TrimSpace(r.FormValue("title"))
	tags := strings.TrimSpace(r.FormValue("tags"))
	if len(tags) > 0 {
		tags = strings.Replace(tags, "，", ",", -1)
		tags = strings.Replace(tags, " ", ",", -1)
		tag := strings.Split(tags, ",")
		articleMetaData.Tags = tag
	}
	articleMetaData.Summary = r.FormValue("summary")
	articleMetaData.Content = []byte(r.FormValue("content"))
	now := time.Now()
	articleMetaData.PostTime = now
	articleMetaData.UpdateTime = now
	articleMetaData.Flag = 1
	//article.Flag, _ = strconv.ParseInt(r.FormValue("flag"), 10, 64)
	articleMetaData.Count = 0

	article := &Article{MetaData: *articleMetaData,
		Text: template.HTML([]byte(blackfriday.MarkdownBasic(articleMetaData.Content)))}
	data := make(map[string]interface{})
	data["article"] = article

	args["title"] = articleMetaData.Title
	data["config"] = args
	viewTPL.ExecuteTemplate(w, "main", data)
}

func SaveArticleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	args := ctx.Args
	article := &ArticleMetaData{}
	article.Author, _ = args["author"].(string)
	article.Title = strings.TrimSpace(r.FormValue("title"))
	tags := strings.TrimSpace(r.FormValue("tags"))
	if len(tags) > 0 {
		tags = strings.Replace(tags, "，", ",", -1)
		tags = strings.Replace(tags, " ", ",", -1)
		tag := strings.Split(tags, ",")
		article.Tags = tag
	}
	article.Summary = r.FormValue("summary")
	article.Content = []byte(r.FormValue("content"))

	now := time.Now()
	article.PostTime = now
	article.UpdateTime = now
	article.Flag = 1
	//article.Flag, _ = strconv.ParseInt(r.FormValue("flag"), 10, 64)
	article.Count = 0
	err := article.Save(ctx)
	if err == nil {
		PingServer(w, r)
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		fmt.Fprint(w, err.Error())
	}
}

func EditArticleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	id := r.FormValue("id")
	if id != "" {
		article := new(ArticleMetaData)
		article.Id = id
		err := article.GetOne(ctx)
		if err != nil {
			serveError(w, fmt.Errorf("edit old article : id = %v, err = %v", id, err))
			return
		} else {
			editTPL.Execute(w, article)
		}
	} else {
		editTPL.Execute(w, nil)
	}
}

func UpdateArticleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	article := new(ArticleMetaData)
	article.Id = r.FormValue("id")
	tags := strings.TrimSpace(r.FormValue("tags"))
	if len(tags) > 0 {
		tags = strings.Replace(tags, "，", ",", -1)
		tags = strings.Replace(tags, " ", ",", -1)
		tag := strings.Split(tags, ",")
		article.Tags = tag
	}
	article.Summary = r.FormValue("summary")
	article.Content = []byte(r.FormValue("content"))

	now := time.Now()
	article.UpdateTime = now
	//article.Flag, _ = strconv.ParseInt(r.FormValue("flag"), 10, 64)
	article, err := article.Update(ctx)
	if err != nil {
		serveError(w, err)
		return
	}
	PingServer(w, r)
	urlStr := "/blog/" + article.PostTime.Format("2006/01/02") + "/" + article.Title
	http.Redirect(w, r, urlStr, http.StatusFound)

}

func DeleteArticleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	id := r.FormValue("id")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	article := new(ArticleMetaData)
	fmt.Println(id)
	article.Id = id
	err := article.Delete(ctx)
	if err != nil {
		serveError(w, err)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusFound)

}

// view article by id or by date and title
// url like /blog/6f1135c940fc5e928084e38d62065f50
// or url like /blog/2013/05/08/golang
func ViewArticleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	beginTime := time.Now()
	articleMetaData := &ArticleMetaData{}
	params := r.URL.Query()
	id := params.Get(":id")
	if id != "" {
		articleMetaData.Id = id
	} else {
		year := params.Get(":year")
		month := params.Get(":month")
		day := params.Get(":day")
		title := params.Get(":title")
		//month only in 1~12
		if m, err := strconv.Atoi(month); err != nil || m > 12 || m < 1 {
			http.NotFound(w, r)
			return
		}
		//day only in 1~31
		if d, err := strconv.Atoi(month); err != nil || d > 31 || d < 1 {
			http.NotFound(w, r)
			return
		}
		postTime, err := time.Parse("2006-01-02", year+"-"+month+"-"+day)
		if err != nil {
			serveError(w, err)
			return
		}
		articleMetaData.PostTime = postTime
		articleMetaData.UpdateTime = postTime.AddDate(0, 0, 1)
		articleMetaData.Title = title
	}
	// get article
	err := articleMetaData.GetOne(ctx)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// get comments
	comment := &Comment{ArticleId: articleMetaData.Id}
	comments, err := comment.GetAll(ctx)
	if err != nil {
		serveError(w, err)
		return
	}

	article := &Article{MetaData: *articleMetaData,
		Comments: comments,
		Text: template.HTML([]byte(blackfriday.
			MarkdownBasic(articleMetaData.Content)))}

	data := make(map[string]interface{})
	data["article"] = article
	endTime := time.Now()
	ctx.Args["costtime"] = endTime.Sub(beginTime)
	ctx.Args["title"] = articleMetaData.Title
	data["args"] = ctx.Args
	viewTPL.ExecuteTemplate(w, "main", data)

}

func SaveCommentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	comment := &Comment{}
	comment.ArticleId = r.FormValue("articleId")
	comment.ParentId = r.FormValue("parentId")
	comment.Author = r.FormValue("name")
	comment.Email = r.FormValue("email")
	comment.Website = r.FormValue("website")
	comment.Flag = 2 //public
	comment.Content = r.FormValue("content")
	now := time.Now()
	comment.PostTime = now
	err := comment.Save(ctx)
	if err != nil {
		serveError(w, err)
	}
	urlStr := r.FormValue("urlStr")
	http.Redirect(w, r, urlStr, http.StatusFound)
}

func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	id := r.FormValue("id")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	comment := new(Comment)
	fmt.Println(id)
	comment.Id = id
	err := comment.Delete(ctx)
	if err != nil {
		serveError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/comment", http.StatusFound)
}

func ListCommentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	beginTime := time.Now()
	if r.FormValue("size") != "" {
		size, err := strconv.Atoi(r.FormValue("size"))
		if err != nil {
			serveError(w, err)
			return
		}
		ctx.Args["size"] = size
	}
	if r.FormValue("pageSize") != "" {
		pageSize, err := strconv.Atoi(r.FormValue("pageSize"))
		if err != nil {
			serveError(w, err)
			return
		}
		ctx.Args["pageSize"] = pageSize
		if pageSize > 100 {
			http.NotFound(w, r)
			return
		}
	}

	comments, err := GetAllComments(ctx)
	if err != nil {
		serveError(w, err)
		return
	}
	data := make(map[string]interface{})
	data["data"] = comments
	prePageSize := ctx.Args["pageSize"].(int) - 1
	ctx.Args["prePageSize"] = prePageSize
	if prePageSize > 0 {
		ctx.Args["hasPre"] = true
	}
	nextPageSize := ctx.Args["pageSize"].(int) + 1
	ctx.Args["nextPageSize"] = nextPageSize

	ctx.Args["url"] = r.URL.Path
	endTime := time.Now()
	ctx.Args["costtime"] = endTime.Sub(beginTime)
	data["args"] = ctx.Args

	commentTPL.ExecuteTemplate(w, "main", data)
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	beginTime := time.Now()

	if r.FormValue("size") != "" {
		size, err := strconv.Atoi(r.FormValue("size"))
		if err != nil {
			serveError(w, err)
			return
		}
		ctx.Args["size"] = size
	}
	if r.FormValue("pageSize") != "" {
		pageSize, err := strconv.Atoi(r.FormValue("pageSize"))
		if err != nil {
			serveError(w, err)
			return
		}
		ctx.Args["pageSize"] = pageSize
	}

	articleMetaData := &ArticleMetaData{}
	articleMetaDatas, err := articleMetaData.GetAll(ctx)
	//err = fmt.Errorf("format %v ", "get all err")
	if err != nil {
		serveError(w, err)
		return
	}
	if len(articleMetaDatas) == 0 {
		http.NotFound(w, r)
		return
	}
	articles := make([]Article, (len(articleMetaDatas)))
	for key, value := range articleMetaDatas {
		articles[key].MetaData = value
		articles[key].Text = template.HTML(blackfriday.MarkdownBasic([]byte(value.Summary)))
	}
	tags, err := GetAllTag(ctx)
	if err != nil {
		serveError(w, err)
		return
	}

	archives, err := GetAllArchive(ctx)
	if err != nil {
		serveError(w, err)
		return
	}

	indexData := &IndexData{Articles: articles, Tags: tags, Archives: archives}
	data := make(map[string]interface{})
	data["data"] = indexData
	prePageSize := ctx.Args["pageSize"].(int) - 1
	ctx.Args["prePageSize"] = prePageSize
	if prePageSize > 0 {
		ctx.Args["hasPre"] = true
	}
	nextPageSize := ctx.Args["pageSize"].(int) + 1
	ctx.Args["nextPageSize"] = nextPageSize

	ctx.Args["url"] = r.URL.Path

	ctx.Args["isListComments"] = true
	endTime := time.Now()
	ctx.Args["costtime"] = endTime.Sub(beginTime)
	data["args"] = ctx.Args

	adminTPL.ExecuteTemplate(w, "main", data)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)

	beginTime := time.Now()
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.FormValue("size") != "" {
		size, err := strconv.Atoi(r.FormValue("size"))
		if err != nil {
			serveError(w, err)
			return
		}
		ctx.Args["size"] = size
	}
	if r.FormValue("pageSize") != "" {
		pageSize, err := strconv.Atoi(r.FormValue("pageSize"))
		if err != nil {
			serveError(w, err)
			return
		}
		ctx.Args["pageSize"] = pageSize
		if pageSize > 100 {
			http.NotFound(w, r)
			return
		}
	}

	articleMetaData := &ArticleMetaData{}
	articleMetaDatas, err := articleMetaData.GetAll(ctx)
	if err != nil {
		serveError(w, err)
		return
	}
	if len(articleMetaDatas) == 0 {
		http.NotFound(w, r)
		return
	}
	articles := make([]Article, (len(articleMetaDatas)))
	for key, value := range articleMetaDatas {
		articles[key].MetaData = value
		articles[key].Text = template.HTML(blackfriday.MarkdownBasic([]byte(value.Summary)))
	}
	tags, err := GetAllTag(ctx)
	if err != nil {
		serveError(w, err)
		return
	}
	archives, err := GetAllArchive(ctx)
	if err != nil {
		serveError(w, err)
		return
	}

	indexData := &IndexData{Articles: articles, Tags: tags, Archives: archives}
	data := make(map[string]interface{})
	data["data"] = indexData

	prePageSize := ctx.Args["pageSize"].(int) - 1
	ctx.Args["prePageSize"] = prePageSize
	if prePageSize > 0 {
		ctx.Args["hasPre"] = true
	}
	nextPageSize := ctx.Args["pageSize"].(int) + 1
	ctx.Args["nextPageSize"] = nextPageSize

	ctx.Args["url"] = r.URL.Path
	endTime := time.Now()
	ctx.Args["costtime"] = endTime.Sub(beginTime)
	data["args"] = ctx.Args
	indexTPL.ExecuteTemplate(w, "main", data)
}

//list article by tag
func TagHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	beginTime := time.Now()
	params := r.URL.Query()
	tag := params.Get(":tag")
	if r.FormValue("size") != "" {
		size, err := strconv.Atoi(r.FormValue("size"))
		if err != nil {
			serveError(w, err)
			return
		}
		ctx.Args["size"] = size
	}
	if r.FormValue("pageSize") != "" {
		pageSize, err := strconv.Atoi(r.FormValue("pageSize"))
		if err != nil {
			serveError(w, err)
			return
		}
		ctx.Args["pageSize"] = pageSize
		if pageSize > 100 {
			http.NotFound(w, r)
			return
		}
	}

	articleMetaData := &ArticleMetaData{}
	articleMetaDatas, err := articleMetaData.GetAllByTag(ctx, tag)
	if err != nil {
		serveError(w, err)
		return
	}
	if len(articleMetaDatas) == 0 {
		http.NotFound(w, r)
		return
	}
	articles := make([]Article, (len(articleMetaDatas)))
	for key, value := range articleMetaDatas {
		articles[key].MetaData = value
		articles[key].Text = template.HTML(blackfriday.MarkdownBasic([]byte(value.Summary)))
	}
	tags, err := GetAllTag(ctx)
	if err != nil {
		serveError(w, err)
	}

	archives, err := GetAllArchive(ctx)
	if err != nil {
		serveError(w, err)
		return
	}

	indexData := &IndexData{Articles: articles, Tags: tags, Archives: archives}
	data := make(map[string]interface{})
	data["data"] = indexData
	prePageSize := ctx.Args["pageSize"].(int) - 1
	ctx.Args["prePageSize"] = prePageSize
	if prePageSize > 0 {
		ctx.Args["hasPre"] = true
	}
	nextPageSize := ctx.Args["pageSize"].(int) + 1
	ctx.Args["nextPageSize"] = nextPageSize

	ctx.Args["url"] = r.URL.Path

	ctx.Args["title"] = string("Tags " + tag + " - Wendyeq blog")
	endTime := time.Now()
	ctx.Args["costtime"] = endTime.Sub(beginTime)

	data["args"] = ctx.Args
	indexTPL.ExecuteTemplate(w, "main", data)
}

//show archive
func ArchiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	beginTime := time.Now()
	//archive := r.URL.Path[len("/blog/archive/"):]
	params := r.URL.Query()
	year := params.Get(":year")
	month := params.Get(":month")
	if r.FormValue("size") != "" {
		size, err := strconv.Atoi(r.FormValue("size"))
		if err != nil {
			serveError(w, err)
			return
		}
		ctx.Args["size"] = size
	}
	if r.FormValue("pageSize") != "" {
		pageSize, err := strconv.Atoi(r.FormValue("pageSize"))
		if err != nil {
			serveError(w, err)
			return
		}
		ctx.Args["pageSize"] = pageSize
		if pageSize > 100 {
			http.NotFound(w, r)
			return
		}
	}

	articleMetaData := ArticleMetaData{}
	articleMetaDatas, err := articleMetaData.GetAllByArchive(ctx, year, month)
	if err != nil {
		serveError(w, err)
		return
	}
	if len(articleMetaDatas) == 0 {
		http.NotFound(w, r)
		return
	}
	articles := make([]Article, (len(articleMetaDatas)))
	for key, value := range articleMetaDatas {
		articles[key].MetaData = value
		articles[key].Text = template.HTML(blackfriday.MarkdownBasic([]byte(value.Summary)))
	}
	tags, err := GetAllTag(ctx)
	if err != nil {
		serveError(w, err)
	}
	archives, err := GetAllArchive(ctx)
	if err != nil {
		serveError(w, err)
		return
	}

	indexData := &IndexData{Articles: articles, Tags: tags, Archives: archives}
	data := make(map[string]interface{})
	data["data"] = indexData

	prePageSize := ctx.Args["pageSize"].(int) - 1
	ctx.Args["prePageSize"] = prePageSize
	if prePageSize > 0 {
		ctx.Args["hasPre"] = true
	}
	nextPageSize := ctx.Args["pageSize"].(int) + 1
	ctx.Args["nextPageSize"] = nextPageSize

	ctx.Args["url"] = r.URL.Path
	ctx.Args["title"] = string("Archive " + year + "-" + month + " - Wendyeq home")
	endTime := time.Now()
	ctx.Args["costtime"] = endTime.Sub(beginTime)

	data["args"] = ctx.Args
	indexTPL.ExecuteTemplate(w, "main", data)
}

//show rss
func RssHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	articleMetaData := ArticleMetaData{}
	articleMetaDatas, err := articleMetaData.GetAll(ctx)
	if err != nil {
		serveError(w, err)
		return
	}
	if len(articleMetaDatas) == 0 {
		http.NotFound(w, r)
		return
	}
	articles := make([]Article, (len(articleMetaDatas)))
	for key, value := range articleMetaDatas {
		articles[key].MetaData = value
		articles[key].Text = template.HTML(blackfriday.MarkdownBasic([]byte(value.Content)))
	}
	rssTPL.Execute(w, articles)
}

//show sitemap
func SitemapHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	articleMetaData := ArticleMetaData{}
	articleMetaDatas, err := articleMetaData.GetAll(ctx)
	if err != nil {
		serveError(w, err)
		return
	}
	if len(articleMetaDatas) == 0 {
		http.NotFound(w, r)
		return
	}
	articles := make([]Article, (len(articleMetaDatas)))
	for key, value := range articleMetaDatas {
		articles[key].MetaData = value
		articles[key].Text = template.HTML(blackfriday.MarkdownBasic([]byte(value.Content)))
	}
	sitemapTPL.Execute(w, articles)
}

//show release notes
func ReleaseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext()
	ctx.GAEContext = appengine.NewContext(r)
	beginTime := time.Now()
	buf, err := GetRelease()
	if err != nil {
		serveError(w, fmt.Errorf("Open RELEASE.md fail. Error = %v", err))
		return
	}
	content := template.HTML(blackfriday.MarkdownBasic(buf))

	data := make(map[string]interface{})
	data["data"] = content
	config := ctx.Args
	endTime := time.Now()
	config["costtime"] = endTime.Sub(beginTime)
	data["config"] = config
	config["title"] = "Release Notes"
	data["config"] = config
	releaseTPL.ExecuteTemplate(w, "main", data)
}

func PingServer(w http.ResponseWriter, r *http.Request) {
	urlStr := "http://blogsearch.google.com/ping?" +
		"name=Wendyeq+Blog&url=http://www.wendyeq.me&" +
		"changesURL=http://www.wendyeq.me/rss.xml"
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	resp, err := client.Get(urlStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	//body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Printf("HTTP GET returned status %v", string(body))
}
