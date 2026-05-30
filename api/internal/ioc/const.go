package ioc

const (
	KeyControllerMap = "controllerRouterMap" // 控制器路由映射, 运行前期有效, 运行后勿用
	KeyModelMap      = "modelList"           // 模型映射, 运行前期有效, 运行后勿用
	KeyPluginDataMap = "pluginDataMap"       // 插件数据映射, 运行前期有效, 运行后勿用
	KeyAfterFunc     = "afterFunc"           // 启动后回调函数, 运行前期有效, 运行后勿用
)

const (
	KeyConfig                  = "config"                  // 配置(通用)
	KeyLog                     = "log"                     // 日志(通用)
	KeyDatabase                = "database"                // 数据库(通用)
	KeyCache                   = "cache"                   // 缓存(通用)
	KeyQueue                   = "queue"                   // 消息队列(通用)
	KeyStore                   = "store"                   // 服务数据源(通用)
	KeyRestartFunc             = "restartFunc"             // 重启函数(通用)
	KeyPluginManage            = "pluginManage"            // 插件管理器(通用)
	KeyMiddlewareAuth          = "middlewareAuth"          // 认证中间件(通用)
	KeyMiddlewareWebdavAuth    = "middlewareWebdavAuth"    // webdav认证中间件(通用)
	KeyMiddlewareS3Auth        = "middlewareS3Auth"        // s3认证中间件(通用)
	KeyUserProvider            = "userProvider"            // 用户模块提供者(通用)
	KeyTimerFunc               = "timerFunc"               // 定时任务函数(通用)
	KeyDistributedProvider     = "distributedProvider"     // 分布式提供者(通用)
	KeyOfficialLicenseProvider = "officialLicenseProvider" // 官方授权提供者(通用)
	KeyLocalBillingProvider    = "localBillingProvider"    // 本地系统会员提供者(通用)

	KeyHideCmdFunc     = "hideCmdFunc"     // 隐藏命令函数(仅限win下有效)
	KeySystemCloseFunc = "systemCloseFunc" // app关闭函数(仅限GUI情况下有效)

	KeyThirdPluginMap        = "thirdPluginMap"        // 三方插件映射(分布式专享)
	KeyWebsocketMap          = "websocketMap"          // websocket映射(分布式专享)
	KeyWebsocketUserResolver = "websocketUserResolver" // websocket本机用户解析器
	KeyChatWebsocketHandler  = "chatWebsocketHandler"  // websocket聊天处理器
)
