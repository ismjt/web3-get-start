package models

import (
	"database/sql"
	"go-blog/system"
	"log"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 用户信息
type User struct {
	gorm.Model
	Username  string `gorm:"unique;not null"`
	Password  string `gorm:"not null" json:"-"`
	AvatarUrl string
	Email     string `gorm:"unique;not null"`
	Posts     []Post `gorm:"foreignKey:UserID"` // 一对多关联
}

func (User) TableName() string {
	return "users"
}

// 文章信息
type Post struct {
	gorm.Model
	Title        string `gorm:"type:text;not null"`
	Content      string `gorm:"type:longtext;not null"`
	View         int    // view count
	UserID       uint
	User         User
	Comments     []Comment `gorm:"foreignKey:PostID"`
	CommentTotal int       `gorm:"->"` // count of comment
}

func (Post) TableName() string {
	return "posts"
}

type Comment struct {
	gorm.Model
	Content string `gorm:"not null"`
	UserID  uint
	User    User
	PostID  uint
	Post    Post
}

func (Comment) TableName() string {
	return "comments"
}

// query result
type QrArchive struct {
	ArchiveDate time.Time //month
	Total       int       //total
	Year        int       // year
	Month       int       // month
}

var DB *gorm.DB

// 获取项目根目录 db 文件路径
func GetDBPath(dbName string) string {
	testDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatal("Failed to get absolute path:", err)
	}

	// 返回项目根目录 db 文件路径
	return filepath.Join(testDir, ".", "db", dbName)
}

func InitDB() (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
		cfg = system.GetConfiguration()
	)
	db, err = gorm.Open(sqlite.Open(GetDBPath(cfg.Database.DSN)), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	DB = db

	// 自动迁移模型
	db.AutoMigrate(&User{}, &Post{}, &Comment{})

	return db, err
}

// Post
func (post *Post) Insert() error {
	return DB.Create(post).Error
}

func (post *Post) Update() error {
	return DB.Model(post).Updates(map[string]interface{}{
		"title":      post.Title,
		"content":    post.Content,
		"updated_at": time.Now(),
	}).Error
}

func (post *Post) Delete() error {
	return DB.Delete(post).Error
}

func (post *Post) LogicDelete() error {
	return DB.Model(post).Updates(map[string]interface{}{
		"deleted_at": time.Now(),
	}).Error
}

func ListMaxCommentPost() (posts []*Post, err error) {
	var (
		rows *sql.Rows
	)
	rows, err = DB.Raw("select p.*,c.total comment_total from posts p inner join (select post_id,count(*) total from comments group by post_id) c on p.id = c.post_id order by c.total desc limit 5").Rows()
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var post Post
		DB.ScanRows(rows, &post)
		posts = append(posts, &post)
	}
	return
}

func ListMaxReadPost() (posts []*Post, err error) {
	err = DB.Order("updated_at desc").Limit(5).Find(&posts).Error
	return
}

func ListAllPost() ([]*Post, error) {
	return _listPost(0, 0)
}

func _listPost(pageIndex, pageSize int) ([]*Post, error) {
	var posts []*Post
	var err error
	if pageIndex > 0 {
		err = DB.Order("created_at desc").Limit(pageSize).Offset((pageIndex - 1) * pageSize).Find(&posts).Error
	} else {
		err = DB.Order("created_at desc").Find(&posts).Error
	}
	return posts, err
}

func CountPost() (count int, err error) {
	err = DB.Raw("select count(*) from posts p").Row().Scan(&count)
	return
}

func ListPostArchives() ([]*QrArchive, error) {
	var (
		archives []*QrArchive
	)
	querySql := `select strftime('%Y-%m',created_at) as month,count(*) as total from posts group by month order by month desc`
	rows, err := DB.Raw(querySql).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var archive QrArchive
		var month string
		rows.Scan(&month, &archive.Total)
		//DB.ScanRows(rows, &archive)
		archive.ArchiveDate, _ = time.Parse("2006-01", month)
		archive.Year = archive.ArchiveDate.Year()
		archive.Month = int(archive.ArchiveDate.Month())
		archives = append(archives, &archive)
	}
	return archives, nil
}

// user
// insert user
func (user *User) Insert() error {
	return DB.Create(user).Error
}

// 登录请求体
type LoginRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// 注册
type RegisterRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Email    string `form:"email" json:"email"`
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	err := DB.First(&user, "username = ?", username).Error
	return &user, err
}

func GetUser(id interface{}) (*User, error) {
	var user User
	err := DB.Where("ID = ?", id).First(&user).Error
	return &user, err
}

func (user *User) UpdateEmail(email string) error {
	if len(email) > 0 {
		return DB.Model(user).Update("email", email).Error
	} else {
		return DB.Model(user).Update("email", gorm.Expr("NULL")).Error
	}
}

// Comment
func (comment *Comment) Insert() error {
	return DB.Create(comment).Error
}

func (comment *Comment) Update() error {
	return DB.Model(comment).UpdateColumn("read_state", true).Error
}

func (comment *Comment) Delete() error {
	return DB.Delete(comment, "user_id = ?", comment.UserID).Error
}

func CountComment() int64 {
	var count int64
	DB.Model(&Comment{}).Count(&count)
	return count
}

func ListUserComment(userID uint) ([]*Comment, error) {
	var comments []*Comment
	err := DB.Where("user_id = ?", userID).Order("created_at desc").Find(&comments).Error
	return comments, err
}

func ListAllComment() ([]*Comment, error) {
	var comments []*Comment
	err := DB.Order("created_at desc").Find(&comments).Error
	return comments, err
}

func ListCommentByPostID(id uint) ([]Comment, error) {
	//var comments []Comment
	//rows, err := DB.Raw("select c.*,u.avatar_url from comments c inner join users u on c.user_id = u.id where c.post_id = ? order by created_at desc", id).Rows()
	//if err != nil {
	//	return nil, err
	//}
	//defer rows.Close()
	//for rows.Next() {
	//	var comment Comment
	//	DB.ScanRows(rows, &comment)
	//	comments = append(comments, comment)
	//}
	//return comments, err
	var comments []Comment
	err := DB.Preload("User").
		Where("post_id = ?", id).
		Find(&comments).Error
	return comments, err
}

func GetPostById(id uint) (*Post, error) {
	var post Post
	err := DB.Where("deleted_at IS NULL").First(&post, "id = ?", id).Error
	return &post, err
}
