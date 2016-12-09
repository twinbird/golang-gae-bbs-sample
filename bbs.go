package bbs

import (
	"fmt"
	"net/http"
	"time"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine"
	"strconv"
	"html/template"
	"golang.org/x/net/context"
)

var bbsTemplate *template.Template

func init() {
	http.HandleFunc("/index", indexHandler)
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/like", likeHandler)
	bbsTemplate = template.Must(template.ParseFiles("bbs.tmpl"))
}

type Comment struct {
	Id int64 `datastore:"-"`
	HandleName string
	Comment string
	Like int64
	EntryTime time.Time
}

func (c *Comment) EntryTimeView() string {
	return c.EntryTime.Format("2006-01-02 15:04:05")
}

func allComments(c context.Context) ([]Comment, error) {
	q := datastore.NewQuery("comment").Order("-EntryTime")

	comments := make([]Comment, 0)
	iter := q.Run(c)
	for {
		var com Comment
		key, err := iter.Next(&com)
		if err == datastore.Done {
			break
		} else if err != nil {
			return nil, err
		}
		com.Id = key.IntID()
		comments = append(comments, com)
	}
	return comments, nil
}

func (com *Comment) addComment(c context.Context) error {
	key := datastore.NewIncompleteKey(c, "comment", nil)
	key, err := datastore.Put(c, key, com)
	if err != nil {
		return err
	}
	return nil
}

func removeComment(c context.Context, keyStr string) error {
	keyInt, err := strconv.ParseInt(keyStr, 10, 64)
	if err != nil {
		return err
	}
	key := datastore.NewKey(c, "comment", "", keyInt, nil)

	q := datastore.NewQuery("comment").Filter("__key__=", key).KeysOnly()
	iter := q.Run(c)
	var com Comment
	_, err = iter.Next(&com)
	if err == datastore.Done {
		return fmt.Errorf("key not found")
	}
	if err != nil {
		return err
	}
	if err = datastore.Delete(c, key); err != nil {
		return err
	}
	return nil
}

func likeComment(c context.Context, keyStr string) error {
	keyInt, err := strconv.ParseInt(keyStr, 10, 64)
	if err != nil {
		return err
	}
	key := datastore.NewKey(c, "comment", "", keyInt, nil)

	var com Comment
	err = datastore.Get(c, key, &com)
	if err != nil {
		return err
	}

	com.Like += 1

	if _, err = datastore.Put(c, key, &com); err != nil {
		return err
	}
	return nil
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	comment := Comment {
		HandleName: r.FormValue("handleName"),
		Comment: r.FormValue("comment"),
		EntryTime: time.Now(),
	}
	if err := comment.addComment(c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	comments, err := allComments(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := bbsTemplate.Execute(w, comments); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	keyStr := r.FormValue("key")
	removeComment(c, keyStr)

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func likeHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	keyStr := r.FormValue("key")
	likeComment(c, keyStr)

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}
