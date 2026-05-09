package database

import (
	"database/sql"
	"log"
	"raddit/config"
	_ "modernc.org/sqlite"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func Initialize() {
	var err error
	DB, err = sql.Open("sqlite", config.DBPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatal("Database unreachable:", err)
	}

	createTables()
	seedAdmin()
	log.Println("Database initialized")
}

func seedAdmin() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM users WHERE username=?", config.AdminUsername).Scan(&count)
	if count > 0 {
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(config.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash admin password:", err)
	}
	_, err = DB.Exec(
		"INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, 'admin')",
		config.AdminUsername, config.AdminUsername+"@raddit.local", string(hashed),
	)
	if err != nil {
		log.Fatal("Failed to seed admin user:", err)
	}
	log.Printf("Admin user created: %s", config.AdminUsername)
}

func createTables() {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			bio TEXT DEFAULT '',
			avatar TEXT DEFAULT '',
			role TEXT DEFAULT 'user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			subreddit TEXT DEFAULT 'general',
			upvotes INTEGER DEFAULT 0,
			downvotes INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			parent_id INTEGER DEFAULT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS votes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			vote_type TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(post_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT NOT NULL,
			original_name TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			size INTEGER NOT NULL,
			mime_type TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			log.Fatalf("Failed to create table: %v\nSQL: %s", err, stmt)
		}
	}
}
