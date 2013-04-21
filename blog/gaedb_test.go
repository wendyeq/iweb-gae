// +build !appengine
package blog

import (
	"github.com/wendyeq/appenginetesting"
	"strconv"
	"testing"
	"time"
)

func GetArticleForTest() (a *ArticleMetaData) {
	now := time.Now()
	return &ArticleMetaData{Title: "title",
		PostTime: now, UpdateTime: now,
		Tags: []string{"go", "golang", "gae"},
	}
}

// Save, Update, GetOne, Delete 
func TestArticle(t *testing.T) {
	ctx := GetContext()
	c, err := appenginetesting.NewContext(nil)
	defer c.Close()

	if err != nil {
		t.Fatalf("NewContext: %v", err)
	}
	ctx.GAEContext = c
	//test Save
	firstArticle := GetArticleForTest()
	err = firstArticle.Save(ctx)
	if err != nil {
		t.Fatalf("Save Article: %v", err)
	}

	//test GetOne using Id
	secondArticle := &ArticleMetaData{Id: firstArticle.Id}
	err = secondArticle.GetOne(ctx)
	if err != nil {
		t.Fatalf("GetOne Article Using Id : %v", err)
	}
	if secondArticle.PostTime != firstArticle.PostTime {
		t.Fatal("GetOne Article Using Id Error! PostTime isn't equals.")
	}

	//test GetOne using time and title
	secondArticle = &ArticleMetaData{Title: firstArticle.Title,
		PostTime:   firstArticle.PostTime,
		UpdateTime: firstArticle.PostTime.AddDate(0, 0, 1),
	}
	err = secondArticle.GetOne(ctx)
	if err != nil {
		t.Fatalf("GetOne Article Using time and title : %v", err)
	}
	if secondArticle.Id != firstArticle.Id {
		t.Fatal("GetOne Article Using time and title Error! Id isn't equals.")
	}

	//test Update
	secondArticle.UpdateTime = secondArticle.PostTime.AddDate(0, 0, 1)
	threeArticle, err := secondArticle.Update(ctx)
	if err != nil {
		t.Fatalf("Update Article: %v", err)
	}
	diffHours := threeArticle.UpdateTime.Sub(threeArticle.PostTime).Hours()
	if diffHours != 24 {
		t.Fatalf("Upate Article Error! UpdateTime - PostTime != 24 , It's : %v", diffHours)
	}

	//test Delete
	err = threeArticle.Delete(ctx)
	if err != nil {
		t.Fatalf("Delete Article: %v", err)
	}

	//test GetAll
	firstArticle = GetArticleForTest()
	firstArticle.Title = "test1"
	err = firstArticle.Save(ctx)
	if err != nil {
		t.Fatalf("Test GetAll1 Save Article: %v", err)
	}

	firstArticle = GetArticleForTest()
	firstArticle.Title = "test2"
	err = firstArticle.Save(ctx)
	if err != nil {
		t.Fatalf("Test GetAll2 Save Article: %v", err)
	}

	articles, err := firstArticle.GetAll(ctx)
	if err != nil {
		t.Fatalf("Test GetAll: %v", err)
	}
	for i, v := range articles {
		str_i := strconv.Itoa(1 + i)
		if v.Title != "test"+str_i {
			t.Fatalf("Test GetAll fatal, title = %v, index = %v", v.Title, str_i)
		}
	}
}
