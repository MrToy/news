package main

import (
	"github.com/gohttp/app"
	"github.com/gohttp/response"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Article struct {
	Id      bson.ObjectId `json:"id" bson:"_id"`
	Title   string        `json:"title"`
	Content string        `json:"content"`
	Date    time.Time     `json:"date"`
}

type Result struct {
	Success bool
	Info    interface{}
}

func main() {
	MongoAddr, MongoDB, MongoCol := os.Getenv("MONGO_ADDR"), os.Getenv("MONGO_DB"), "news"
	if len(MongoAddr) == 0 {
		MongoAddr = "localhost"
	}
	if len(MongoDB) == 0 {
		MongoDB = "test"
	}
	oldSess, err := mgo.Dial(MongoAddr)
	if err != nil {
		panic(err)
	}
	defer oldSess.Close()
	m := app.New()
	m.Post("/", func(w http.ResponseWriter, r *http.Request) {
		title, content := r.FormValue("title"), r.FormValue("content")
		sess := oldSess.Clone()
		defer sess.Close()
		if err := sess.DB(MongoDB).C(MongoCol).Insert(&Article{Id: bson.NewObjectId(), Title: title, Content: content, Date: time.Now()}); err != nil {
			response.JSON(w, &Result{false, err.Error()})
			return
		}
		response.JSON(w, &Result{true, "ok"})
	})
	m.Put("/:id", func(w http.ResponseWriter, r *http.Request) {
		id, title, content := r.FormValue(":id"), r.FormValue("title"), r.FormValue("content")
		if !bson.IsObjectIdHex(id) {
			response.JSON(w, &Result{false, "error id"})
			return
		}
		sess := oldSess.Clone()
		defer sess.Close()
		if err := sess.DB(MongoDB).C(MongoCol).UpdateId(bson.ObjectIdHex(id), bson.M{"$set": bson.M{"title": title, "content": content}}); err != nil {
			response.JSON(w, &Result{false, err.Error()})
			return
		}
		response.JSON(w, &Result{true, "ok"})
	})
	m.Del("/:id", func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue(":id")
		if !bson.IsObjectIdHex(id) {
			response.JSON(w, &Result{false, "error id"})
			return
		}
		sess := oldSess.Clone()
		defer sess.Close()
		if err := sess.DB(MongoDB).C(MongoCol).RemoveId(bson.ObjectIdHex(id)); err != nil {
			response.JSON(w, &Result{false, err.Error()})
			return
		}
		response.JSON(w, &Result{true, "ok"})
	})
	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		skipStr, limitStr := r.FormValue("skip"), r.FormValue("limit")
		skip, _ := strconv.Atoi(skipStr)
		limit, _ := strconv.Atoi(limitStr)
		sess := oldSess.Clone()
		defer sess.Close()
		arr := []Article{}
		sess.DB(MongoDB).C(MongoCol).Find(nil).Select(bson.M{"content": 0}).Skip(skip).Limit(limit).Sort("-date").All(&arr)
		response.JSON(w, &Result{true, arr})
	})
	m.Get("/:id", func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue(":id")
		if !bson.IsObjectIdHex(id) {
			response.JSON(w, &Result{false, "error id"})
			return
		}
		sess := oldSess.Clone()
		defer sess.Close()
		res := Article{}
		if err := sess.DB(MongoDB).C(MongoCol).FindId(bson.ObjectIdHex(id)).One(&res); err != nil {
			response.JSON(w, &Result{false, err.Error()})
			return
		}
		response.JSON(w, &Result{true, res})
	})
	if err := http.ListenAndServe(":80", m); err != nil {
		panic(err)
	}
}
