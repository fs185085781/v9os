package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/database"
	"github.com/fs185085781/v9os/internal/fuse"
	"github.com/fs185085781/v9os/internal/inface/distributed"
	infaceUser "github.com/fs185085781/v9os/internal/inface/user"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/middleware"
	"github.com/fs185085781/v9os/internal/model/system"
	"github.com/fs185085781/v9os/internal/plugin"
	"github.com/fs185085781/v9os/internal/plugin/manager"
	"github.com/fs185085781/v9os/internal/queue"
	"github.com/fs185085781/v9os/internal/store"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type Server interface {
	StartSync() error
	StartAsync(func(error))
	Close()
}
type server struct {
	engine *gin.Engine
	http   *http.Server
	cfg    config.Config
	// Middleware
	middlewares struct {
		cors       *middleware.CORS
		rateLimit  *middleware.RateLimiter
		requestLog *middleware.RequestLog
		user       *middleware.User
	}
	log      logger.Logger
	needStop bool
	sch      *plugin.TimeTaskSchedule
}

func (s *server) Close() {
	//关闭插件
	if uioc.Has(ioc.KeyPluginManage) {
		all := uioc.Get[*manager.AllPluginManage](ioc.KeyPluginManage)
		all.Switch(1).CloseAll()
		all.Switch(2).CloseAll()
		all.Switch(3).CloseAll()
	}
	//关闭Mq
	if uioc.Has(ioc.KeyQueue) {
		uioc.Queue().Close()
	}
	if uioc.Has(ioc.KeyDistributedProvider) {
		uioc.Get[distributed.DistributedProvider](ioc.KeyDistributedProvider).Close()
	}
	//关闭缓存
	if uioc.Has(ioc.KeyCache) {
		uioc.Cache().Close()
	}
	//关闭数据库
	if uioc.Has(ioc.KeyDatabase) {
		uioc.Database().Close()
	}
	//关闭日志
	if uioc.Has(ioc.KeyLog) {
		uioc.Log().Close()
	}
	//关闭定时任务
	if s.sch != nil {
		s.sch.Stop()
	}
	//关闭http
	s.stop()
}

// StartAsync implements Server.
func (s *server) StartAsync(fn func(error)) {
	util.Go(func() {
		err := s.StartSync()
		if s.needStop {
			err = nil
		}
		fn(err)
	})
}

var beforeCloseFn func()

func (s *server) StartSync() error {
	s.needStop = false
	addr := s.http.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.log.Println("[HTTP服务" + cast.ToString(s.cfg.Machine().Port) + "]已初始化")
	err = s.http.Serve(ln)
	if err != nil {
		return err
	}
	return nil
}

func (s *server) stop() error {
	s.needStop = true
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if beforeCloseFn != nil {
		beforeCloseFn()
	}
	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}
	return nil
}

func NewServer(cfg config.Config, log logger.Logger) (Server, error) {
	util.ToSetPasswordKey(cfg.Server().PasswordKey)
	//将配置,日志,数据库,缓存,消息队列注入到ioc容器中
	ioc.Ioc().Register(ioc.KeyConfig, cfg)
	ioc.Ioc().Register(ioc.KeyLog, log)
	waitNetWork := cfg.Machine().WaitNetwork
	waitNetWorkFn := func(title string, fn func() (any, error)) (any, error) {
		for {
			res, err := fn()
			if !waitNetWork || err == nil {
				return res, err
			}
			log.Println(title + "失败, 等待5秒后重试")
			time.Sleep(5 * time.Second)
		}
	}
	// 初始化缓存
	c, err := waitNetWorkFn("初始化缓存", func() (any, error) {
		return cache.NewCache(cfg.Cachebase(), log)
	})
	if err != nil {
		return nil, err
	}
	cache := c.(cache.Cache)
	ioc.Ioc().Register(ioc.KeyCache, cache)
	// 初始化数据库
	d, err := waitNetWorkFn("初始化数据库", func() (any, error) {
		return database.NewDatabase(cfg.Database(), cache, log)
	})
	if err != nil {
		return nil, err
	}
	db := d.(database.Database)
	ioc.Ioc().Register(ioc.KeyDatabase, db)
	if cfg.Machine().NeedUpdate {
		err = database.AutoMigrate()
		if err != nil {
			return nil, err
		}
	}
	log.Write(func(lvl logger.Level, msg string, fields ...logger.Field) {
		tmpDb := db.Write()
		if tmpDb != nil {
			system.WriteLog(tmpDb, lvl, msg, fields...)
		}
	})
	// 初始化消息队列
	q, err := waitNetWorkFn("初始化消息队列", func() (any, error) {
		return queue.NewQueue(cfg.Queuebase(), cache, log, plugin.NewUserQueueCallback(log))
	})
	if err != nil {
		return nil, err
	}
	queue := q.(queue.Queue)
	ioc.Ioc().Register(ioc.KeyQueue, queue)
	distributedProvider := uioc.Get[distributed.DistributedProvider](ioc.KeyDistributedProvider)
	err = distributedProvider.Init(distributed.RuntimeContext{
		Config:   cfg,
		Database: db,
		Cache:    cache,
		Queue:    queue,
		Log:      log,
	})
	if err != nil {
		return nil, err
	}
	dbDriver, cacheDriver, queueDriver := "", "", ""
	if cfg.Database() != nil {
		dbDriver = cfg.Database().Driver
	}
	if cfg.Cachebase() != nil {
		cacheDriver = cfg.Cachebase().Driver
	}
	if cfg.Queuebase() != nil {
		queueDriver = cfg.Queuebase().Driver
	}
	log.Info("distributed runtime prepared",
		logger.NewField("enabled", cfg.Distributed() != nil && cfg.Distributed().Enabled),
		logger.NewField("editionSupport", distributedProvider.SupportDistributed()),
		logger.NewField("databaseDriver", dbDriver),
		logger.NewField("databaseSupport", db.SupportDistributed()),
		logger.NewField("cacheDriver", cacheDriver),
		logger.NewField("cacheSupport", cache.SupportDistributed()),
		logger.NewField("queueDriver", queueDriver),
		logger.NewField("queueSupport", queue.SupportDistributed()))
	if err = distributedProvider.ValidateRuntime(); err != nil {
		log.Error("distributed runtime validation failed", logger.NewField("err", err))
		return nil, err
	}
	if err = distributedProvider.Start(); err != nil {
		log.Error("distributed runtime start failed", logger.NewField("err", err))
		return nil, err
	}
	log.Info("distributed runtime started",
		logger.NewField("enabled", distributedProvider.Enabled()),
		logger.NewField("localMachineID", distributedProvider.Nodes().LocalMachineID()),
		logger.NewField("localIp", distributedProvider.Nodes().LocalIp()))
	// 初始化服务数据源
	store, err := store.NewStore(cfg.Server(), log)
	if err != nil {
		return nil, err
	}
	ioc.Ioc().Register(ioc.KeyStore, store)
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	// 初始化 server
	s := &server{
		cfg:    cfg,
		engine: gin.Default(),
		log:    log,
	}
	// 初始化中间件 - CORS
	if s.cfg.CORS().Enabled {
		s.middlewares.cors = middleware.NewCORS(cfg.CORS(), log)
		s.engine.Use(s.middlewares.cors.Middleware())
	}
	// 初始化中间件 - 速率限制
	if s.cfg.RateLimit().Enabled {
		s.middlewares.rateLimit = middleware.NewRateLimiter(cfg.RateLimit(), log)
		s.engine.Use(s.middlewares.rateLimit.Middleware())
	}
	// 初始化中间件 - 用户信息
	s.middlewares.user = middleware.NewUser(s.cfg.Auth(), log)
	s.engine.Use(s.middlewares.user.Middleware())
	ioc.Ioc().Register(ioc.KeyMiddlewareAuth, s.middlewares.user)
	// 初始化中间件 - 日志
	if s.cfg.Server().RequestLog {
		s.middlewares.requestLog = middleware.NewRequestLog(s.log)
		s.engine.Use(s.middlewares.requestLog.Middleware())
	}
	//Controller导火线,通过导火线触发init函数进行依赖注入
	fuse.ControllerFuse()
	//构建Auth WebdavAuth S3Auth 中间件
	auth := middleware.NewAuth(log)
	webdavAuth := middleware.NewWebdavAuth(log)
	s3Auth := middleware.NewS3Auth(log)
	ioc.Ioc().Register(ioc.KeyMiddlewareWebdavAuth, webdavAuth)
	ioc.Ioc().Register(ioc.KeyMiddlewareS3Auth, s3Auth)
	var middlewareMap sync.Map
	middlewareMap.LoadOrStore(auth.Middleware())
	middlewareMap.LoadOrStore(webdavAuth.Middleware())
	middlewareMap.LoadOrStore(s3Auth.Middleware())
	//初始化路由
	initRouter(s.engine, &middlewareMap)
	// 初始化健康检查
	s.engine.GET("/health", s.healthCheck)
	// 初始化HTTP服务
	s.http = &http.Server{
		Addr:           "0.0.0.0:" + strconv.Itoa(s.cfg.Machine().Port),
		Handler:        s.engine,
		ReadTimeout:    s.cfg.Server().ReadTimeout,
		WriteTimeout:   s.cfg.Server().WriteTimeout,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	//注册重启函数
	ioc.Ioc().Register(ioc.KeyRestartFunc, s.restart)
	//注册插件管理
	ioc.Ioc().Register(ioc.KeyPluginManage, manager.NewPluginManage(s.cfg.Machine().Port, cfg, cache, log))
	s.removeScript()
	// 启动后回调函数
	afterFuncs := uioc.AfterFuncs()
	if afterFuncs != nil {
		for _, fn := range afterFuncs {
			fn.(func())()
		}
		ioc.Ioc().Unregister(ioc.KeyAfterFunc)
	}
	// 启动后的定时执行
	s.sch = plugin.NewTimeTaskSchedule()
	s.sch.Start()
	return s, nil
}

func (s *server) removeScript() {
	distDir := util.RunDir()
	switch runtime.GOOS {
	case "windows":
		batPath := filepath.Join(distDir, "update.bat")
		os.Remove(batPath)
	default: // linux,macos
		scriptPath := filepath.Join(distDir, "update.sh")
		os.Remove(scriptPath)
	}
}

func (s *server) restart(restart bool) {
	if restart {
		beforeCloseFn = func() {
			exePath := util.RunFile()
			if exePath == "" {
				return
			}
			fileName := filepath.Base(exePath)
			distDir := filepath.Dir(exePath)
			switch runtime.GOOS {
			case "windows":
				batPath := filepath.Join(distDir, "update.bat")
				batContent := `:while
if exist ` + fileName + `update (
   del ` + fileName + `
   rename ` + fileName + `update ` + fileName + `
   goto :while
)
start "" "` + fileName + `"
exit
`
				os.WriteFile(batPath, []byte(batContent), 0666)
				cmd := exec.Command("cmd.exe", "/C", "update.bat")
				cmd.Dir = distDir
				fn := uioc.HideCmdFunc()
				if fn != nil {
					fn(cmd)
				}
				cmd.Start()
			default: // linux,macos
				scriptPath := filepath.Join(distDir, "update.sh")
				// 生成 Shell 脚本
				scriptContent := `#!/bin/sh
chmod +x ` + fileName + `
exec ./` + fileName + `
exit 0
`
				os.WriteFile(scriptPath, []byte(scriptContent), 0744)
				cmd := exec.Command("sh", "update.sh")
				cmd.Dir = distDir
				cmd.Start()
			}
		}
	}
	closeFn := uioc.SystemCloseFunc()
	if closeFn != nil {
		closeFn()
	} else {
		s.Close()
	}

}

//go:embed all:web
var webFiles embed.FS

func initRouter(engine *gin.Engine, middlewareMap *sync.Map) {
	engine.Use(proxyFilter())
	routerMap := uioc.ControllerMap()
	ioc.Ioc().Unregister(ioc.KeyControllerMap)
	// 收集带权限元数据的路由，用于同步到 auth_registry
	var authRoutes []*ioc.RouterStruct
	routerMap.Range(func(key, value interface{}) bool {
		gr := value.(*ioc.GroupRoutes)
		gr.Mu.Lock()
		defer gr.Mu.Unlock()
		wareMap := make(map[string]gin.IRoutes)
		keyStr := key.(string)
		for _, r := range gr.Routes {
			tmp, ok := wareMap[keyStr+r.Ware]
			if !ok {
				tmp = engine.Group(keyStr)
				if r.Ware != "" {
					ware, ok := middlewareMap.Load(r.Ware)
					if ok {
						tmp.Use(ware.(gin.HandlerFunc))
					}
				}
				wareMap[keyStr+r.Ware] = tmp
			}
			tmp.Handle(r.Method, r.Path, r.Handler)
			if r.AuthName != "" && r.AuthFeature != "" && r.AuthLabel != "" {
				authRoutes = append(authRoutes, r)
			}
		}
		return true
	})
	routerMap.Clear()
	middlewareMap.Clear()
	// 自动同步内核权限到 auth_registry
	uioc.Get[infaceUser.UserProvider](ioc.KeyUserProvider).SyncKernelAuths(authRoutes)
	//处理静态数据
	subFS, _ := fs.Sub(webFiles, "web")
	wrappedFS := &customFS{fs: http.FS(subFS)}
	engine.Use(static.Serve("/", wrappedFS))
	path := util.RunDir()
	engine.Use(static.Serve("/api/fonts", static.LocalFile(filepath.Join(path, "fonts"), false)))
}

func (s *server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"version": s.cfg.Machine().Version,
	})
}
