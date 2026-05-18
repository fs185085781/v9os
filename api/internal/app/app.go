package app

import (
	"fmt"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	app2 "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/controller/api"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/server"
	"github.com/fs185085781/v9os/pkg/util"
)

type App interface {
	StartSync() error
	Close()
}
type appStruct struct {
	cfg    config.Config
	server server.Server
	log    logger.Logger
	app    fyne.App
}

func (app *appStruct) Close() {
	go func() {
		if app.server != nil {
			app.server.Close()
		}
		if app.app != nil {
			app.app.Quit()
		}
	}()
}

func (app *appStruct) StartSync() error {
	appTitle := "V9OS"
	appContact := "QQ185085781"
	guiSuccess := false
	// 1. 启动失败进行启动Server
	defer func() {
		// GUI 失败时同步启动服务
		if !guiSuccess {
			app.server.StartSync()
		}
	}()
	// 2. 捕获 GUI 初始化 panic
	defer func() {
		if r := recover(); r != nil {
			app.log.Debug("GUI初始化异常:", logger.NewField("error", r))
		}
	}()
	// 3. 启动 GUI
	a := app2.New()
	w := a.NewWindow(appTitle)
	w.Resize(fyne.NewSize(420, 290))
	w.CenterOnScreen()
	w.SetFixedSize(true)
	// 4. 顶部居中欢迎语
	title := widget.NewLabel("欢迎使用" + appTitle + ", 客户服务联系 " + appContact)
	title.Alignment = fyne.TextAlignCenter
	titleContainer := container.NewHBox(layout.NewSpacer(), title, layout.NewSpacer())
	// 5. 配置信息显示
	textData := binding.NewString()
	textData.Set(fmt.Sprintf("端口:%d   缓存:%s", app.cfg.Machine().Port, app.cfg.Cachebase().Driver))
	textData2 := binding.NewString()
	textData2.Set(fmt.Sprintf("数据库:%s   队列:%s", app.cfg.Database().Driver, app.cfg.Queuebase().Driver))
	configLabel := widget.NewLabelWithData(textData)
	configLabel.Alignment = fyne.TextAlignCenter
	configLabel2 := widget.NewLabelWithData(textData2)
	configLabel2.Alignment = fyne.TextAlignCenter
	// 6. 双按钮水平布局
	openBtn := widget.NewButton("打开首页", func() {
		url := fmt.Sprintf("http://127.0.0.1:%d", app.cfg.Machine().Port)
		app.openFile(url)
	})
	hideBtn := widget.NewButton("隐藏到托盘区", func() {
		w.Hide() // 隐藏主窗口
	})

	//7.容器面板
	btnContainer := container.NewHBox(
		layout.NewSpacer(),
		openBtn,
		layout.NewSpacer(),
		hideBtn,
		layout.NewSpacer(),
	)

	//8. 整合所有组件
	content := container.NewVBox(
		layout.NewSpacer(),
		titleContainer,
		layout.NewSpacer(),
		configLabel,
		layout.NewSpacer(),
		configLabel2,
		layout.NewSpacer(),
		btnContainer,
		layout.NewSpacer(),
	)
	w.SetContent(content)
	//9. 初始化托盘区
	var desk desktop.App
	var ok bool
	if desk, ok = a.(desktop.App); ok {
		// 创建托盘菜单
		menu := fyne.NewMenu(appTitle,
			fyne.NewMenuItem("显示主界面", func() {
				w.Show()
			}),
			fyne.NewMenuItem("打开运行目录", func() {
				app.openRunDir()
			}),
			fyne.NewMenuItem("重启服务", func() {
				f := uioc.RestartFunc()
				if f == nil {
					return
				}
				f(true)
			}),
			fyne.NewMenuItemSeparator(), // 分隔线
			fyne.NewMenuItem("退出", func() {
				app.Close()
			}),
		)
		desk.SetSystemTrayMenu(menu)
	}
	util.Go(func() {
		jc := 0
		for {
			if jc > 30 {
				break
			}
			jc++
			wset := api.WebSettingsGet()
			if wset == nil {
				time.Sleep(time.Second)
				continue
			}
			if wset.Logo == "" {
				time.Sleep(time.Second)
				continue
			}
			url := fmt.Sprintf("http://127.0.0.1:%d%s", app.cfg.Machine().Port, wset.Logo)
			if strings.HasPrefix(wset.Logo, "http") {
				url = wset.Logo
			}
			icon, err := fyne.LoadResourceFromURLString(url)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			w.SetIcon(icon)
			if desk != nil {
				desk.SetSystemTrayIcon(icon)
			}
			break
		}
	})
	// 窗口关闭时终止服务
	w.SetCloseIntercept(func() {
		app.Close()
	})
	ioc.Ioc().Register(ioc.KeyHideCmdFunc, app.hideCmdFunc)
	ioc.Ioc().Register(ioc.KeySystemCloseFunc, func() {
		app.Close()
	})
	guiSuccess = true
	app.app = a
	app.server.StartAsync(func(err error) {
		if err != nil {
			app.log.Debug("启动服务失败:", logger.NewField("error", err))
		}
		app.Close()
	})
	w.ShowAndRun()
	return nil
}
func (app *appStruct) hideCmdFunc(cmd *exec.Cmd) {
	if runtime.GOOS != "windows" {
		return
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	sysProcAttrValue := reflect.ValueOf(cmd.SysProcAttr).Elem()
	hideWindowField := sysProcAttrValue.FieldByName("HideWindow")
	if !hideWindowField.IsValid() {
		return
	}
	if !hideWindowField.CanSet() {
		return
	}
	if hideWindowField.Kind() != reflect.Bool {
		return
	}
	hideWindowField.SetBool(true)
}

func (app *appStruct) openRunDir() {
	app.openFile(util.RunDir())
}

func (app *appStruct) openFile(pathOrurl string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		if strings.HasPrefix(pathOrurl, "http") {
			cmd = exec.Command("cmd", "/c", "start", pathOrurl)
		} else {
			cmd = exec.Command("explorer", pathOrurl)
		}
	case "darwin":
		cmd = exec.Command("open", pathOrurl)
	default:
		cmd = exec.Command("xdg-open", pathOrurl)
	}
	_ = cmd.Start()
}

func NewApp(cfg config.Config, log logger.Logger) (App, error) {
	server, err := server.NewServer(cfg, log)
	if err != nil {
		return nil, err
	}
	return &appStruct{
		cfg:    cfg,
		log:    log,
		server: server,
	}, nil
}
