package blog

import (
	"fmt"
	//htmlTemplate "html/template"
	"html/template"
	"net/http"
	//"strconv"
	"appengine"
	"appengine/urlfetch"
	"github.com/russross/blackfriday"
	"io"
	//"io/ioutil"
	"regexp"
	"strings"
	textTemplate "text/template"
	"time"
)

var (
	themes, _ = GetConfig()["themes"].(string)

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

	rssTPL = textTemplate.Must(textTemplate.ParseFiles(
		"templates/" + themes + "/rss.xml"))

	sitemapTPL = textTemplate.Must(textTemplate.ParseFiles(
		"templates/" + themes + "/sitemap.xml"))

	releaseTPL = template.Must(template.ParseFiles(
		"templates/"+themes+"/release.html",
		"templates/"+themes+"/common/header.html",
		"templates/"+themes+"/common/footer.html"))
)

//func serveError(w http.ResponseWriter, err error) {
//w.WriteHeader(http.StatusInternalServerError)
//w.Header().Set("Content-Type", "text/plain; charset=utf-8")
//io.WriteString(w, "Internal Server Error! "+err.Error())
//c.Errorf("%v", err)
//	editTPL.Execute(w, nil)
//}

func serveError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, "Internal Server Error "+err.Error())
	//c.Errorf("%v", err)
}

func PreViewArticleHandler(w http.ResponseWriter, r *http.Request) {
	articleMetaData := &ArticleMetaData{}
	articleMetaData.Author, _ = GetConfig()["author"].(string)
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

	article := &Article{MetaData: *articleMetaData, Text: template.HTML([]byte(blackfriday.MarkdownBasic(articleMetaData.Content)))}
	data := make(map[string]interface{})
	data["article"] = article
	config := GetConfig()
	config["title"] = articleMetaData.Title
	data["config"] = config
	viewTPL.ExecuteTemplate(w, "main", data)
}

func SaveArticleHandler(w http.ResponseWriter, r *http.Request) {
	article := &ArticleMetaData{}
	article.Author, _ = GetConfig()["author"].(string)
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
	err := article.Save(r)
	if err == nil {
		PingServer(w, r)
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		fmt.Fprint(w, err.Error())
	}
}

func EditArticleHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id != "" {
		article := new(ArticleMetaData)
		article.Id = id
		err := article.GetOne(r)
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
	article, err := article.Update(r)
	if err != nil {
		serveError(w, err)
		return
	}
	PingServer(w, r)
	urlStr := "/blog/" + article.PostTime.Format("2006/01/02") + "/" + article.Title
	http.Redirect(w, r, urlStr, http.StatusFound)

}

func DeleteArticleHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	article := new(ArticleMetaData)
	fmt.Println(id)
	article.Id = id
	err := article.Delete(r)
	if err != nil {
		serveError(w, err)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusFound)

}

var dateTime = regexp.MustCompile("^[0-9]{4}/[0-9]{2}/[0-9]{2}/+")

func ViewArticleHandler(w http.ResponseWriter, r *http.Request) {
	beginTime := time.Now()

	url := r.URL.Path[len("/blog/"):]
	if !dateTime.MatchString(url) {
		http.NotFound(w, r)
		return
	}
	year := url[0:4]
	month := url[5:7]
	day := url[8:10]
	title := url[11:]

	postTime, _ := time.Parse("2006-01-02", year+"-"+month+"-"+day)
	updateTime := postTime.AddDate(0, 0, 1)
	articleMetaData := &ArticleMetaData{Title: title, PostTime: postTime, UpdateTime: updateTime}
	err := articleMetaData.GetOne(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	comment := &Comment{ArticleId: articleMetaData.Id}
	comments, err := comment.GetAll(r)
	if err != nil {
		serveError(w, err)
		return
	}

	//article.Comments = 
	//postTime.Format("2006/01")
	article := &Article{MetaData: *articleMetaData,
		Comments: comments,
		Text:     template.HTML([]byte(blackfriday.MarkdownBasic(articleMetaData.Content)))}

	data := make(map[string]interface{})
	data["article"] = article
	config := GetConfig()
	endTime := time.Now()
	config["costtime"] = endTime.Sub(beginTime)
	config["title"] = articleMetaData.Title
	data["config"] = config
	viewTPL.ExecuteTemplate(w, "main", data)

}

func SaveCommentHandler(w http.ResponseWriter, r *http.Request) {
	comment := &Comment{}
	comment.ArticleId = r.FormValue("articleId")
	comment.ParentId = r.FormValue("parentId")
	comment.Author = r.FormValue("name")
	comment.Email = r.FormValue("email")
	comment.Website = r.FormValue("website")
	comment.Flag = 2
	comment.Content = r.FormValue("content")
	now := time.Now()
	comment.PostTime = now
	err := comment.Save(r)
	if err != nil {
		serveError(w, err)
	}
	urlStr := r.FormValue("urlStr")
	http.Redirect(w, r, urlStr, http.StatusFound)
}
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	beginTime := time.Now()
	articleMetaData := &ArticleMetaData{}
	articleMetaDatas, err := articleMetaData.GetAll(r)
	//err = fmt.Errorf("format %v ", "get all err")
	if err != nil {
		serveError(w, err)
		return
	}
	articles := make([]Article, (len(articleMetaDatas)))
	for key, value := range articleMetaDatas {
		articles[key].MetaData = value
		articles[key].Text = template.HTML(blackfriday.MarkdownBasic([]byte(value.Content)))
	}
	tags, err := GetAllTag(r)
	if err != nil {
		serveError(w, err)
		return
	}

	archives, err := GetAllArchive(r)
	if err != nil {
		serveError(w, err)
		return
	}

	indexData := &IndexData{Articles: articles, Tags: tags, Archives: archives}
	data := make(map[string]interface{})
	data["data"] = indexData
	config := GetConfig()
	endTime := time.Now()
	config["costtime"] = endTime.Sub(beginTime)
	data["config"] = config
	adminTPL.ExecuteTemplate(w, "main", data)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	beginTime := time.Now()

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	articleMetaData := &ArticleMetaData{}
	articleMetaDatas, err := articleMetaData.GetAll(r)
	//err = fmt.Errorf("format %v ", "get all err")
	if err != nil {
		serveError(w, err)
		return
	}
	articles := make([]Article, (len(articleMetaDatas)))
	for key, value := range articleMetaDatas {
		articles[key].MetaData = value
		articles[key].Text = template.HTML(blackfriday.MarkdownBasic([]byte(value.Content)))
	}
	tags, err := GetAllTag(r)
	if err != nil {
		serveError(w, err)
		return
	}
	archives, err := GetAllArchive(r)
	if err != nil {
		serveError(w, err)
		return
	}

	indexData := &IndexData{Articles: articles, Tags: tags, Archives: archives}
	data := make(map[string]interface{})
	data["data"] = indexData
	config := GetConfig()
	endTime := time.Now()
	config["costtime"] = endTime.Sub(beginTime)
	data["config"] = config
	indexTPL.ExecuteTemplate(w, "main", data)
}

//show tag
func TagHandler(w http.ResponseWriter, r *http.Request) {
	beginTime := time.Now()
	tag := r.URL.Path[len("/blog/tag/"):]
	articleMetaData := &ArticleMetaData{}
	articleMetaDatas, err := articleMetaData.GetAllByTag(r, tag)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	articles := make([]Article, (len(articleMetaDatas)))
	for key, value := range articleMetaDatas {
		articles[key].MetaData = value
		articles[key].Text = template.HTML(blackfriday.MarkdownBasic([]byte(value.Content)))
	}
	tags, err := GetAllTag(r)
	if err != nil {
		serveError(w, err)
	}

	archives, err := GetAllArchive(r)
	if err != nil {
		serveError(w, err)
		return
	}

	indexData := &IndexData{Articles: articles, Tags: tags, Archives: archives}
	data := make(map[string]interface{})
	data["data"] = indexData
	config := GetConfig()
	endTime := time.Now()
	config["costtime"] = endTime.Sub(beginTime)
	config["title"] = string("Tags " + tag + ", Article list")
	data["config"] = config
	indexTPL.ExecuteTemplate(w, "main", data)
}

//show archive
func ArchiveHandler(w http.ResponseWriter, r *http.Request) {
	beginTime := time.Now()
	archive := r.URL.Path[len("/blog/archive/"):]
	articleMetaData := ArticleMetaData{}
	articleMetaDatas, err := articleMetaData.GetAllByArchive(r, archive)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	articles := make([]Article, (len(articleMetaDatas)))
	for key, value := range articleMetaDatas {
		articles[key].MetaData = value
		articles[key].Text = template.HTML(blackfriday.MarkdownBasic([]byte(value.Content)))
	}
	tags, err := GetAllTag(r)
	if err != nil {
		serveError(w, err)
	}
	archives, err := GetAllArchive(r)
	if err != nil {
		serveError(w, err)
		return
	}

	indexData := &IndexData{Articles: articles, Tags: tags, Archives: archives}

	data := make(map[string]interface{})
	data["data"] = indexData
	config := GetConfig()
	endTime := time.Now()
	config["costtime"] = endTime.Sub(beginTime)
	data["config"] = config
	config["title"] = string("Archive " + archive + ", Article list")
	data["config"] = config
	indexTPL.ExecuteTemplate(w, "main", data)
}

//show rss
func RssHandler(w http.ResponseWriter, r *http.Request) {
	articleMetaData := ArticleMetaData{}
	articleMetaDatas, err := articleMetaData.GetAll(r)
	if err != nil {
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
	articleMetaData := ArticleMetaData{}
	articleMetaDatas, err := articleMetaData.GetAll(r)
	if err != nil {
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
	beginTime := time.Now()
	buf, err := GetRelease()
	if err != nil {
		serveError(w, fmt.Errorf("Open RELEASE.md fail. Error = %v", err))
		return
	}
	content := template.HTML(blackfriday.MarkdownBasic(buf))

	data := make(map[string]interface{})
	data["data"] = content
	config := GetConfig()
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
