package main

import (
	"encoding/json"
	"errors"
	"github.com/lambrospetrou/lpgoblog/lpdb"
	"github.com/russross/blackfriday"
	"html/template"
	"strconv"
	"time"
)

type BPost struct {
	Id                 int           `json:"id"`
	Title              string        `json:"title"`
	Author             string        `json:"author"`
	DateCreated        time.Time     `json:"date_created"`
	UrlFriendlyLink    string        `json:"url_friendly_link"`
	ContentMarkdown    string        `json:"content_markdown"`
	DateEditedMarkdown time.Time     `json:"date_edited_markdown"`
	ContentHtml        string        `json:"content_html"`
	DateCompiledHtml   time.Time     `json:"date_edited_html"`
	BodyHtml           template.HTML `json:"-"`
}

func (p *BPost) IdStr() string {
	return strconv.Itoa(p.Id)
}

func (p *BPost) Save() error {
	// update the HTML content
	p.DateEditedMarkdown = time.Now()
	p.ContentHtml = string(blackfriday.MarkdownCommon([]byte(p.ContentMarkdown)))
	p.DateCompiledHtml = p.DateEditedMarkdown

	// store into the database
	db, err := lpdb.CDBInstance()
	if err != nil {
		return errors.New("Could not get instance of Couchbase")
	}
	jsonBytes, err := json.Marshal(p)
	if err != nil {
		return errors.New("Could not convert post to JSON format!")
	}
	return db.SetRaw("bp::"+p.IdStr(), 0, jsonBytes)
}

/////////////////////////////////////////////////////
////////////////// GENERAL FUNCTIONS
/////////////////////////////////////////////////////

func LoadBlogPost(id int) (*BPost, error) {
	p := &BPost{}
	db, err := lpdb.CDBInstance()
	if err != nil {
		return nil, errors.New("Could not get instance of Couchbase")
	}
	err = db.Get("bp::"+strconv.Itoa(id), &p)
	return p, err
}

// Creates a new blog post with auto-incremented key and returns it empty
func NewBPost() (*BPost, error) {
	p := &BPost{}
	db, err := lpdb.CDBInstance()
	if err != nil {
		return nil, errors.New("Could not get instance of Couchbase")
	}
	bp_id, err := db.FAI("bp::count")
	p.Id = int(bp_id)
	// update created time
	p.DateCreated = time.Now()
	return p, err
}
