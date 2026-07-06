package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.validSession(r) {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/admin/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if s.validSession(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	data := newPageData("Hyrje — Paneli i Administrimit", r)
	s.render(w, "login", "layout", data)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	fail := func() {
		data := newPageData("Hyrje — Paneli i Administrimit", r)
		data.Error = "Emri i përdoruesit ose fjalëkalimi është i pasaktë."
		w.WriteHeader(http.StatusUnauthorized)
		s.render(w, "login", "layout", data)
	}

	user, err := s.getUserByUsername(username)
	if err != nil {
		fail()
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		fail()
		return
	}

	token, err := s.createSession(user.ID)
	if err != nil {
		http.Error(w, "Gabim i brendshëm", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.destroySession(r)
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	posts, _, err := s.listPosts("", false, 500, 0)
	if err != nil {
		http.Error(w, "Gabim i brendshëm", http.StatusInternalServerError)
		return
	}
	data := newPageData("Paneli — ePer7shme Jobs", r)
	data.Posts = posts
	s.render(w, "dashboard", "layout", data)
}

func (s *Server) handleNewPost(w http.ResponseWriter, r *http.Request) {
	data := newPageData("Postim i ri — Paneli", r)
	data.FormPost = &Post{Published: true}
	data.FormAction = "/admin/posts"
	s.render(w, "postform", "layout", data)
}

func postFromForm(r *http.Request) Post {
	return Post{
		Title:     strings.TrimSpace(r.FormValue("title")),
		Slug:      strings.TrimSpace(r.FormValue("slug")),
		Excerpt:   strings.TrimSpace(r.FormValue("excerpt")),
		Content:   r.FormValue("content"),
		Cover:     strings.TrimSpace(r.FormValue("cover")),
		Published: r.FormValue("published") == "on",
	}
}

func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	p := postFromForm(r)
	if p.Title == "" {
		s.renderFormError(w, r, &p, "/admin/posts", false, "Titulli është i detyrueshëm.")
		return
	}
	if p.Slug == "" {
		p.Slug = slugify(p.Title)
	} else {
		p.Slug = slugify(p.Slug)
	}
	p.Slug = s.uniqueSlug(p.Slug, 0)
	if err := s.createPost(p); err != nil {
		http.Error(w, "Gabim i brendshëm", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (s *Server) handleEditPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := s.getPostByID(id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "Gabim i brendshëm", http.StatusInternalServerError)
		return
	}
	data := newPageData("Ndrysho postimin — Paneli", r)
	data.FormPost = &post
	data.FormAction = "/admin/posts/" + strconv.FormatInt(id, 10)
	data.IsEdit = true
	s.render(w, "postform", "layout", data)
}

func (s *Server) handleUpdatePost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	existing, err := s.getPostByID(id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "Gabim i brendshëm", http.StatusInternalServerError)
		return
	}

	p := postFromForm(r)
	p.ID = id
	p.CreatedAt = existing.CreatedAt
	if p.Title == "" {
		s.renderFormError(w, r, &p, "/admin/posts/"+strconv.FormatInt(id, 10), true, "Titulli është i detyrueshëm.")
		return
	}
	if p.Slug == "" {
		p.Slug = slugify(p.Title)
	} else {
		p.Slug = slugify(p.Slug)
	}
	p.Slug = s.uniqueSlug(p.Slug, id)
	if err := s.updatePost(p); err != nil {
		http.Error(w, "Gabim i brendshëm", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (s *Server) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := s.deletePost(id); err != nil {
		http.Error(w, "Gabim i brendshëm", http.StatusInternalServerError)
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK) // HTMX removes the row via hx-swap
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (s *Server) renderFormError(w http.ResponseWriter, r *http.Request, p *Post, action string, isEdit bool, msg string) {
	data := newPageData("Postim — Paneli", r)
	data.FormPost = p
	data.FormAction = action
	data.IsEdit = isEdit
	data.Error = msg
	w.WriteHeader(http.StatusUnprocessableEntity)
	s.render(w, "postform", "layout", data)
}
