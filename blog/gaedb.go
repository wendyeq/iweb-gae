package blog

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"net/http"
	"time"
)

func (this *ArticleMetaData) Save(r *http.Request) (err error) {
	c := appengine.NewContext(r)
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
		}
		_, err = datastore.Put(c, k, this)
		return err
	}, &datastore.TransactionOptions{XG: true})

	return err
}

func (this *ArticleMetaData) Update(r *http.Request) (articleMetaData *ArticleMetaData, err error) {
	articleMetaData = new(ArticleMetaData)
	c := appengine.NewContext(r)
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

func (this *ArticleMetaData) Delete(r *http.Request) (err error) {
	c := appengine.NewContext(r)
	fmt.Println(this.Id)
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

func (this *ArticleMetaData) GetOne(r *http.Request) (err error) {
	c := appengine.NewContext(r)
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

func (this *ArticleMetaData) GetAll(r *http.Request) (articles []ArticleMetaData, err error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Article").Order("-PostTime").Limit(1000)
	_, err = q.GetAll(c, &articles)
	return articles, err
}

func (this *ArticleMetaData) GetAllByTag(r *http.Request, tag string) (articles []ArticleMetaData, err error) {
	/*
		var tags []Tags
		var keys []*datastore.Key
		c := appengine.NewContext(r)
		q := datastore.NewQuery("Tags").Filter("Tag =", tag).Limit(100)
		_, err = q.GetAll(c, &tags)
		if err != nil {
			return articles, err
		}
		for _, value := range tags {
			keys = append(keys, datastore.NewKey(c, "Article", value.ArticleId, 0, nil))
		}
		articles = make([]ArticleMetaData, len(keys))
		//len(articles) must equal len(keys)
		err = datastore.GetMulti(c, keys, articles)
		return articles, err
	*/
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Article").Filter("Tags = ", tag).Limit(1000)
	_, err = q.GetAll(c, &articles)
	return articles, err
}

func (this *ArticleMetaData) GetAllByArchive(r *http.Request, archive string) (articles []ArticleMetaData, err error) {
	year := archive[0:4]
	month := archive[5:]
	fmt.Println("year=" + year)
	fmt.Println("month=" + month)
	if len(month) == 1 {
		month = "0" + month
	}
	beginTime, err := time.Parse("2006-01-02", year+"-"+month+"-01")
	endTime := beginTime.AddDate(0, 1, 0)

	c := appengine.NewContext(r)
	q := datastore.NewQuery("Article").
		Filter("PostTime >=", beginTime).
		Filter("PostTime <", endTime).
		Order("-PostTime").
		Limit(100)

	_, err = q.GetAll(c, &articles)
	return articles, err
}

func GetAllTag(r *http.Request) (m map[string]int, err error) {
	var tags []Tags
	m = make(map[string]int)
	c := appengine.NewContext(r)
	_, err = datastore.NewQuery("Tags").GetAll(c, &tags)
	for _, value := range tags {
		m[value.Tag]++
	}
	return m, err
}

func GetAllArchive(r *http.Request) (m map[NewString]int, err error) {
	var articleMetaData []ArticleMetaData
	m = make(map[NewString]int)
	c := appengine.NewContext(r)
	_, err = datastore.NewQuery("Article").Order("-PostTime").GetAll(c, &articleMetaData)
	for _, value := range articleMetaData {
		timeStr := value.PostTime.Format("2006-01")
		m[NewString(timeStr)]++
	}
	return m, err
}

func (this *Comment) Save(r *http.Request) (err error) {
	c := appengine.NewContext(r)
	uuid, err := GenUUID()
	if err != nil {
		return err
	}
	this.Id = uuid
	k := datastore.NewKey(c, "Comment", uuid, 0, nil)
	_, err = datastore.Put(c, k, this)
	return err
}

func (this *Comment) GetAll(r *http.Request) (comments []Comment, err error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Comment").
		Filter("ArticleId = ", this.ArticleId).Order("PostTime").Limit(10)
	_, err = q.GetAll(c, &comments)
	return comments, err
}
