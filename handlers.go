package main

import (
	"bytes"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yuin/goldmark"
)

const postsPerPage = 6

type PageData struct {
	Title    string
	Desc     string
	Path     string
	Year     int
	Posts    []Post
	Post     *Post
	Content  template.HTML
	Page     int
	NextPage int
	HasMore  bool
	Query    string
	Error    string
	// Admin form
	FormPost   *Post
	FormAction string
	IsEdit     bool
}

func newPageData(title string, r *http.Request) PageData {
	return PageData{
		Title: title,
		Desc:  "ePer7shme Jobs — platformë moderne rekrutimi që lidh kompanitë me kandidatët e duhur në të gjithë Shqipërinë.",
		Path:  r.URL.Path,
		Year:  time.Now().Year(),
	}
}

func loadTemplates() map[string]*template.Template {
	funcs := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"md":  mdToHTML,
	}
	sets := map[string][]string{
		"home":         {"templates/layout.html", "templates/partials/post_cards.html", "templates/home.html"},
		"blog":         {"templates/layout.html", "templates/partials/post_cards.html", "templates/blog.html"},
		"post":         {"templates/layout.html", "templates/post.html"},
		"notfound":     {"templates/layout.html", "templates/notfound.html"},
		"blog_partial": {"templates/partials/post_cards.html", "templates/partials/blog_partial.html"},
		"login":        {"templates/admin/layout.html", "templates/admin/login.html"},
		"dashboard":    {"templates/admin/layout.html", "templates/admin/dashboard.html"},
		"postform":     {"templates/admin/layout.html", "templates/admin/form.html"},
	}
	out := make(map[string]*template.Template, len(sets))
	for name, files := range sets {
		out[name] = template.Must(template.New(name).Funcs(funcs).ParseFiles(files...))
	}
	return out
}

func mdToHTML(src string) template.HTML {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(src), &buf); err != nil {
		return template.HTML("<p>" + template.HTMLEscapeString(src) + "</p>")
	}
	return template.HTML(buf.String())
}

func (s *Server) render(w http.ResponseWriter, name, root string, data PageData) {
	var buf bytes.Buffer
	if err := s.tmpl[name].ExecuteTemplate(&buf, root, data); err != nil {
		log.Printf("gabim në template %q: %v", name, err)
		http.Error(w, "Gabim i brendshëm", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

// ---- Public handlers ----

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	posts, _, err := s.listPosts("", true, 3, 0)
	if err != nil {
		http.Error(w, "Gabim i brendshëm", http.StatusInternalServerError)
		return
	}
	data := newPageData("ePer7shme Jobs — Platformë Rekrutimi", r)
	data.Posts = posts
	s.render(w, "home", "layout", data)
}

func (s *Server) handleBlog(w http.ResponseWriter, r *http.Request) {
	posts, hasMore, err := s.listPosts("", true, postsPerPage, 0)
	if err != nil {
		http.Error(w, "Gabim i brendshëm", http.StatusInternalServerError)
		return
	}
	data := newPageData("Blog — ePer7shme Jobs", r)
	data.Posts = posts
	data.Page = 1
	data.NextPage = 2
	data.HasMore = hasMore
	s.render(w, "blog", "layout", data)
}

func (s *Server) handleBlogPartial(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	posts, hasMore, err := s.listPosts(q, true, postsPerPage, (page-1)*postsPerPage)
	if err != nil {
		http.Error(w, "Gabim i brendshëm", http.StatusInternalServerError)
		return
	}
	data := PageData{Posts: posts, Page: page, NextPage: page + 1, HasMore: hasMore, Query: q}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl["blog_partial"].ExecuteTemplate(w, "blog_partial", data); err != nil {
		log.Printf("gabim në blog_partial: %v", err)
	}
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	post, err := s.getPostBySlug(slug, true)
	if err == sql.ErrNoRows {
		s.handleNotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "Gabim i brendshëm", http.StatusInternalServerError)
		return
	}
	data := newPageData(post.Title+" — ePer7shme Jobs", r)
	if post.Excerpt != "" {
		data.Desc = post.Excerpt
	}
	data.Post = &post
	data.Content = mdToHTML(post.Content)
	s.render(w, "post", "layout", data)
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	data := newPageData("Faqja nuk u gjet — ePer7shme Jobs", r)
	s.render(w, "notfound", "layout", data)
}
