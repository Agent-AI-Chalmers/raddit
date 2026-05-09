package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"` // hashed credential
	Bio       string    `json:"bio"`
	Avatar    string    `json:"avatar"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username,omitempty"`
	Subreddit string    `json:"subreddit"`
	Upvotes   int       `json:"upvotes"`
	Downvotes int       `json:"downvotes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Comment struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	PostID    int64     `json:"post_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username,omitempty"`
	ParentID  *int64    `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Vote struct {
	ID       int64  `json:"id"`
	PostID   int64  `json:"post_id"`
	UserID   int64  `json:"user_id"`
	VoteType string `json:"vote_type"` // "up" or "down"
}

type FileRecord struct {
	ID           int64     `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	UserID       int64     `json:"user_id"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// Auth request payloads
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Next     string `json:"next"` // redirect destination after login
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreatePostRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Subreddit string `json:"subreddit"`
}

type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreateCommentRequest struct {
	Content  string `json:"content"`
	PostID   int64  `json:"post_id"`
	ParentID *int64 `json:"parent_id"`
}

// UpdateProfileRequest handles all updatable user fields
type UpdateProfileRequest struct {
	Email    string `json:"email"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatar"`
	Password string `json:"password"`
}

// TokenClaims holds the data encoded in the JWT
type TokenClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Exp      int64  `json:"exp"`
}
