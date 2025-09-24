# 1、项目背景
任务4-个人博客系统

# 2、工程结构
本项目工程化实践方面参考了开源项目，目录结构说明：
```
-task4(personal blog)
|-conf 配置文件目录
|-controllers 控制器目录
|-helpers 公共方法目录
|-log 项目运行日志存储目录
|-models 数据库访问目录
|-static 静态资源目录
    |-css css文件目录
    |-fonts 项目字体文件目录
    |-img 图片目录
    |-js js文件目录
    |-lib 前端组件类库
    |-upload 附件上传
|-system 系统配置文件加载目录
|-tests 测试目录
|-views HTML模板文件目录
|-main.go 程序执行入口
```

# 3、开发说明
## 3.1、IDE启动
启动main.go中的主函数
## 3.2、终端启动
```
go run main.go
```
## 3.3、访问项目
在conf/conf.toml中指定了项目启动地址为[本地8081端口](http://127.0.0.1:8081)
## 3.4、测试用例
单元测试：目前只实现了两个简单的单元测试，目的是为了了解如何实现
## 3.5、接口结果

# 4、配置文件
* 实现方式：基于```github.com/pelletier/go-toml/v2```，加载指定文件中的配置到运行时；
* 默认配置：本项目在main.go启动运行时会读取conf.toml该配置文件中的预置配置；
* 生成配置：启动main.go时添加```-g```参数，将在conf文件夹下生成conf.sample.toml配置样例文件，方便真实环境部署使用；

# 5、日志记录
* 实现方式：基于```github.com/cihub/seelog```，，加载指定日志配置到运行时；
* 默认配置：详见conf/seelog.xml
* 存储位置：log/*.log文件
* 使用方式：可参考controllers/comment.go中的使用

# 6、数据库
## 6.1、sqlite
本项目使用的是sqlite嵌入式数据库，程序中指定数据库文件存储位置为项目根目录下的db文件夹内的```personal_blog.db```文件
## 6.2、数据模型设计
* users 表：存储用户信息
```sqlite
create table users
(
    id         integer primary key autoincrement,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    username   text not null constraint uni_users_username unique,
    password   text not null,
    avatar_url text,
    email      text not null constraint uni_users_email unique
);
create index idx_users_deleted_at
    on users (deleted_at);
```

* posts 表：存储博客文章信息
```sqlite
create table posts
(
    id            integer primary key autoincrement,
    created_at    datetime,
    updated_at    datetime,
    deleted_at    datetime,
    title         text     not null,
    content       longtext not null,
    view          integer,
    user_id       integer constraint fk_users_posts references users,
    comment_total integer
);
create index idx_posts_deleted_at
    on posts (deleted_at);
```

* comments 表：存储文章评论信息
```sqlite
create table comments
(
    id         integer primary key autoincrement,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    content    text not null,
    user_id    integer constraint fk_comments_user references users,
    post_id    integer constraint fk_posts_comments references posts
);
create index idx_comments_deleted_at
    on comments (deleted_at);
```

# 7、用户认证与授权
* JWT（JSON Web Token）：实现用户认证和授权，组件使用github.com/golang-jwt/jwt/v5，JWT的生成与验证在helpers/jwt.go文件内，Claims声明信息为：
```go
type MyClaims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    jwt.RegisteredClaims
}
```
同时实现Claims自定义校验方法Validate()
* Session Store：通过会话机制与客户端进行交互，主要用于认证数据交互，基于cookie方式存储，main.go文件中的关键代码为：
```go
func setSessions(router *gin.Engine) {
	cfg := system.GetConfiguration()
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{HttpOnly: true, MaxAge: 7 * 86400, Path: "/"}) //Also set Secure: true if using SSL, you should though
	router.Use(sessions.Sessions("gin-session", store))
}
```
* Gin中间件实现：详见main.go中的JWTAuthMiddleware、SharedData方法，后端通过读取session中的token数据并完成解析，实现用户身份标记
* 用户登陆：
  * 接口地址：http://127.0.0.1:8081/signin
  * 请求数据格式：可参考models.go中的LoginRequest结构体定义
  * 响应数据格式：可参考models/response.go中的LoginData结构体，参考示例：
```json
{
"code": 200,
"msg": "success",
"payload": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJ1c2VybmFtZSI6InRlc3QxIiwiZXhwIjoxNzU4ODE0MjMzLCJpYXQiOjE3NTg3Mjc4MzMsImlzcyI6InBlcnNvbmFsLWJsb2ctc2VydmVyIn0.EA_0qtiqmhLrbeFECGXGP5M35Cl3kmc6AsBnSTr5p2E",
    "expires_at": 1758814233,
    "user_id": 2,
    "username": "test1"
  }
}
```

# 8、附件上传
接收博文相关附件，目前程序中设置的是仅限图片附件，存储位置为static/upload目录，实际页面中暂未提供相关功能

# 9、项目需求与实现情况
## 9.1、文章管理功能：
* 实现文章的创建功能，只有已认证的用户才能创建文章，创建文章时需要提供文章的标题和内容。<br/>
  http://127.0.0.1:8081/admin/new_post
* 实现文章的读取功能，支持获取所有文章列表和单个文章的详细信息。 <br/>
  http://127.0.0.1:8081/post/:id
* 实现文章的更新功能，只有文章的作者才能更新自己的文章。<br/>
  http://127.0.0.1:8081/admin/post/:id/edit
  > 后端实现了post归属用户与当前登陆用户比对
* 实现文章的删除功能，只有文章的作者才能删除自己的文章。<br/>
  http://127.0.0.1:8081/admin/post/:id/delete
  > 该接口仅为API数据接口，后端根据deleted_at是否有NULL标识来实现逻辑删除（同时实现了真实删除）

## 9.2、评论功能
* 实现评论的创建功能，已认证的用户可以对文章发表评论。<br/>
  http://127.0.0.1:8081/visitor/new_comment
* 实现评论的读取功能，支持获取某篇文章的所有评论列表。<br/>
  http://127.0.0.1:8081/post/:id
  > 进入文章页会加载并解析出该文章的评论数据


# 10、Q&A
## 调试问题（hot reload）？

## 工程化最佳实践？

## 测试任务有哪些？

## AI辅助编程利用情况