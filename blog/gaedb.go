// +build appengine
package blog

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"time"
)

//save article and save tags transaction
func (this *ArticleMetaData) Save(ctx Context) (err error) {
	c := ctx.GAEContext
	uuid, err := GenUUID()
	if err != nil {
		return err
	}
	this.Id = uuid
	k := datastore.NewKey(c, "Article", uuid, 0, nil)

	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		if len(this.Tags) > 0 {
			tags := make([]Tags, len(this.Tags))
			tagsKey := make([]*datastore.Key, len(this.Tags))
			for id, tag := range this.Tags {
				tags[id].ArticleId = uuid
				tags[id].Tag = tag
				tagId := uuid + tag
				tagsKey[id] = datastore.NewKey(c, "Tags", tagId, 0, nil)
			}
			_, err = datastore.PutMulti(c, tagsKey, tags)
			if err != nil {
				return err
			}
		}
		_, err = datastore.Put(c, k, this)
		return err

	}, &datastore.TransactionOptions{XG: true})

	return err
}

//update article ,
func (this *ArticleMetaData) Update(ctx Context) (articleMetaData *ArticleMetaData, err error) {
	c := ctx.GAEContext
	articleMetaData = new(ArticleMetaData)
	k := datastore.NewKey(c, "Article", this.Id, 0, nil)
	err = datastore.Get(c, k, articleMetaData)
	if err != nil {
		return articleMetaData, err
	}
	articleMetaData.Tags = this.Tags
	articleMetaData.Summary = this.Summary
	articleMetaData.Content = this.Content
	articleMetaData.UpdateTime = this.UpdateTime
	_, err = datastore.Put(c, k, articleMetaData)
	return articleMetaData, err
}

func (this *ArticleMetaData) Delete(ctx Context) (err error) {
	c := ctx.GAEContext
	k := datastore.NewKey(c, "Article", this.Id, 0, nil)
	err = datastore.RunInTransaction(c, func(c appengine.Context) error {
		err = datastore.Get(c, k, this)
		if err != nil {
			return err
		}
		if len(this.Tags) > 0 {
			tags := make([]Tags, len(this.Tags))
			tagsKey := make([]*datastore.Key, len(this.Tags))
			for id, tag := range this.Tags {
				tags[id].ArticleId = this.Id
				tags[id].Tag = tag
				tagId := this.Id + tag
				tagsKey[id] = datastore.NewKey(c, "Tags", tagId, 0, nil)
			}
			err = datastore.DeleteMulti(c, tagsKey)
			if err != nil {
				return err
			}
		}
		err = datastore.Delete(c, k)
		return err
	}, &datastore.TransactionOptions{XG: true})
	return err
}

func (this *ArticleMetaData) GetOne(ctx Context) (err error) {
	c := ctx.GAEContext
	if len(this.Id) > 0 {
		err = datastore.RunInTransaction(c, func(c appengine.Context) error {
			k := datastore.NewKey(c, "Article", this.Id, 0, nil)
			err = datastore.Get(c, k, this)
			if err != nil && err != datastore.ErrNoSuchEntity {
				return err
			}
			this.Count++
			_, err = datastore.Put(c, k, this)
			return err
		}, nil)
	} else {
		var articles []ArticleMetaData
		q := datastore.NewQuery("Article").
			Filter("PostTime >=", this.PostTime).
			Filter("PostTime <=", this.UpdateTime).
			Filter("Title =", this.Title).
			Limit(1)

		_, err = q.GetAll(c, &articles)
		if err != nil && err != datastore.ErrNoSuchEntity && len(articles) == 0 {
			return err
		}

		if len(articles) == 0 {
			return fmt.Errorf("Fail to find article title=%v: %v", this.Title, err)
		}
		*this = articles[0]
		this.Count++
		k := datastore.NewKey(c, "Article", this.Id, 0, nil)
		_, err = datastore.Put(c, k, this)
		return err
	}
	return err
}

func (this *ArticleMetaData) GetAll(ctx Context) (articles []ArticleMetaData, err error) {
	c := ctx.GAEContext
	size, ok := ctx.Args["size"].(int)
	if !ok || size <= 0 {
		size = 5
		ctx.Args["size"] = size
	}
	pageSize, ok := ctx.Args["pageSize"].(int)
	if !ok || pageSize <= 0 {
		pageSize = 1
		ctx.Args["pageSize"] = pageSize
	}
	offset := size * (pageSize - 1)

	q := datastore.NewQuery("Article").Order("-PostTime").Offset(offset).Limit(size)
	_, err = q.GetAll(c, &articles)
	return articles, err
}

func (this *ArticleMetaData) GetAllByTag(ctx Context, tag string) (articles []ArticleMetaData, err error) {
	c := ctx.GAEContext

	size, ok := ctx.Args["size"].(int)
	if !ok || size <= 0 {
		size = 5
		ctx.Args["size"] = size
	}
	fmt.Println("size tag : ", size)
	pageSize, ok := ctx.Args["pageSize"].(int)
	if !ok || pageSize <= 0 {
		pageSize = 1
		ctx.Args["pageSize"] = pageSize
	}
	offset := size * (pageSize - 1)

	q := datastore.NewQuery("Article").Filter("Tags = ", tag).Offset(offset).Limit(size)
	_, err = q.GetAll(c, &articles)
	return articles, err
}

func (this *ArticleMetaData) GetAllByArchive(ctx Context, archive string) (articles []ArticleMetaData, err error) {
	c := ctx.GAEContext

	year := archive[0:4]
	month := archive[5:]
	fmt.Println("year=" + year)
	fmt.Println("month=" + month)
	if len(month) == 1 {
		month = "0" + month
	}
	beginTime, err := time.Parse("2006-01-02", year+"-"+month+"-01")
	endTime := beginTime.AddDate(0, 1, 0)

	size, ok := ctx.Args["size"].(int)
	if !ok || size <= 0 {
		size = 5
		ctx.Args["size"] = size
	}
	pageSize, ok := ctx.Args["pageSize"].(int)
	if !ok || pageSize <= 0 {
		pageSize = 1
		ctx.Args["pageSize"] = pageSize
	}
	offset := size * (pageSize - 1)

	q := datastore.NewQuery("Article").
		Filter("PostTime >=", beginTime).
		Filter("PostTime <", endTime).
		Order("-PostTime").
		Offset(offset).
		Limit(size)

	_, err = q.GetAll(c, &articles)
	return articles, err
}

func GetAllTag(ctx Context) (m map[string]int, err error) {
	c := ctx.GAEContext
	var tags []Tags
	m = make(map[string]int)
	_, err = datastore.NewQuery("Tags").GetAll(c, &tags)
	for _, value := range tags {
		m[value.Tag]++
	}
	return m, err
}

func GetAllArchive(ctx Context) (m map[NewString]int, err error) {
	c := ctx.GAEContext
	var articleMetaData []ArticleMetaData
	m = make(map[NewString]int)
	_, err = datastore.NewQuery("Article").Order("-PostTime").GetAll(c, &articleMetaData)
	for _, value := range articleMetaData {
		timeStr := value.PostTime.Format("2006-01")
		m[NewString(timeStr)]++
	}
	return m, err
}

func (this *Comment) Save(ctx Context) (err error) {
	c := ctx.GAEContext
	uuid, err := GenUUID()
	if err != nil {
		return err
	}
	this.Id = uuid
	k := datastore.NewKey(c, "Comment", uuid, 0, nil)
	_, err = datastore.Put(c, k, this)
	return err
}

func (this *Comment) Delete(ctx Context) (err error) {
	c := ctx.GAEContext
	k := datastore.NewKey(c, "Comment", this.Id, 0, nil)
	err = datastore.Delete(c, k)
	return err
}

func (this *Comment) GetAll(ctx Context) (comments []Comment, err error) {
	c := ctx.GAEContext
	q := datastore.NewQuery("Comment").
		Filter("ArticleId = ", this.ArticleId).Order("PostTime").Limit(10)
	_, err = q.GetAll(c, &comments)
	return comments, err
}

func GetAllComments(ctx Context) (comments []Comment, err error) {
	c := ctx.GAEContext
	q := datastore.NewQuery("Comment").Order("-PostTime").Limit(100)
	_, err = q.GetAll(c, &comments)
	return comments, err
}
