package main

import (
	"flag"
	"go-blog/controllers"
	"go-blog/helpers"
	"go-blog/models"
	"go-blog/system"
	"strings"

	"github.com/cihub/seelog"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

func main() {

	configFilePath := flag.String("C", "conf/conf.toml", "config file path")
	logConfigPath := flag.String("L", "conf/seelog.xml", "log config file path")
	generate := flag.Bool("g", false, "generate sample config file")
	flag.Parse()

	if *generate {
		system.Generate()
		os.Exit(0)
	}

	logger, err := seelog.LoggerFromConfigAsFile(*logConfigPath)
	if err != nil {
		seelog.Critical("err parsing seelog config file", err)
		return
	}
	seelog.ReplaceLogger(logger)
	defer seelog.Flush()

	if err := system.LoadConfiguration(*configFilePath); err != nil {
		seelog.Critical("err parsing config log file", err)
		return
	}

	db, err := models.InitDB()
	if err != nil {
		seelog.Critical("err open databases", err)
		return
	}
	defer func() {
		dbInstance, _ := db.DB()
		_ = dbInstance.Close()
	}()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	setTemplate(router)
	setSessions(router)
	router.Use(SharedData())

	router.Static("/static", filepath.Join(helpers.GetCurrentDirectory(), "./static"))

	router.NoRoute(controllers.Handle404)
	router.GET("/", controllers.IndexGet)
	router.GET("/index", controllers.IndexGet)

	// 登陆与注册
	router.GET("/signup", controllers.SignupGet)
	router.POST("/signup", controllers.SignupPost)
	router.GET("/signin", controllers.SigninGet)
	router.POST("/signin", controllers.SigninPost)
	router.GET("/logout", controllers.LogoutGet)

	// captcha
	router.GET("/captcha", controllers.CaptchaGet)
	router.GET("/captcha/image/:captchaId", controllers.CaptchaImage)

	// comment
	router.POST("/comment/:id", controllers.CommentRead)
	visitor := router.Group("/visitor")
	visitor.Use(JWTAuthMiddleware())
	{
		visitor.POST("/new_comment", controllers.CommentPost)
		visitor.POST("/comment/:id/delete", controllers.CommentDelete)
	}

	router.GET("/post/:id", controllers.PostGet)

	authorized := router.Group("/admin")
	authorized.Use(JWTAuthMiddleware())
	{
		// index
		authorized.GET("/index", controllers.PostIndex)

		// image upload
		authorized.POST("/upload", controllers.Upload)

		authorized.GET("/post", controllers.PostIndex)
		authorized.GET("/new_post", controllers.PostNew)
		authorized.POST("/new_post", controllers.PostCreate)
		authorized.GET("/post/:id/edit", controllers.PostEdit)
		authorized.POST("/post/:id/edit", controllers.PostUpdate)
		// authorized.POST("/post/:id/publish", controllers.PostPublish)
		authorized.POST("/post/:id/delete", controllers.PostDelete)

		//authorized.GET("/user", controllers.UserIndex)
		//authorized.POST("/user/:id/lock", controllers.UserLock)
	}

	err = router.Run(system.GetConfiguration().Addr)
	if err != nil {
		seelog.Critical(err)
	} else {
		logger.Infof("服务启动")
	}
}

func setTemplate(engine *gin.Engine) {

	funcMap := template.FuncMap{
		"dateFormat": helpers.DateFormat,
		"substring":  helpers.Substring,
		"isOdd":      helpers.IsOdd,
		"isEven":     helpers.IsEven,
		"truncate":   helpers.Truncate,
		"length":     helpers.Len,
		"add":        helpers.Add,
		"minus":      helpers.Minus,
	}

	engine.SetFuncMap(funcMap)
	engine.LoadHTMLGlob(filepath.Join(helpers.GetCurrentDirectory(), system.GetConfiguration().ViewDir))
}

// setSessions initializes sessions & csrf middlewares
func setSessions(router *gin.Engine) {
	cfg := system.GetConfiguration()
	//https://github.com/gin-gonic/contrib/tree/master/sessions
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{HttpOnly: true, MaxAge: 7 * 86400, Path: "/"}) //Also set Secure: true if using SSL, you should though
	router.Use(sessions.Sessions("blog-session", store))
	//https://github.com/utrack/gin-csrf
	/*router.Use(csrf.Middleware(csrf.Options{
		Secret: config.SessionSecret,
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))*/
}

//+++++++++++++ middlewares +++++++++++++++++++++++

// SharedData fills in common data, such as user info, etc...
func SharedData() gin.HandlerFunc {
	return func(c *gin.Context) {
		// userID, _ := c.Get(controllers.SessionKey)
		session := sessions.Default(c)
		userID := session.Get(controllers.SessionKey)
		_, exists := c.Get(controllers.ContextUserKey)
		if userID != nil && !exists {
			user, err := models.GetUser(userID)
			if err == nil {
				c.Set(controllers.ContextUserKey, user)
			}
		}
		c.Next()
	}
}

// 验证通过，放行
func authNext(c *gin.Context, claims *helpers.MyClaims) {
	c.Set(controllers.SessionKey, claims.UserID)
	if user, exist := c.Get(controllers.ContextUserKey); !exist || user == nil {
		temp, _ := models.GetUser(claims.UserID)
		c.Set(controllers.ContextUserKey, temp)
	}
	c.Next()
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// 消息头中不存在认证数据时，采取session机制
		if authHeader == "" {
			session := sessions.Default(c)
			jwtToken := session.Get(controllers.SessionJwtKey)
			jwtTokenStr, _ := jwtToken.(string)
			//session.Get("ExpiresAt")
			//session.Get("UserID")
			//session.Get("Username")

			claims, err := helpers.ParseToken(jwtTokenStr)
			if err != nil {
				c.HTML(http.StatusOK, "error/error.html", gin.H{
					"message": "签名验证不通过",
				})
				c.Abort()
				return
			}

			// 验证通过，放行
			authNext(c, claims)
		} else {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
				c.Abort()
				return
			}

			claims, err := helpers.ParseToken(parts[1])
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				c.Abort()
				return
			}

			// 验证通过，放行
			authNext(c, claims)
		}
	}
}
