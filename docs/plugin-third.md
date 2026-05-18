# 第三方插件开发

[返回总入口](starts.md) | [主程序插件开发](plugin-main.md) | [前端插件开发](plugin-web.md) | [远程应用](remote-app.md)

第三方插件是 `PluginType = 3` 的插件。它用于包装任何可独立运行的 Web 程序，例如 Node、Java、Python、Go、Rust 或已有商业系统的本地服务。

## 一、核心规则

第三方插件由 V9OS 管理启动和停止，但真实 Web 服务由你的程序提供。内核需要：

- `index.json`：声明插件编码、类型、版本和启动端口。
- `restart.bat` / `restart.sh`：启动或重启第三方服务。
- `stop.bat` / `stop.sh`：停止第三方服务。
- 插件表中的 `AccessUrl`：第三方服务访问地址。

内核访问 `/api/thirdplugin/{code}` 时，会确保插件进程已启动，然后重定向到 `AccessUrl`。

## 二、目录结构

运行目录建议：

```text
plugins/third/my_third_app/
├── index.json
├── restart.bat
├── restart.sh
├── stop.bat
├── stop.sh
├── logo.png
└── app/
    └── 你的第三方程序
```

Windows 下使用 `.bat`，Linux/macOS 下使用 `.sh`。内核会按当前系统检查对应脚本是否存在。

## 三、index.json

最小示例：

```json
{
  "Name": "我的第三方应用",
  "Description": "包装已有 Web 程序",
  "Code": "my_third_app",
  "Version": "1.0.0",
  "PluginType": "3",
  "ThirdPort": "18080",
  "LimitVersion": "1.0.0",
  "NeedLogin": "1",
  "IconUrl": "/api/appstore/img/my_third_app",
  "Log": "首发"
}
```

`ThirdPort` 是内核等待服务就绪时检查的端口。安装本地包时，如果安装入口提供了访问源地址，内核会用 `AccessOrigin + ":" + ThirdPort` 写入插件表的 `AccessUrl`。

## 四、启动脚本

`restart.bat` 示例：

```bat
@echo off
cd /d %~dp0
start "" app\my-server.exe --port 18080
```

`stop.bat` 示例：

```bat
@echo off
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :18080') do taskkill /F /PID %%a
```

`restart.sh` 示例：

```bash
#!/usr/bin/env sh
cd "$(dirname "$0")"
nohup ./app/my-server --port 18080 > ./app.log 2>&1 &
```

`stop.sh` 示例：

```bash
#!/usr/bin/env sh
PID=$(lsof -ti :18080)
if [ -n "$PID" ]; then
  kill "$PID"
fi
```

Linux/macOS 下内核会给 `restart.sh` 设置可执行权限后启动。

## 五、插件表配置

必须存在同编码插件记录：

| 字段 | 示例 | 说明 |
| --- | --- | --- |
| `Code` | `my_third_app` | 与目录名、`index.json` 一致 |
| `PluginType` | `3` | 第三方插件 |
| `Status` | `1` | 启用 |
| `Version` | `1.0.0` | 版本 |
| `AccessUrl` | `http://127.0.0.1:18080` | 第三方服务真实地址 |
| `FirstMachine` | 当前机器 ID | 分布式时指定由哪台机器运行 |
| `CloseDelay` | `0` | 0 常驻；大于 0 表示空闲延迟关闭 |

分布式模式下，第三方插件只适合在指定机器运行。内核会根据机器 ID、插件白名单和节点解析结果决定是否本机启动或转发启动请求。

## 六、打包建议

- 包根目录必须包含 `index.json` 和当前系统需要的启动/停止脚本。
- 不要把端口写死到多个地方后忘记同步；`index.json`、脚本、应用配置和 `AccessUrl` 要一致。
- 服务启动后 30 秒内端口必须可连接，否则内核会认为启动失败并写入 `runtime_error`。
- `stop` 脚本应可重复执行，服务未运行时也应正常退出。

## 七、和远程应用的区别

如果你的系统已经部署在公网或内网固定 URL 上，并不需要 V9OS 管理进程，使用 [远程应用](remote-app.md) 更简单。第三方插件适合“随 V9OS 本地启动、停止、升级”的程序。

