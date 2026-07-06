package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
)

type Server struct {
	db   *sql.DB
	tmpl map[string]*template.Template
}

func main() {
	db, err := openDB("eper7shme.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := migrate(db); err != nil {
		log.Fatal(err)
	}
	if err := seed(db); err != nil {
		log.Fatal(err)
	}

	s := &Server{db: db, tmpl: loadTemplates()}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Public
	mux.HandleFunc("GET /{$}", s.handleHome)
	mux.HandleFunc("GET /blog", s.handleBlog)
	mux.HandleFunc("GET /blog/partial", s.handleBlogPartial)
	mux.HandleFunc("GET /blog/{slug}", s.handlePost)

	// Admin
	mux.HandleFunc("GET /admin/login", s.handleLoginPage)
	mux.HandleFunc("POST /admin/login", s.handleLogin)
	mux.HandleFunc("POST /admin/logout", s.handleLogout)
	mux.HandleFunc("GET /admin", s.requireAuth(s.handleDashboard))
	mux.HandleFunc("GET /admin/new", s.requireAuth(s.handleNewPost))
	mux.HandleFunc("POST /admin/posts", s.requireAuth(s.handleCreatePost))
	mux.HandleFunc("GET /admin/posts/{id}/edit", s.requireAuth(s.handleEditPost))
	mux.HandleFunc("POST /admin/posts/{id}", s.requireAuth(s.handleUpdatePost))
	mux.HandleFunc("DELETE /admin/posts/{id}", s.requireAuth(s.handleDeletePost))

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Println("ePer7shme Jobs po punon në http://localhost" + addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
