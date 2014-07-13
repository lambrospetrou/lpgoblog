package main

import (
	"fmt"
	"github.com/lambrospetrou/lpgoauth"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type FooterStruct struct {
	Year int
}

type HeaderStruct struct {
	Title string
}

type TemplateBundle struct {
	Post   *BPost
	Footer *FooterStruct
	Header *HeaderStruct
}

type TemplateBundleIndex struct {
	Posts  []*BPost
	Footer *FooterStruct
	Header *HeaderStruct
}

const (
	DIR_POSTS_SRC = "posts/pub/"
)

var validPath = regexp.MustCompile("^/(view|edit|save)/([a-zA-Z0-9_-]+)$")

var templates = template.Must(template.ParseFiles(
	"templates/partials/header.html",
	"templates/partials/footer.html",
	"templates/view.html",
	"templates/edit.html",
	"templates/add.html",
	"templates/index.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, o interface{}) {
	// now we can call the correct template by the basename filename
	err := templates.ExecuteTemplate(w, tmpl+".html", o)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// BLOG HANDLERS
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, id string) {
	bp_id, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bp, err := LoadBlogPost(bp_id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	bp.BodyHtml = template.HTML(bp.ContentHtml)
	bundle := &TemplateBundle{
		Post:   bp,
		Footer: &FooterStruct{Year: time.Now().Year()},
		Header: &HeaderStruct{Title: bp.Title},
	}

	renderTemplate(w, "view", bundle)
}

func editHandler(w http.ResponseWriter, r *http.Request, id string) {
	bp_id, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bp, err := LoadBlogPost(bp_id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "edit", bp)
}

func saveHandler(w http.ResponseWriter, r *http.Request, id string) {
	// avoid changing the data with a GET request
	if strings.ToLower(r.Method) == "get" {
		http.Error(w, "Only through the /edit/:id url", http.StatusMethodNotAllowed)
		return
	}

	bp_id, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bp, err := LoadBlogPost(bp_id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	bp.ContentMarkdown = r.FormValue("markdown")
	bp.Title = r.FormValue("title")
	bp.Save()

	http.Redirect(w, r, "/view/"+bp.IdStr(), http.StatusFound)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if strings.ToLower(r.Method) == "get" {
		renderTemplate(w, "add", "")
		return
	} else if strings.ToLower(r.Method) == "post" {
		bp, err := NewBPost()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bp.ContentMarkdown = r.FormValue("markdown")
		bp.Title = r.FormValue("title")
		bp.Save()

		http.Redirect(w, r, "/view/"+bp.IdStr(), http.StatusFound)
		return
	}
	http.Error(w, "Not supported method", http.StatusMethodNotAllowed)
}

// show all posts
func rootHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hi there, I love you %s\n", r.URL.Path[1:])
	posts, err := LoadAllBlogPosts()
	if err != nil {
		http.Error(w, "Could not load blog posts", http.StatusInternalServerError)
		return
	}
	bundle := &TemplateBundleIndex{
		Footer: &FooterStruct{Year: time.Now().Year()},
		Header: &HeaderStruct{Title: "All posts"},
		Posts:  posts,
	}
	renderTemplate(w, "index", bundle)
}

// Checks if the username:password are correct and valid
func isBasicCredValid(user string, pass string) bool {
	body, err := ioutil.ReadFile("sec/users.txt")
	if err != nil {
		fmt.Println("Cannot read users.txt!!!")
		return false
	}
	users := strings.Split(string(body), "\n")
	for _, u := range users {
		utokens := strings.SplitN(u, ":", 2)
		if lpgoauth.SecureCompare(user, utokens[0]) {
			return lpgoauth.SecureCompare(pass, utokens[1])
		}
	}
	return false
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))

	http.HandleFunc("/edit/", lpgoauth.BasicAuthHandler(isBasicCredValid,
		makeHandler(editHandler)))
	http.HandleFunc("/save/", lpgoauth.BasicAuthHandler(isBasicCredValid,
		makeHandler(saveHandler)))

	http.HandleFunc("/add", lpgoauth.BasicAuthHandler(isBasicCredValid, addHandler))

	http.HandleFunc("/all", rootHandler)
	http.HandleFunc("/", rootHandler)

	fs := http.FileServer(http.Dir("static_data"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(":40080", nil)
}
