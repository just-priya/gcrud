package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Post struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func PostF() *Post {
	return &Post{}
}

func (p Post) Save(post *Post) error {
	result := DB.Create(post)
	return result.Error
}

func (p Post) Find(id int) (post *Post, err error) {
	err = DB.Model(&Post{}).Where("id=?", id).Scan(&post).Error
	return
}

func (p Post) GetAll() ([]Post, error) {
	var posts []Post
	err := DB.Model(&Post{}).Find(&posts).Error
	return posts, err
}

func init() {
	db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	DB = db
	DB.AutoMigrate(&Post{})
}

func SendErr(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, "Error: "+err.Error())
}

func main() {
	r := mux.NewRouter()

	fs := http.FileServer(http.Dir("./public/"))
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", fs))

	home := template.Must(template.ParseFiles("home.html"))
	create := template.Must(template.ParseFiles("new.html"))
	view := template.Must(template.ParseFiles("view.html"))

	port := flag.String("p", ":8080", "prot")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		posts, err := PostF().GetAll()
		if err != nil {
			SendErr(w, err)
			return
		}
		home.Execute(w, map[string]interface{}{"Title": "home", "posts": posts})
	}).Methods("GET")
	r.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		create.Execute(w, map[string]interface{}{"Title": "new"})
	}).Methods("GET")
	r.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		post := Post{
			Title: r.FormValue("title"),
			Body:  r.FormValue("body"),
		}
		err := post.Save(&post)
		if err != nil {
			SendErr(w, err)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}).Methods("POST")

	r.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		iid, err := strconv.Atoi(id)
		if err != nil {
			SendErr(w, err)
			return
		}
		post, err := PostF().Find(iid)
		if err != nil {
			SendErr(w, err)
			return
		}
		view.Execute(w, post)
	}).Methods("GET")

	log.Fatal(http.ListenAndServe(*port, r))
}
