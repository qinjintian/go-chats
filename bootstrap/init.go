package bootstrap

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
	"go-chats/app/global/variable"
	"go-chats/app/model"
	"go-chats/app/utils/filer"
	"go-chats/routers"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Bootstrap struct{}

func Init(e *gin.Engine, cfg *ini.File) {
	// 自定义日志格式
	LoggerWithFormatter(e)

	// 定义静态资源路由与实际目录映射关系
	MappingDirectory(e, cfg)

	// 初始化数据库连接
	InitDB(cfg)

	// 加载模板
	LoadHTMLGlob(e)

	// 启用Session
	EnableSession(e)

	// 初始化路由
	routers.InitRouter(e)

	// 监听HTTP服务，必须放在最后
	ListenAndServe(e, cfg)
}

// 定义静态资源路由与实际目录映射关系
func MappingDirectory(engine *gin.Engine, cfg *ini.File) {
	staticDir := cfg.Section(ini.DefaultSection).Key("STATIC_DIR").MustString("/static")
	if !filer.IsDir(staticDir) {
		_ = os.MkdirAll(staticDir, os.ModePerm)
	}
	engine.StaticFS("/static", http.Dir(staticDir))

	storageDir := cfg.Section(ini.DefaultSection).Key("STORAGE_DIR").MustString("/storage/app/public")
	if !filer.IsDir(storageDir) {
		_ = os.MkdirAll(storageDir, os.ModePerm)
	}
	engine.StaticFS("/storage", http.Dir(storageDir))
}

// 自定义日志格式
func LoggerWithFormatter(engine *gin.Engine) {
	// 日志文件不需要颜色
	gin.DisableConsoleColor()

	// 创建日志文件并设置为 gin.DefaultWriter
	f, _ := os.OpenFile("storage/logs/gin.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout) // 同时写入日志文件和控制台

	// LoggerWithFormatter 中间件会将日志写入 gin.DefaultWriter
	// 默认情况下 gin.DefaultWriter 是 os.Stdout
	// engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
	// 自定义日志输出格式
	// return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
	// 	param.ClientIP,
	// 	param.TimeStamp.Format("2006-01-02 15:04:05"),
	// 	param.Method,
	// 	param.Path,
	// 	param.Request.Proto,
	// 	param.StatusCode,
	// 	param.Latency,
	// 	param.Request.UserAgent(),
	// 	param.ErrorMessage,
	// )
	// }))
}

// 监听HTTP服务和信号
func ListenAndServe(r *gin.Engine, cfg *ini.File) {
	addr := cfg.Section(ini.DefaultSection).Key("HTTP_ADDR").MustString("")
	port := cfg.Section(ini.DefaultSection).Key("HTTP_PORT").MustString("8080")
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", addr, port),
		Handler:        r,                // 处理程序调用，路由引擎
		ReadTimeout:    10 * time.Second, // 读取超时
		WriteTimeout:   10 * time.Second, // 写入超时
		MaxHeaderBytes: 1 << 20,          // 最大报头字节
	}

	b := new(Bootstrap)

	// 监听HTTP服务
	go b.listenServe(srv)

	// 监听信号平滑重启HTTP服务
	b.listenSignal(context.Background(), srv)
}

func (b *Bootstrap) listenServe(srv *http.Server) {
	// 服务连接
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

// 监听信号
func (b *Bootstrap) listenSignal(ctx context.Context, srv *http.Server) {
	sig := make(chan os.Signal, 1)

	// kill （无参数）默认发送 syscall.SIGTERM
	// kill -2 指的是 syscall.SIGINT
	// kill -9 指的是 syscall.SIGKILL 但是不能被捕获，所以不需要添加它

	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sig // 接收到信号量
	log.Println("正在关闭服务器 ...")
	timeoutCtx, _ := context.WithTimeout(ctx, 1 * time.Second) // 设置超过N秒所有程序未闲置也会硬终止服务
	if err := srv.Shutdown(timeoutCtx); err != nil {
		log.Fatal("服务器关闭:", err)
	}

	select {
	case <-timeoutCtx.Done():
		// 捕获ctx.Done()  5秒超时。
		log.Println("5秒超时。")
	}
	log.Println("服务器退出。")
}

// 初始化数据库连接
func InitDB(cfg *ini.File) {
	_, _ = model.InitDB(cfg)
}

// 加载模板
func LoadHTMLGlob(r *gin.Engine) {
	r.LoadHTMLGlob("templates/*.html")
}

// 启用Session
func EnableSession(r *gin.Engine) {
	gob.Register(variable.UserSessionData{}) // 跨路由存取复杂结构的session数据，需要注册数据类型
	store := cookie.NewStore([]byte("secret"))
	// store.Options(sessions.Options{
	// 	MaxAge: int(30 * time.Minute), // 30min
	// 	Path:   "/",
	// })
	r.Use(sessions.Sessions("session", store))
}
