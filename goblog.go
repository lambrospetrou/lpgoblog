package main

import (
	"fmt"
	//"github.com/russross/blackfriday"
	"html/template"
	//"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strings"
)

const (
	DIR_POSTS_SRC = "posts/pub/"
)

var validPath = regexp.MustCompile("^/(view|edit|save)/([a-zA-Z0-9_-]+)$")
var templates = template.Must(template.ParseFiles("templates/partials/header.html",
	"templates/partials/footer.html", "templates/view.html", "templates/edit.html",
	"templates/login.html"))

/*
func renderTemplate(w http.ResponseWriter, tmpl string, p *BPost) {
	// now we can call the correct template by the basename filename
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
*/
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

// BasicRealm is used when setting the WWW-Authenticate response header.
var BasicRealm = "Authorization Required"

// SecureCompare performs a constant time compare of two strings to limit timing attacks.
func SecureCompare(given string, actual string) bool {
	givenSha := sha256.Sum256([]byte(given))
	actualSha := sha256.Sum256([]byte(actual))

	return subtle.ConstantTimeCompare(givenSha[:], actualSha[:]) == 1
}

func authenticate(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	fmt.Println(user, ":", pass)
	fmt.Println(r.Header.Get("Authorization"))
	fmt.Println(r.Header.Get("Content-type"))

	w.Header().Set("Authorization", "Basic realm=\"Login credentials\"")
	http.Error(w, "Not Authorized", http.StatusUnauthorized)
}

// THIS WILL DISPLAY THE CREDENTIALS BOX IN THE BROWSERS
func makeBasicAuthHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// authenticate the token
		auth := r.Header.Get("Authorization")
		if len(auth) < 6 || auth[:6] != "Basic " {
			rejectAuthorization(w)
			return
		}
		b, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			rejectAuthorization(w)
			return
		}
		tokens := strings.SplitN(string(b), ":", 2)
		if len(tokens) != 2 || !isAuthValid(tokens[0], tokens[1]) {
			rejectAuthorization(w)
			return
		}

		// delegate the call
		fn(w, r)
	}
}

func isAuthValid(user string, pass string) bool {
	return user == "lambros" && pass == "petrou"
}

func rejectAuthorization(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=\""+BasicRealm+"\"")
	http.Error(w, "Not Authorized", http.StatusUnauthorized)
}

func main() {
	/*
		p, err := NewBPost()
		if err != nil {
			fmt.Errorf("error while creating new empty blog post", err.Error())
			return
		}
		p.Author = "Lambros Petrou"
		p.ContentMarkdown = "	paragraph 1"
		p.Save()

		p, err = LoadBlogPost(p.Id)
		if err != nil {
			fmt.Println("error " + err.Error())
		}
		fmt.Println(p.IdStr(), p.Author, p.ContentMarkdown, p.DateCreated)

		fmt.Println("Finished script!")

		markdownContent, err := ioutil.ReadFile("articles/hello.md")
		if err != nil {
			fmt.Println("error while reading markdown hello.md")
		}
		htmlContent := blackfriday.MarkdownCommon(markdownContent)

		fmt.Println("\n\n", string(htmlContent))

		p.ContentMarkdown = string(markdownContent)
		p.Save()
		fmt.Println("updated content of the markdown")
	*/
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeBasicAuthHandler(makeHandler(editHandler)))
	http.HandleFunc("/save/", makeBasicAuthHandler(makeHandler(saveHandler)))

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/auth", authenticate)

	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":8080", nil)
}
