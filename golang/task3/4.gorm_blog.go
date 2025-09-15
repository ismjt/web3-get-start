package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User 用户模型，一个用户可以有多篇文章
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null;unique"`
	Password  string `gorm:"size:100;not null" json:"-"`
	Email     string `gorm:"size:100;uniqueIndex;not null"`
	Posts     []Post `gorm:"foreignKey:UserID"` // 一对多关联
	PostCount int    // 文章数量统计字段
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Post 文章模型，一篇文章属于一个用户，有多条评论
type Post struct {
	ID            uint      `gorm:"primaryKey"`
	Title         string    `gorm:"size:200;not null"`
	Content       string    `gorm:"type:text"`
	UserID        uint      // 外键，关联 User
	Comments      []Comment `gorm:"foreignKey:PostID"` // 一对多关联
	CommentStatus string    `gorm:"default:'无评论'"`     // 评论状态字段，如 "无评论" / "有评论"
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func updateUserPostCount(tx *gorm.DB, userID uint) error {
	var count int64
	// 统计指定用户文章数量
	if err := tx.Model(&Post{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return err
	}
	// 更新用户 post_count 字段
	return tx.Model(&User{}).Where("id = ?", userID).Update("post_count", count).Error
}

// 创建前或创建后钩子
func (p *Post) AfterCreate(tx *gorm.DB) (err error) {
	return updateUserPostCount(tx, p.UserID)
}

func (p *Post) AfterDelete(tx *gorm.DB) (err error) {
	return updateUserPostCount(tx, p.UserID)
}

// Comment 评论模型，一条评论属于一篇文章
type Comment struct {
	ID        uint   `gorm:"primaryKey"`
	Content   string `gorm:"type:text;not null"`
	PostID    uint   // 外键，关联 Post
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Post) BeforeCreate(tx *gorm.DB) (err error) {
	if p.CommentStatus == "" {
		p.CommentStatus = "无评论"
	}
	return nil
}

func (c *Comment) AfterCreate(tx *gorm.DB) (err error) {
	// 新增评论后，更新关联文章的 CommentStatus 为 "有评论"
	return tx.Model(&Post{}).
		Where("id = ?", c.PostID).
		Update("comment_status", "有评论").Error
}

// 删除后钩子
func (c *Comment) AfterDelete(tx *gorm.DB) (err error) {
	var count int64
	// 查询文章剩余评论数量
	if err := tx.Model(&Comment{}).Where("post_id = ?", c.PostID).Count(&count).Error; err != nil {
		return err
	}

	// 如果评论数量为0，更新文章评论状态
	if count == 0 {
		return tx.Model(&Post{}).
			Where("id = ?", c.PostID).
			Update("comment_status", "无评论").Error
	}
	return nil
}

// 定义函数
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

type UserResponse struct {
	ID        uint           `json:"id"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	Posts     []PostResponse `json:"posts"`
	PostCount int            `json:"post_count"`
}

type UserSummary struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type PostResponse struct {
	ID            uint              `json:"id"`
	Title         string            `json:"title"`
	Content       string            `json:"content"`
	UpdatedAt     string            `json:"updated_at"`
	CommentCount  int64             `json:"comment_count"`
	Comments      []CommentResponse `json:"comments"`
	CommentStatus string            `json:"comment_status"`
}

type CommentResponse struct {
	ID        uint   `json:"id"`
	Content   string `json:"content"`
	UpdatedAt string `json:"updated_at"`
}

type CommentRequest struct {
	PostID  int    `json:"postId"`
	Content string `json:"content"`
	UserId  int    `json:"userId"`
}

type MostCommentPost struct {
	Title      string
	Author     string
	CommentCnt int
}

func randomTime(now time.Time, minutesMin, minutesMax int) time.Time {
	rand.Seed(time.Now().UnixNano())
	delta := rand.Intn(minutesMax-minutesMin+1) + minutesMin // 随机分钟数
	return now.Add(time.Duration(delta) * time.Minute)
}

// ----------------- 数据初始化 -----------------
func initData(db *gorm.DB) {
	db.AutoMigrate(&User{}, &Post{}, &Comment{})
	// 清空表
	db.Exec("DELETE FROM comments")
	db.Exec("DELETE FROM posts")
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM sqlite_sequence WHERE name='comments'")
	db.Exec("DELETE FROM sqlite_sequence WHERE name='posts'")
	db.Exec("DELETE FROM sqlite_sequence WHERE name='users'")

	// 当前时间
	now := time.Now().Add(-24 * time.Hour)

	// 插入示例数据
	user1 := User{Name: "张三", Email: "zhangsan@demo.com", Password: "123", PostCount: 2, CreatedAt: randomTime(now, -120, 0), UpdatedAt: now}
	user2 := User{Name: "李四", Email: "lisi@demo.com", Password: "abc", PostCount: 2, CreatedAt: randomTime(now, -120, 0), UpdatedAt: now}
	db.Create(&user1)
	db.Create(&user1)
	db.Create(&user2)

	post1 := Post{Title: "Go语言入门和放弃", Content: "Go 是一门编程语言...", UserID: user1.ID, CommentStatus: "有评论", CreatedAt: randomTime(now, 0, 29), UpdatedAt: randomTime(now, 30, 60)}
	post2 := Post{Title: "GORM使用技巧", Content: "GORM 是 Go 的 ORM 框架...", UserID: user1.ID, CommentStatus: "有评论", CreatedAt: randomTime(now, 0, 29), UpdatedAt: randomTime(now, 30, 60)}
	post3 := Post{Title: "前端开发", Content: "前端技术栈分享...", UserID: user2.ID, CommentStatus: "有评论", CreatedAt: randomTime(now, 0, 29), UpdatedAt: randomTime(now, 30, 60)}
	post4 := Post{Title: "WEB3开发", Content: "WEB3兴趣班...", UserID: user2.ID, CreatedAt: randomTime(now, 0, 29), UpdatedAt: randomTime(now, 30, 60)}
	db.Create(&post1)
	db.Create(&post2)
	db.Create(&post3)
	db.Create(&post4)

	db.Create(&Comment{Content: "很好的一篇文章！", PostID: post1.ID, CreatedAt: randomTime(now, 60, 90), UpdatedAt: randomTime(now, 91, 120)})
	db.Create(&Comment{Content: "受益匪浅", PostID: post1.ID, CreatedAt: randomTime(now, 60, 90), UpdatedAt: randomTime(now, 91, 120)})
	db.Create(&Comment{Content: "不错的教程", PostID: post2.ID, CreatedAt: randomTime(now, 60, 90), UpdatedAt: randomTime(now, 91, 120)})
	db.Create(&Comment{Content: "学习了", PostID: post3.ID, CreatedAt: randomTime(now, 60, 90), UpdatedAt: randomTime(now, 91, 120)})
	db.Create(&Comment{Content: "博主赶快更新", PostID: post3.ID, CreatedAt: randomTime(now, 60, 90), UpdatedAt: randomTime(now, 91, 120)})
}

// ----------------- HTTP 处理 -----------------
func indexHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 获取请求参数 userId
		userIdStr := r.URL.Query().Get("userId")
		if userIdStr == "" {
			// 参数不存在 → 重定向到 ?userId=1
			http.Redirect(w, r, r.URL.Path+"?userId=1", http.StatusFound)
			return
		}

		// 2. 将 userId 转为 int
		userId, err := strconv.Atoi(userIdStr)
		if err != nil || userId <= 0 {
			http.Error(w, "无效的 userId 参数", http.StatusBadRequest)
			return
		}

		// 3. 渲染模板
		tmpl, err := template.ParseFiles("templates/blog.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 4. 查询所用用户信息
		var users []UserSummary
		// 只查询 id 和 name 字段
		if err := db.Model(&User{}).Select("id", "name").Find(&users).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 5. 查询最多评论的文章名称
		var mostPosts []MostCommentPost
		// 查询每篇文章的评论数量
		subQuery := db.Model(&Post{}).
			Select("posts.id, posts.title, users.name AS author, COUNT(comments.id) AS comment_cnt").
			Joins("LEFT JOIN comments ON comments.post_id = posts.id").
			Joins("LEFT JOIN users ON users.id = posts.user_id").
			Group("posts.id")
		// 找出最大评论数
		var maxCnt int
		err = db.Table("(?) as t", subQuery).Select("MAX(comment_cnt)").Scan(&maxCnt).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 查询评论数等于最大值的所有文章
		err = db.Table("(?) as t", subQuery).Where("comment_cnt = ?", maxCnt).Scan(&mostPosts).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			UserID       int
			UserNameList []UserSummary
			MostPost     []MostCommentPost
		}{
			UserID:       userId,
			UserNameList: users,
			MostPost:     mostPosts,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// 查询某用户文章及评论
func GetUserPostsJSON(db *gorm.DB, userID uint) (UserResponse, error) {
	var user User
	resp := UserResponse{}

	if err := db.First(&user, userID).Error; err != nil {
		return resp, err
	}

	if err := db.Preload("Comments").Where("user_id = ?", userID).Find(&user.Posts).Error; err != nil {
		return resp, err
	}

	// 构造返回结构
	resp.ID = user.ID
	resp.Name = user.Name
	resp.Email = user.Email
	resp.PostCount = user.PostCount
	for _, p := range user.Posts {
		postResp := PostResponse{
			ID:            p.ID,
			Title:         p.Title,
			Content:       p.Content,
			UpdatedAt:     p.UpdatedAt.Format("2006-01-02 15:04:05"),
			CommentStatus: p.CommentStatus,
		}
		for _, c := range p.Comments {
			postResp.Comments = append(postResp.Comments, CommentResponse{
				ID:        c.ID,
				Content:   c.Content,
				UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		postResp.CommentCount = int64(len(postResp.Comments))
		resp.Posts = append(resp.Posts, postResp)
	}

	return resp, nil
}

// 查询评论数量最多文章
func GetPostWithMostCommentsJSON(db *gorm.DB) (PostResponse, error) {
	var post Post
	var resp PostResponse

	// 查询评论最多文章
	if err := db.Model(&Post{}).
		Joins("LEFT JOIN comments ON comments.post_id = posts.id").
		Group("posts.id").
		Order("COUNT(comments.id) DESC").
		Preload("Comments").
		Limit(1).
		Find(&post).Error; err != nil {
		return resp, err
	}

	resp.ID = post.ID
	resp.Title = post.Title
	resp.Content = post.Content
	resp.UpdatedAt = post.UpdatedAt.Format("2006-01-02 15:04:05")
	resp.CommentCount = int64(len(post.Comments))
	resp.CommentStatus = post.CommentStatus
	for _, c := range post.Comments {
		resp.Comments = append(resp.Comments, CommentResponse{
			ID:        c.ID,
			Content:   c.Content,
			UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return resp, nil
}

func commentHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析 JSON 请求体
		var req CommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "请求 JSON 格式错误: "+err.Error(), http.StatusBadRequest)
			return
		}

		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/?userId="+strconv.Itoa(req.UserId), http.StatusSeeOther)
			return
		}

		// 校验必要字段
		if req.PostID == 0 || req.Content == "" {
			http.Error(w, "缺少postId 或 content", http.StatusBadRequest)
			return
		}

		fmt.Printf("用户ID: %s评论内容：%s\n", req.UserId, req.Content)
		comment := Comment{
			PostID:    uint(req.PostID),
			Content:   req.Content,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(&comment)

		// 返回成功响应 JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"comment": comment,
		})
	}
}

func userPostsHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取 userId 参数
		userIdStr := r.URL.Query().Get("userId")
		if userIdStr == "" {
			http.Error(w, "缺少必传参数 userId", http.StatusBadRequest)
			return
		}

		// 转换为整数
		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil || userId <= 0 {
			http.Error(w, "无效的 userId 参数", http.StatusBadRequest)
			return
		}

		resp, err := GetUserPostsJSON(db, uint(userId))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func postMostCommentsHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 查询评论数量最多的文章
		var post Post
		err := db.Preload("Comments").
			Joins("LEFT JOIN comments ON comments.post_id = posts.id").
			Group("posts.id").
			First(&post).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "没有文章或评论", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// 渲染模板
		tmpl, err := template.ParseFiles("templates/post_most_comments.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			Post Post
		}{
			Post: post,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func createPostHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// 返回发布文章页面
			tmpl, err := template.ParseFiles("templates/post_create.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			userIdStr := r.URL.Query().Get("userId")
			var targetUser User
			if err := db.Model(&User{}).Where("id = ?", userIdStr).Select("name").First(&targetUser).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			data := struct {
				UserID string
				User   User
			}{
				UserID: userIdStr,
				User:   targetUser,
			}
			_ = tmpl.Execute(w, data)
			return
		}

		if r.Method == http.MethodPost {
			// 解析表单
			if err := r.ParseForm(); err != nil {
				http.Error(w, "解析表单失败: "+err.Error(), http.StatusBadRequest)
				return
			}

			title := r.FormValue("title")
			userIdStr := r.FormValue("userId")
			content := r.FormValue("content")
			if title == "" || content == "" {
				http.Error(w, "标题或内容不能为空", http.StatusBadRequest)
				return
			}
			userId, _ := strconv.ParseInt(userIdStr, 10, 64)
			post := Post{
				UserID:        uint(userId),
				Title:         title,
				Content:       content,
				CommentStatus: "无评论",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}

			if err := db.Create(&post).Error; err != nil {
				http.Error(w, "创建文章失败: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// 创建成功后跳转到首页或文章详情页
			http.Redirect(w, r, "/?userId="+userIdStr, http.StatusSeeOther)
		}
	}
}

func deletePostHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		postIDStr := r.URL.Query().Get("id")
		postID, err := strconv.ParseUint(postIDStr, 10, 64)
		if err != nil {
			http.Error(w, "文章 ID 无效", http.StatusBadRequest)
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			var post Post
			if err := tx.First(&post, postID).Error; err != nil {
				return err
			}

			// 先删除评论
			if err := tx.Where("post_id = ?", post.ID).Delete(&Comment{}).Error; err != nil {
				return err
			}

			// 删除文章
			if err := tx.Delete(&post).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			http.Error(w, "删除文章失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "删除成功"}`))
	}
}

func main() {
	db, err := gorm.Open(sqlite.Open("blog.db"), &gorm.Config{})
	if err != nil {
		log.Fatalln("数据库连接失败:", err)
	}

	initData(db)

	http.HandleFunc("/", indexHandler(db))
	http.HandleFunc("/api/comment", commentHandler(db))
	http.HandleFunc("/api/user_posts", userPostsHandler(db))
	http.HandleFunc("/post_most_comments", postMostCommentsHandler(db))
	http.HandleFunc("/post_create", createPostHandler(db))
	http.HandleFunc("/api/post_delete", deletePostHandler(db))

	log.Println("服务启动: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
