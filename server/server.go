package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/ai-marketing/ai-marketing-server/config"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/middlewares"
)

type ginServer struct {
	app  *gin.Engine
	db   *sqlx.DB
	conf *config.Config
}

var (
	once   sync.Once
	server *ginServer
)

func NewGinServer(conf *config.Config, db *sqlx.DB) *ginServer {
	mode := conf.Server.Environment

	if mode == "dev" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ginApp := gin.New()

	once.Do(func() {
		server = &ginServer{
			app:  ginApp,
			db:   db,
			conf: conf,
		}
	})

	return server
}

func (s *ginServer) Start() {
	serverMiddleware := middlewares.NewServerMiddleware()

	s.app.Use(serverMiddleware.GetCorsMiddleware())
	s.app.Use(serverMiddleware.GetBodyLimitMiddleware(s.conf.Server.BodyLimit))
	s.app.Use(serverMiddleware.GetTimeoutMiddleware(s.conf.Server.TimeOut))
	s.app.Use(gin.Logger())
	s.app.Use(gin.Recovery())

	s.app.GET("/health", s.healthCheck)

	s.initAuthRouter()
	s.initUserRouter()
	s.initCompanyRouter()
	s.initBrandRouter()
	s.initCategoryRouter()
	s.initPromptRouter()
	s.initVisibilityRouter()
	s.initCitationRouter()
	s.initNotificationRouter()
	s.initApiKeyRouter()
	s.initMemberRouter()
	s.initSubdomainRouter()
	s.initDashboardRouter()

	url := fmt.Sprintf(":%d", s.conf.Server.Port)
	srv := &http.Server{
		Addr:    url,
		Handler: s.app,
	}

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGINT, syscall.SIGTERM)
	go s.gracefullyShutdown(quitCh, srv)

	logs.Info("AI Marketing Server started at port: " + strconv.Itoa(s.conf.Server.Port))
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logs.Info("Error: AI Marketing Server stopped: " + err.Error())
		return
	}
}

func (s *ginServer) gracefullyShutdown(quitCh chan os.Signal, srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-quitCh
	log.Println("Shutdown Server...")

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	<-ctx.Done()
	log.Println("timeout of 5 seconds.")
	log.Println("Server exiting")
}

func (s *ginServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}
