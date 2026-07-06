package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type Post struct {
	ID        int64
	Slug      string
	Title     string
	Excerpt   string
	Content   string
	Cover     string
	Published bool
	CreatedAt time.Time
}

type User struct {
	ID           int64
	Username     string
	PasswordHash string
}

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// modernc.org/sqlite works best with a single writer connection.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;`); err != nil {
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		expires_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slug TEXT NOT NULL UNIQUE,
		title TEXT NOT NULL,
		excerpt TEXT NOT NULL DEFAULT '',
		content TEXT NOT NULL DEFAULT '',
		cover TEXT NOT NULL DEFAULT '',
		published INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL
	);
	`)
	return err
}

func seed(db *sql.DB) error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		user := os.Getenv("ADMIN_USER")
		if user == "" {
			user = "admin"
		}
		pass := os.Getenv("ADMIN_PASSWORD")
		if pass == "" {
			pass = "eper7shme2026"
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err := db.Exec(`INSERT INTO users(username, password_hash) VALUES(?,?)`, user, string(hash)); err != nil {
			return err
		}
		log.Printf("U krijua administratori: %s (fjalëkalimi: %s) — ndryshojeni me ADMIN_USER/ADMIN_PASSWORD", user, pass)
	}

	if err := db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		now := time.Now()
		samples := []Post{
			{
				Slug:    "si-te-shkruash-nje-cv-qe-bie-ne-sy",
				Title:   "Si të shkruash një CV që bie në sy",
				Excerpt: "Rekruterët shpenzojnë mesatarisht 7 sekonda mbi një CV. Ja si t'i shfrytëzosh ato sekonda në maksimum.",
				Content: "Rekruterët shpenzojnë mesatarisht **7 sekonda** mbi një CV përpara se të vendosin nëse do ta lexojnë më tej. Kjo do të thotë se çdo rresht duhet të punojë për ty.\n\n## Fillo me thelbin\n\nVendos në krye emrin, pozicionin që synon dhe 2–3 rreshta përmbledhje ku tregon vlerën tënde. Jo objektiva të përgjithshme — rezultate konkrete.\n\n## Numrat flasin më shumë se fjalët\n\n- *\"Rrita shitjet\"* → **\"Rrita shitjet me 32% në 6 muaj\"**\n- *\"Menaxhova ekipin\"* → **\"Drejtova një ekip prej 8 personash\"**\n\n## Përshtate për çdo aplikim\n\nNjë CV e vetme për çdo pozicion nuk funksionon më. Lexo me kujdes njoftimin e punës dhe përdor fjalët kyçe të tij — shumë kompani përdorin sisteme automatike filtrimi.\n\n## Mbaje të pastër\n\nNjë faqe, maksimumi dy. Font i lexueshëm, hapësirë e bardhë e mjaftueshme dhe zero gabime drejtshkrimore. CV-ja jote është shembulli i parë i punës sate.",
				Published: true,
				CreatedAt: now.AddDate(0, 0, -3),
			},
			{
				Slug:    "5-gabimet-me-te-shpeshta-ne-intervista-pune",
				Title:   "5 gabimet më të shpeshta në intervista pune",
				Excerpt: "Nga mungesa e përgatitjes te fjalët e tepërta — gabimet që i shohim çdo javë dhe si t'i shmangësh.",
				Content: "Pas qindra intervistash të ndjekura nga afër, disa gabime përsëriten vazhdimisht. Lajmi i mirë? Të gjitha shmangen lehtë.\n\n## 1. Nuk njeh kompaninë\n\nPyetja *\"Çfarë di për ne?\"* vjen pothuajse gjithmonë. 10 minuta kërkim në faqen e kompanisë dhe rrjetet sociale të japin avantazh të madh.\n\n## 2. Flet keq për punëdhënësin e mëparshëm\n\nEdhe kur ke pasur përvojë të vështirë, formulo me profesionalizëm: *\"Kërkoja një mjedis me më shumë mundësi rritjeje.\"*\n\n## 3. Përgjigje të mësuara përmendsh\n\nRekruterët e dallojnë menjëherë. Më mirë përgjigje të sinqerta me shembuj realë sesa fraza perfekte pa përmbajtje.\n\n## 4. Nuk bën asnjë pyetje\n\nIntervista është e dyanshme. Pyet për ekipin, sfidat e pozicionit, hapat e mëtejshëm — tregon interes të vërtetë.\n\n## 5. Harron ndjekjen pas intervistës\n\nNjë email i shkurtër falënderimi brenda 24 orëve të mban në mendjen e rekruterit dhe tregon seriozitet.",
				Published: true,
				CreatedAt: now.AddDate(0, 0, -10),
			},
			{
				Slug:    "pse-rekrutimi-i-jashtem-ju-kursen-kohe-dhe-para",
				Title:   "Pse rekrutimi i jashtëm ju kursen kohë dhe para",
				Excerpt: "Një pozicion i lirë kushton më shumë sesa mendoni. Ja pse gjithnjë e më shumë kompani ia besojnë rekrutimin një partneri.",
				Content: "Sa kushton në të vërtetë një pozicion i lirë? Studimet tregojnë se një vend i paplotësuar pune i kushton kompanisë deri në **30% të pagës vjetore** të atij pozicioni — në produktivitet të humbur, mbingarkesë të ekipit dhe mundësi të humbura.\n\n## Koha është kostoja më e madhe\n\nNjë proces i brendshëm rekrutimi zgjat mesatarisht 6–8 javë: shkrimi i njoftimit, filtrimi i qindra CV-ve, intervistat e para... Me një partner rekrutimi, ju merrni vetëm kandidatët e përzgjedhur — të verifikuar dhe të motivuar.\n\n## Rrjeti bën diferencën\n\nKandidatët më të mirë shpesh nuk janë duke kërkuar punë në mënyrë aktive. Një partner rekrutimi me komunitet të ndërtuar i arrin edhe ata — njerëzit që nuk do t'i gjenit kurrë me një njoftim të thjeshtë.\n\n## Diskrecion i plotë\n\nNdonjëherë një pozicion duhet plotësuar pa u bërë publik. Me rekrutim të jashtëm, kompania juaj mbetet anonime deri në momentin e duhur. 😉\n\n## Rezultati\n\n95% e klientëve tanë rikthehen për bashkëpunime të reja — sepse kur procesi funksionon, e sheh në ekipin tënd.",
				Published: true,
				CreatedAt: now.AddDate(0, 0, -18),
			},
		}
		for _, p := range samples {
			if _, err := db.Exec(
				`INSERT INTO posts(slug, title, excerpt, content, cover, published, created_at) VALUES(?,?,?,?,?,?,?)`,
				p.Slug, p.Title, p.Excerpt, p.Content, p.Cover, boolToInt(p.Published), p.CreatedAt.UTC().Format(time.RFC3339),
			); err != nil {
				return err
			}
		}
		log.Println("U shtuan 3 postime shembull në blog")
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ---- Posts ----

const postCols = `id, slug, title, excerpt, content, cover, published, created_at`

func scanPost(row interface{ Scan(...any) error }) (Post, error) {
	var p Post
	var pub int
	var created string
	if err := row.Scan(&p.ID, &p.Slug, &p.Title, &p.Excerpt, &p.Content, &p.Cover, &pub, &created); err != nil {
		return p, err
	}
	p.Published = pub == 1
	if t, err := time.Parse(time.RFC3339, created); err == nil {
		p.CreatedAt = t
	}
	return p, nil
}

// listPosts returns up to limit posts; the second value reports whether more exist.
func (s *Server) listPosts(q string, publishedOnly bool, limit, offset int) ([]Post, bool, error) {
	var where []string
	var args []any
	if publishedOnly {
		where = append(where, "published = 1")
	}
	if q != "" {
		where = append(where, "(title LIKE ? OR excerpt LIKE ?)")
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	query := `SELECT ` + postCols + ` FROM posts`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?"
	args = append(args, limit+1, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, false, err
		}
		posts = append(posts, p)
	}
	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}
	return posts, hasMore, rows.Err()
}

func (s *Server) getPostBySlug(slug string, publishedOnly bool) (Post, error) {
	query := `SELECT ` + postCols + ` FROM posts WHERE slug = ?`
	if publishedOnly {
		query += ` AND published = 1`
	}
	return scanPost(s.db.QueryRow(query, slug))
}

func (s *Server) getPostByID(id int64) (Post, error) {
	return scanPost(s.db.QueryRow(`SELECT `+postCols+` FROM posts WHERE id = ?`, id))
}

func (s *Server) createPost(p Post) error {
	_, err := s.db.Exec(
		`INSERT INTO posts(slug, title, excerpt, content, cover, published, created_at) VALUES(?,?,?,?,?,?,?)`,
		p.Slug, p.Title, p.Excerpt, p.Content, p.Cover, boolToInt(p.Published), time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (s *Server) updatePost(p Post) error {
	_, err := s.db.Exec(
		`UPDATE posts SET slug=?, title=?, excerpt=?, content=?, cover=?, published=? WHERE id=?`,
		p.Slug, p.Title, p.Excerpt, p.Content, p.Cover, boolToInt(p.Published), p.ID,
	)
	return err
}

func (s *Server) deletePost(id int64) error {
	_, err := s.db.Exec(`DELETE FROM posts WHERE id=?`, id)
	return err
}

// uniqueSlug returns slug, appending -2, -3… if it already belongs to another post.
func (s *Server) uniqueSlug(slug string, excludeID int64) string {
	if slug == "" {
		slug = "postim"
	}
	candidate := slug
	for i := 2; ; i++ {
		var id int64
		err := s.db.QueryRow(`SELECT id FROM posts WHERE slug = ?`, candidate).Scan(&id)
		if err == sql.ErrNoRows || (err == nil && id == excludeID) {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", slug, i)
	}
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.NewReplacer("ë", "e", "ç", "c").Replace(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// ---- Auth / sessions ----

const sessionCookie = "eper7shme_session"

func (s *Server) getUserByUsername(username string) (User, error) {
	var u User
	err := s.db.QueryRow(`SELECT id, username, password_hash FROM users WHERE username = ?`, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash)
	return u, err
}

func (s *Server) createSession(userID int64) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	expires := time.Now().Add(7 * 24 * time.Hour).UTC().Format(time.RFC3339)
	if _, err := s.db.Exec(`INSERT INTO sessions(token, user_id, expires_at) VALUES(?,?,?)`, token, userID, expires); err != nil {
		return "", err
	}
	return token, nil
}

func (s *Server) validSession(r *http.Request) bool {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		return false
	}
	var expires string
	if err := s.db.QueryRow(`SELECT expires_at FROM sessions WHERE token = ?`, c.Value).Scan(&expires); err != nil {
		return false
	}
	t, err := time.Parse(time.RFC3339, expires)
	if err != nil || time.Now().After(t) {
		s.db.Exec(`DELETE FROM sessions WHERE token = ?`, c.Value)
		return false
	}
	return true
}

func (s *Server) destroySession(r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil && c.Value != "" {
		s.db.Exec(`DELETE FROM sessions WHERE token = ?`, c.Value)
	}
}
