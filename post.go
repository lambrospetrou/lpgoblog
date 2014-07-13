package main

import (
	"encoding/json"
	"errors"
	"github.com/lambrospetrou/lpgoblog/lpdb"
	"github.com/russross/blackfriday"
	"html/template"
	"sort"
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

// ByAge implements sort.Interface for []Person based on
// the Age field.
type ByDate []*BPost

func (a ByDate) Len() int      { return len(a) }
func (a ByDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool {
	if a[i].DateCreated.Unix() > a[j].DateCreated.Unix() {
		return true
	}
	if a[i].DateCreated.Unix() < a[j].DateCreated.Unix() {
		return false
	}
	return a[i].DateEditedMarkdown.Unix() >= a[j].DateEditedMarkdown.Unix()
}

func (p *BPost) IdStr() string {
	return strconv.Itoa(p.Id)
}

func (p *BPost) FormattedEditedTime() string {
	return p.DateEditedMarkdown.Format("January 02, 2006 | Monday -- 15:04PM")
}

func (p *BPost) FormattedCreatedTime() string {
	return p.DateCreated.Format("January 02, 2006 | Monday")
}

func (p *BPost) HTML5CreatedTime() string {
	return p.DateCreated.Format("2006-01-02")
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

func (p *BPost) Del() error {
	// store into the database
	db, err := lpdb.CDBInstance()
	if err != nil {
		return errors.New("Could not get instance of Couchbase")
	}
	return db.Delete("bp::" + p.IdStr())
}

/////////////////////////////////////////////////////
////////////////// GENERAL FUNCTIONS
/////////////////////////////////////////////////////

func LoadAllBlogPosts() ([]*BPost, error) {
	db, err := lpdb.CDBInstance()
	if err != nil {
		return nil, errors.New("Could not get instance of Couchbase")
	}
	var count int
	err = db.Get("bp::count", &count)
	if err != nil {
		return nil, errors.New("Could not get number of blog posts!")
	}
	// allocate space for all the posts (start from 1 and inclusive count)
	keys := make([]string, count+1)
	for i := 1; i <= count; i++ {
		keys[i] = "bp::" + strconv.Itoa(i)
	}
	postsMap, err := db.GetBulk(keys)
	if err != nil {
		return nil, errors.New("Could not get blog posts!")
	}
	var posts []*BPost = make([]*BPost, count)
	count = 0
	for _, v := range postsMap {
		bp := &BPost{}
		err = json.Unmarshal(v.Body, bp)
		if err == nil {
			posts[count] = bp
			count++
		}
	}
	// we take only a part of the slice since there might were deleted posts
	// and their id returned nothing with the bulk get.
	posts = posts[:count]
	sort.Sort(ByDate(posts))
	return posts, nil
}

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
