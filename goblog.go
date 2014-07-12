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
)

const (
	DIR_POSTS_SRC = "posts/pub/"
)

var validPath = regexp.MustCompile("^/(view|edit|save)/([a-zA-Z0-9_-]+)$")
var templates = template.Must(template.ParseFiles(
	"templates/partials/header.html",
	"templates/partials/footer.html",
	"templates/view.html",
	"templates/edit.html",
	"templates/login.html"))

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

	renderTemplate(w, "view", bp)
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

	http.Redirect(w, r, "/view/"+string(id), http.StatusFound)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love you %s\n", r.URL.Path[1:])
}

// displays the login form
func loginHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login", "")
}

// try to extract the credentials first from the header and then from the FormValue
func authenticate(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	fmt.Println(user, ":", pass)
	fmt.Println(r.Header.Get("Authorization"))
	fmt.Println(r.Header.Get("Content-type"))

	w.Header().Set("Authorization", "Basic realm=\"Login credentials\"")
	http.Error(w, "Not Authorized", http.StatusUnauthorized)
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
		//if user == utokens[0] {
		//	return pass == utokens[1]
		//}
	}
	return false
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))

	http.HandleFunc("/edit/", lpgoauth.BasicAuthHandler(isBasicCredValid,
		makeHandler(editHandler)))
	http.HandleFunc("/save/", lpgoauth.BasicAuthHandler(isBasicCredValid,
		makeHandler(saveHandler)))

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/auth", authenticate)

	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":8080", nil)
}
