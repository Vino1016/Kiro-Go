# Codex 使用 Kiro 模型与积分：安装与配置手册

> 文档版本：2026-08-12  
> 适用平台：macOS 12+（Intel / Apple Silicon）  
> 指定代码：[Vino1016/Kiro-Go `main_vino` 分支](https://github.com/Vino1016/Kiro-Go/tree/main_vino)  
> 本次验证版本：Kiro-Go `795b2ca`、CC Switch 3.19.2  
> 原理说明：[架构与实现原理](./codex-kiro-go-architecture.md)

## 可以使用安装 Skill

仓库已经提供可被 Codex 和 Kiro 共用的开放 Agent Skill：

```text
https://github.com/Vino1016/Kiro-Go/tree/main_vino/skills/kiro-go-installer
```

Skill 会先识别当前是 Codex 辅助还是 Kiro 辅助，再要求选择 Kiro-Go 的部署方式：

1. macOS 本机 Go + launchd（推荐）；
2. Docker 本地构建。

它不会同时启动两种部署，也不会自动覆盖已有仓库、容器、LaunchAgent、配置或 8080 端口进程。

### 在 Codex 中安装和使用

让 Codex 使用内置 Skill 安装器安装上面的 GitHub Skill 子目录，例如：

```text
使用 $skill-installer 安装这个 Skill：
https://github.com/Vino1016/Kiro-Go/tree/main_vino/skills/kiro-go-installer
```

安装后新建任务或重启 Codex，再执行：

```text
使用 $kiro-go-installer 帮我安装并配置 Codex 使用 Kiro 模型和积分
```

Codex 模式使用终端完成安装，通过 Browser Use 操作 Kiro-Go 网页后台，通过 Computer Use 操作 CC Switch 原生应用。密码、SSO 和 API Key 仍由本人输入。

### 在 Kiro 中安装和使用

在 Kiro 的 “Agent Steering & Skills” 中点击 `+` → “Import a skill”，选择 GitHub，并粘贴上面的 Skill 子目录地址。也可以从已经克隆的仓库选择本地 `skills/kiro-go-installer` 文件夹。

导入后使用：

```text
/kiro-go-installer
```

Kiro 模式负责命令行安装和验证；Kiro-Go 网页后台、CC Switch、SSO、密码和 API Key 由本人按 Skill 引导手动配置。Kiro 的 Skill 导入与目录约定参考 [Kiro Agent Skills 官方文档](https://kiro.dev/docs/skills/)。

## 不安装 Skill：使用 Codex 提示词

如果已经安装 Codex Desktop，可以把下面的提示词和本手册一起发给 Codex。Codex 可以通过终端完成安装和服务管理，通过 Browser Use 操作 Kiro-Go 网页管理后台，通过 Computer Use 操作 CC Switch 原生应用，适合完成接近一键式的安装与配置。

登录、SSO、管理密码和 API Key 属于敏感信息，仍应由本人在界面中输入或复制。Codex 应在这些步骤暂停，等本人完成后继续，不应读取、复述或保存凭证。

### 发给 Codex 的提示词

```text
请严格按照我一同提供的《Codex 使用 Kiro 模型与积分：安装与配置手册》，直接在当前 macOS 本机完成 Kiro-Go 和 CC Switch 的安装、配置与验收。

这是一个执行任务，不要只给我命令或操作说明。请按下面的工具分工主动完成能够自动完成的步骤，持续推进到真正可用；只有部署方式、系统授权、账号登录、密码和 API Key 等必须由我决定或输入的步骤才暂停询问：

- 终端：环境检查、Git、构建、测试、Docker、launchd、配置备份和接口验证。
- Browser Use：访问和操作 Kiro-Go 本地网页管理后台，以及需要浏览器完成的 Kiro SSO 页面。
- Computer Use：安装后操作 CC Switch 等 macOS 原生图形应用。

不要使用 Computer Use 操作 Kiro-Go 管理后台；它是网页，应优先使用 Browser Use。不要使用 Browser Use 操作 CC Switch；它是原生应用，应使用 Computer Use。

一、执行原则

1. 开始时先检查 macOS 架构、Homebrew、Git、Docker、Go、8080 端口、现有 Kiro-Go 进程、CC Switch、Codex 配置和当前工作目录。
2. 保留所有已有配置和数据。不要删除 data 目录，不要覆盖未知配置，不要卸载现有软件，不要擅自停止无关进程。
3. 对停止服务、替换容器、覆盖文件、修改 ~/.codex 配置等可能影响现有环境的操作，先备份并说明影响。
4. 不要让我逐条复制执行终端命令；命令由你直接执行，并根据真实输出继续排查和验证。
5. 不要向我索取、打印、记录、提交或回显 Kiro 登录凭证、管理密码、API Key、Token 和公司敏感配置。需要输入敏感信息时，把界面留给我操作，等我回复“已完成”后继续。
6. 每个阶段结束时给出简短进度，未验证的步骤不能标记为完成。

二、部署 Kiro-Go

1. Kiro-Go 必须使用：
   https://github.com/Vino1016/Kiro-Go/tree/main_vino
2. 必须确认当前分支是 main_vino，并包含提交 795b2ca。不要使用 Quorinex 上游旧代码、旧镜像或其他分支。
3. 如果我没有指定部署方式，只问我一次：
   - Docker 本地构建；
   - macOS 原生 Go + launchd（推荐，不依赖 Docker）。
   如果我回复“随便”或“推荐哪个”，默认使用 macOS 原生方式。
4. 两种方式只能选择一种，不能让 Docker 和原生服务同时占用 8080。
5. Docker 方式必须从 main_vino 本地构建镜像，并只映射 127.0.0.1:8080。
6. 原生方式必须安装或确认 Go 1.21+、执行 go test ./...、构建 kiro-go，并使用仓库 deploy/macos 下的 launchd 模板配置登录自启和异常重启。
7. 完成后验证分支、提交、进程、监听地址、日志和 http://127.0.0.1:8080/health。

三、配置 Kiro-Go

1. 使用 Browser Use 打开并操作 http://127.0.0.1:8080/admin，不要使用 Computer Use 操作这个网页后台。
2. 到需要输入默认管理密码、设置新密码时暂停，让我本人操作。不要读取或记录密码。
3. 帮我导航到“账号”→“添加账号”，根据我选择的账号类型继续；需要浏览器 SSO 登录、授权或选择企业 Profile 时暂停，让我本人完成。
4. 登录完成后继续测试账号并刷新模型，确认账号正常且已启用。
5. 帮我导航到“API”，创建名为 codex-local 的本地网关 Key，并开启 API Key 验证。Key 出现时暂停，让我复制到密码管理器；不要读取、复述或写入日志。
6. 由我把 Key 粘贴到需要的位置。随后验证未带 Key 的请求返回 401，并在不暴露 Key 的前提下验证带 Key 的 Responses 请求成功。

四、安装和配置 CC Switch

1. 先备份 ~/.codex/config.toml 和 ~/.codex/auth.json，再使用 Homebrew 安装或更新 CC Switch。
2. 使用 Computer Use 打开并操作 CC Switch 的 Codex 页签，保留自动导入的 default Provider 和 Codex 官方认证，不要删除或覆盖官方登录。
3. 在设置中：
   - 开启“切换第三方 Provider 时保留 Codex 官方认证”；
   - 关闭 Local Routing；
   - 可选开启统一 Codex 会话历史；
   - 保留现有 MCP、插件、Hooks 和项目配置。
4. 新建 Provider：
   - 名称：Kiro-Go Local
   - Base URL：http://127.0.0.1:8080/v1
   - API 格式：Responses（原生）
   - 默认实际模型：gpt-5.6-sol
   - 应用通用配置：开启
   - Local Routing：关闭
5. 到 API Key 输入框时暂停，让我本人从密码管理器粘贴 codex-local Key。不要读取或回显 Key。
6. 添加模型映射：
   - Claude Opus 5 → claude-opus-5 → 1,000,000
   - Claude Sonnet 5 → claude-sonnet-5 → 1,000,000
   - Claude Opus 4.8 → claude-opus-4.8 → 1,000,000
   - gpt-5.6-sol-kiro → gpt-5.6-sol → 200,000
   - gpt-5.6-terra-kiro → gpt-5.6-terra → 200,000
7. 保存后检查 ~/.codex/config.toml，确认原来的 MCP、插件和 Hooks 未丢失；发现差异时只恢复缺失部分，不要覆盖已经正确的 Kiro Provider 配置。

五、切换和验收

1. 在 CC Switch 中启用 Kiro-Go Local。
2. 当前任务运行在 Codex 内，不要直接退出 Codex 导致任务中断。先确认所有安装和界面配置已保存，再明确告诉我需要 Cmd+Q 完全退出并重新打开 Codex。
3. 重启后让我新建任务，并使用 /model 确认能看到 gpt-5.6-sol-kiro 等映射模型。
4. 指导我执行一个只读工具调用测试，确认 Shell、文件和 MCP 工具可用。
5. 检查 Kiro-Go 后台的请求数、tokens 或 credits 是否增加。
6. 再验证 CC Switch 可以切回 default；不要仅根据模型显示名判断请求来源。
7. 最终输出验收结果，分别标记：已自动完成、已由我手动完成、验证通过、尚未验证。不要把“已配置”与“已验证可用”混为一谈。

如果 Browser Use 不可用，就逐项引导我手动操作 Kiro-Go 网页后台；如果 Computer Use 不可用，就逐项引导我手动操作 CC Switch。任何一种界面工具不可用时，都不要假装已经点击或配置成功。继续完成所有终端步骤，我每完成一个手动步骤，你再验证并进入下一项。
```

### Codex 与本人操作边界

| 环节 | Codex 自动完成 | 需要本人配合 |
| --- | --- | --- |
| 环境与备份 | 检查环境、端口和现有配置，创建备份 | 批准必要的系统权限 |
| Kiro-Go 部署 | 拉取代码、测试、构建、Docker 或 launchd 部署、健康检查 | 选择部署方式 |
| Kiro-Go 后台 | 使用 Browser Use 导航、测试账号、刷新模型、开启鉴权 | 输入密码、完成 SSO、选择企业 Profile、保存 API Key |
| CC Switch | 使用 Computer Use 操作界面、添加 Provider 和模型映射、检查配置差异 | 粘贴本地网关 Key，确认敏感设置 |
| Codex 验收 | 准备测试、检查服务和配置、核对后台计量 | 重启 Codex、新建任务并确认最终效果 |

## 不安装 Skill：使用 Kiro 提示词

可以把下面的提示词和本手册一起发送给 Kiro，让 Kiro 通过终端完成 Kiro-Go 与 CC Switch 的安装、构建和基础验证。

> Kiro 通常没有 Codex Desktop 的 Browser Use 和 Computer Use 能力，无法可靠操作 Kiro-Go 网页管理后台和 CC Switch 原生应用。因此它只能完成命令行安装；账号登录、API Key 创建、Provider 和模型映射等实际配置仍需本人按照本手册手动完成。

### 发给 Kiro 的提示词

```text
请严格按照我一同提供的《Codex 使用 Kiro 模型与积分：安装与配置手册》，在当前 macOS 本机安装 Kiro-Go 和 CC Switch。

执行要求：

1. 先检查 macOS 架构、Homebrew、Git、Docker、Go、8080 端口占用和当前工作目录，不要覆盖已有配置或删除已有数据。
2. Kiro-Go 必须使用下面这个仓库的 main_vino 分支：
   https://github.com/Vino1016/Kiro-Go/tree/main_vino
3. 必须确认当前代码包含提交 795b2ca。不要使用 Quorinex 上游旧代码或旧镜像。
4. 如果我没有指定部署方式，先让我在以下两种方式中选择一种：
   - Docker：从 main_vino 本地构建镜像，仅映射到 127.0.0.1:8080。
   - macOS 原生：安装 Go、执行 go test ./...、构建二进制，并使用仓库提供的 launchd 模板配置登录自启。
5. Docker 和 macOS 原生版本不能同时运行，不要让两个服务同时占用 8080。
6. 使用 Homebrew 安装 CC Switch。只完成应用安装，不要直接改写 ~/.codex/config.toml 或 ~/.codex/auth.json。
7. 安装过程中遇到停止服务、替换容器、覆盖文件或其他可能影响现有环境的操作时，先说明影响并征得我的确认。
8. 不要向我索取、打印、记录或提交任何 Kiro 登录凭证、管理密码、API Key、Token 或公司敏感配置。
9. 安装结束后验证：
   - 当前分支为 main_vino；
   - 包含提交 795b2ca；
   - Kiro-Go 健康检查可以访问；
   - 服务仅允许本机访问；
   - CC Switch 已安装；
   - 所选部署方式能够正常启动并可按手册维护。
10. 你没有 Browser Use、Computer Use 或其他可靠的界面操作能力。不要尝试代替我点击 Kiro-Go 管理后台、浏览器 SSO 登录、CC Switch 或 Codex 界面。命令行安装和验证完成后停止，并逐项提醒我手动完成：
    - 登录 Kiro-Go 管理后台并修改默认管理密码；
    - 添加 Kiro 账号，完成浏览器登录并选择正确的企业 Profile；
    - 测试账号并刷新模型；
    - 创建 codex-local 本地网关 API Key，并开启 API Key 验证；
    - 在 CC Switch 中保留 default Provider 和官方认证；
    - 新建 Kiro-Go Local Provider，填写 Base URL、API Key 和 Responses 原生格式；
    - 添加本文档指定的模型映射，应用通用配置并关闭 Local Routing；
    - 启用 Kiro-Go Local，完全重启 Codex，再进行模型、工具调用和 credits 验收。

请按“环境检查 → 执行安装 → 命令行验证 → 输出手动配置清单”的顺序推进。每完成一项都说明结果；如果命令失败，先根据实际输出排查，不要跳过验证或宣称已经完成 GUI 配置。
```

### Kiro 与本人操作边界

| 环节 | Kiro 可以协助 | 需要本人操作 |
| --- | --- | --- |
| 环境检查 | 检查 Git、Homebrew、Docker、Go 和端口 | 处理系统权限弹窗 |
| Kiro-Go 安装 | 拉取指定分支、测试、构建、启动、配置 launchd 或 Docker | 选择 Docker 或原生部署方式 |
| Kiro-Go 基础验证 | 检查分支、提交、进程、端口和健康接口 | 确认公司账号使用合规 |
| Kiro-Go 后台 | 只能提供操作清单 | 登录后台、改密码、添加账号、完成 SSO、创建 API Key |
| CC Switch | 使用 Homebrew 安装应用 | 打开界面、添加 Provider、填写 Key、配置模型映射和通用配置 |
| Codex 验收 | 提供验收命令和检查项 | 切换 Provider、重启 Codex、选择模型并确认 credits 消耗 |

本手册供团队成员独立安装使用。Kiro-Go 提供两种部署方式，任选一种即可：

| 方式 | 适合场景 | 自启动方式 | 数据目录 |
| --- | --- | --- | --- |
| Docker | 已安装 Docker Desktop，希望运行环境隔离 | Docker `--restart unless-stopped` | 仓库 `data/` 挂载到容器 |
| macOS 原生 | 不想依赖 Docker，希望占用更少资源 | macOS launchd | 仓库 `data/` |

两种方式不能同时监听本机 8080 端口。部署完成后的 Kiro-Go 账号、API Key、CC Switch 和 Codex 配置步骤完全相同。

## 1. 准备源码与环境

### 1.1 通用前置条件

- Codex App 或 Codex CLI；
- 可用且获授权使用的 Kiro 账号；
- macOS 12+；
- Homebrew；
- Docker 方式安装 Docker Desktop；原生方式安装 Go 1.21+。

### 1.2 获取指定分支

新安装：

```bash
cd ~/IdeaProjects
git clone --branch main_vino --single-branch \
  https://github.com/Vino1016/Kiro-Go.git
cd Kiro-Go
```

已经克隆过仓库：

```bash
git remote set-url origin https://github.com/Vino1016/Kiro-Go.git
git fetch origin
git switch main_vino
git pull --ff-only origin main_vino
```

确认分支和修复提交：

```bash
git branch --show-current
git merge-base --is-ancestor 795b2ca HEAD && echo 'required fixes: OK'
git log -1 --oneline
```

必须看到当前分支为 `main_vino`，并输出 `required fixes: OK`。提交 `795b2ca` 修复了 Codex 工具传递和企业 IdC 响应终止判断问题；不要改用未包含该提交的上游旧镜像。

## 2. 方式 A：Docker 部署

本方式从当前 `main_vino` 源码本机构建镜像，确保镜像包含指定修复。

### 2.1 构建镜像

确认 Docker Desktop 已启动：

```bash
docker version
```

在 Kiro-Go 仓库根目录执行：

```bash
docker build -t kiro-go:main_vino .
mkdir -p data
```

### 2.2 启动容器

```bash
docker run -d \
  --name kiro-go-codex \
  -p 127.0.0.1:8080:8080 \
  -e CONFIG_PATH=/app/data/config.json \
  -v "$PWD/data:/app/data" \
  --restart unless-stopped \
  kiro-go:main_vino
```

这里显式使用 `127.0.0.1:8080:8080`，只允许本机访问。仓库自带的 `docker-compose.yml` 使用 `8080:8080`，会监听所有本机网络接口；如果没有先修改端口映射，不建议直接用于本机私有网关。

不要把 `ADMIN_PASSWORD` 或真实 API Key 直接写在 `docker run` 命令中，避免敏感信息进入 Shell 历史。

### 2.3 验证 Docker 服务

```bash
docker ps --filter name=kiro-go-codex
docker logs --tail 50 kiro-go-codex
curl -fsS http://127.0.0.1:8080/health
```

健康检查应返回类似：

```json
{"status":"ok","version":"1.1.5"}
```

管理后台地址：

```text
http://127.0.0.1:8080/admin
```

### 2.4 Docker 更新、停止与卸载

更新代码并重建：

```bash
git switch main_vino
git pull --ff-only origin main_vino
git merge-base --is-ancestor 795b2ca HEAD && echo 'required fixes: OK'
docker build -t kiro-go:main_vino .
docker stop kiro-go-codex
docker rm kiro-go-codex
docker run -d \
  --name kiro-go-codex \
  -p 127.0.0.1:8080:8080 \
  -e CONFIG_PATH=/app/data/config.json \
  -v "$PWD/data:/app/data" \
  --restart unless-stopped \
  kiro-go:main_vino
curl -fsS http://127.0.0.1:8080/health
```

账号和配置保存在仓库 `data/` 中，删除并重建容器不会清除它们。

停止服务：

```bash
docker stop kiro-go-codex
```

重新启动：

```bash
docker start kiro-go-codex
```

删除容器但保留配置数据：

```bash
docker stop kiro-go-codex
docker rm kiro-go-codex
```

除非明确要清除账号和配置，否则不要删除仓库 `data/` 目录。

完成本章后，跳到[第 4 章：配置 Kiro-Go](#4-配置-kiro-go)。

## 3. 方式 B：macOS 原生部署

本方式将 Kiro-Go 编译成本机二进制，并使用 launchd 实现登录自启和异常重启。

### 3.1 安装 Go、测试并构建

```bash
brew install go
go version
go test ./...
go build -o kiro-go .
```

项目要求 Go 1.21+。`go test ./...` 必须通过后再部署。

### 3.2 首次前台启动

```bash
mkdir -p data logs
CONFIG_PATH="$PWD/data/config.json" ./kiro-go
```

首次启动会生成 `data/config.json`。看到管理后台地址后，在启动它的终端按 `Ctrl+C` 停止服务。

首次生成的配置可能监听 `0.0.0.0`。只供本机 Codex 使用时，在添加账号前改为 `127.0.0.1`：

```bash
sed -i '' 's/"host": "0.0.0.0"/"host": "127.0.0.1"/' data/config.json
```

可以再次前台启动并验证：

```bash
CONFIG_PATH="$PWD/data/config.json" ./kiro-go
```

另开一个终端执行：

```bash
curl -fsS http://127.0.0.1:8080/health
```

确认成功后按 `Ctrl+C`，再配置 launchd。

### 3.3 安装 macOS LaunchAgent

仓库提供通用模板：

```text
deploy/macos/com.vino.kiro-go.plist.template
```

先确认 Docker 或其他进程没有占用 8080：

```bash
lsof -nP -iTCP:8080 -sTCP:LISTEN
```

如果之前启动过 Docker 版，应先停止：

```bash
docker stop kiro-go-codex
```

在 Kiro-Go 仓库根目录安装 LaunchAgent：

```bash
REPO_DIR="$PWD"
PLIST="$HOME/Library/LaunchAgents/com.vino.kiro-go.plist"

mkdir -p "$REPO_DIR/logs" "$HOME/Library/LaunchAgents"
sed "s|__KIRO_GO_DIR__|$REPO_DIR|g" \
  deploy/macos/com.vino.kiro-go.plist.template > "$PLIST"

chmod 600 "$PLIST"
plutil -lint "$PLIST"
launchctl bootout "gui/$(id -u)/com.vino.kiro-go" 2>/dev/null || true
launchctl bootstrap "gui/$(id -u)" "$PLIST"
launchctl enable "gui/$(id -u)/com.vino.kiro-go"
launchctl kickstart -k "gui/$(id -u)/com.vino.kiro-go"
```

模板会让服务登录后自动启动、异常退出后自动重启，并把日志写入仓库 `logs/`。

### 3.4 验证和维护 LaunchAgent

```bash
launchctl print "gui/$(id -u)/com.vino.kiro-go"
curl -fsS http://127.0.0.1:8080/health
tail -n 30 logs/kiro-go.log
tail -n 30 logs/kiro-go.error.log
```

更新 Kiro-Go：

```bash
git switch main_vino
git pull --ff-only origin main_vino
git merge-base --is-ancestor 795b2ca HEAD && echo 'required fixes: OK'
go test ./...
go build -o kiro-go .
launchctl kickstart -k "gui/$(id -u)/com.vino.kiro-go"
curl -fsS http://127.0.0.1:8080/health
```

停止并禁用：

```bash
launchctl bootout "gui/$(id -u)/com.vino.kiro-go"
```

重新启用：

```bash
launchctl bootstrap "gui/$(id -u)" \
  "$HOME/Library/LaunchAgents/com.vino.kiro-go.plist"
```

卸载 LaunchAgent，但保留程序和配置：

```bash
launchctl bootout "gui/$(id -u)/com.vino.kiro-go" 2>/dev/null || true
rm "$HOME/Library/LaunchAgents/com.vino.kiro-go.plist"
```

## 4. 配置 Kiro-Go

以下步骤适用于 Docker 和 macOS 原生两种部署方式。

### 4.1 修改管理密码

访问：

```text
http://127.0.0.1:8080/admin
```

默认密码是 `changeme`。首次登录后立即在“设置”中修改管理密码。不要把密码写入 README、Shell 脚本或 Git。

### 4.2 添加 Kiro 账号

进入“账号”→“添加账号”，按实际账号类型选择：

- IAM Identity Center（企业 SSO）；
- Microsoft Enterprise SSO；
- AWS Builder ID；
- Kiro API Key；
- 其他页面支持的凭证导入方式。

企业账号按页面提示完成浏览器登录；如果账号关联多个 Profile，选择实际使用的 Kiro Profile。

添加后确认：

- 状态为“正常”；
- 账号“已启用”；
- 账号测试成功；
- “刷新模型”能够完成，或已确认目标模型可实际请求。

### 4.3 创建供 Codex 使用的 Kiro-Go API Key

这里创建的是 **Kiro-Go 本地网关 Key**，不是 Kiro 账号的 `ksk_...` Key。

1. 打开 Kiro-Go 管理后台的“设置 -> API设置”。
2. 创建一个 Key，例如命名为 `codex-local`。
3. 启用该 Key。
4. 在设置中启用 API Key 验证。
5. 把生成的 Key 自己保存。

验证未认证请求应返回 401：

```bash
curl -i http://127.0.0.1:8080/v1/responses \
  -H 'Content-Type: application/json' \
  -d '{"model":"gpt-5.6-sol","input":"test"}'
```

使用本地网关 Key 验证 Responses：

```bash
read -s KIRO_GO_API_KEY
export KIRO_GO_API_KEY

curl -fsS http://127.0.0.1:8080/v1/responses \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $KIRO_GO_API_KEY" \
  -d '{"model":"gpt-5.6-sol","input":"只回复 KIRO_GO_OK","stream":false,"store":false}'

unset KIRO_GO_API_KEY
```

## 5. 安装 CC Switch

CC Switch 官方仓库：<https://github.com/farion1231/cc-switch>。

使用 Homebrew 安装：

```bash
brew install --cask cc-switch
```

更新：

```bash
brew upgrade --cask cc-switch
```

也可以从 [CC Switch Releases](https://github.com/farion1231/cc-switch/releases) 下载 macOS DMG。

## 6. 配置 CC Switch

### 6.1 首次使用前备份 Codex 配置

```bash
BACKUP_DIR="$HOME/.codex/backups/$(date +%Y%m%d-%H%M%S)-pre-cc-switch"
mkdir -p "$BACKUP_DIR"
cp "$HOME/.codex/config.toml" "$BACKUP_DIR/config.toml"
cp "$HOME/.codex/auth.json" "$BACKUP_DIR/auth.json"
chmod 600 "$BACKUP_DIR"/*
```

如果某个文件不存在，跳过对应 `cp` 即可。首次打开 CC Switch 时，它会把现有 Codex 配置导入为 `default` Provider。以后切回 OpenAI 官方 Codex 时，启用这个 `default` Provider。

### 6.2 推荐设置

打开 CC Switch 设置：

- 开启“切换第三方 Provider 时保留 Codex 官方认证”；
- 关闭“本地路由/代理接管”；
- 可选开启“统一 Codex 会话历史”；
- 不要删除自动导入的 `default` Provider。

`main_vino` 已原生提供 `/v1/responses`，因此本方案使用 Responses 原生直连。Local Routing 会多增加一层协议转换，不需要开启。

### 6.3 添加 `Kiro-Go Local` Provider

进入 Codex 页签，点击“添加 Provider”，填写：

| 字段 | 值 |
| --- | --- |
| 名称 | `Kiro-Go Local` |
| 备注 | `本机 Kiro-Go · Kiro ` |
| Base URL | `http://127.0.0.1:8080/v1` |
| API Key | Kiro-Go 后台创建的 `codex-local` Key |
| API 格式 | `Responses（原生）` |
| 默认实际模型 | `gpt-5.6-sol-kiro` |
| 应用通用配置 | 开启 |
| Local Routing | 关闭 |

如果界面提供“保留官方认证”或“使用现有 auth.json”，选择保留，避免第三方 Provider 覆盖原来的 OpenAI 登录。

### 6.4 配置模型映射

按当前团队需要添加：

| Codex 菜单显示名 | 实际请求模型 |   建议上下文窗口 |
| --- | --- |----------:|
| `Claude Opus 5` | `claude-opus-5` | 1,000,000 |
| `Claude Sonnet 5` | `claude-sonnet-5` | 1,000,000 |
| `Claude Opus 4.8` | `claude-opus-4.8` | 1,000,000 |
| `gpt-5.6-sol-kiro` | `gpt-5.6-sol` |   272,000 |
| `gpt-5.6-terra-kiro` | `gpt-5.6-terra` |    272,000 |

说明：

- GPT 显示名加 `-kiro` 仅用于识别额度来源；
- 最终以 Kiro 管理后台“刷新模型”得到的可用模型和 token limit 为准；
- Kiro-Go 对未命中别名的模型 ID 会透传，但账号不一定拥有该模型权限；
- 保存映射后必须完全退出并重启 Codex，因为 `model_catalog_json` 在 Codex 启动时加载。

### 6.5 保留 MCP、插件、Hooks 和项目配置 (如果原来就一直使用codex的开发人员才需要注意)

Provider 切换会改写 `~/.codex/config.toml`。在 CC Switch 中维护 Codex“通用配置”，并让 `Kiro-Go Local` 勾选“应用通用配置”。

切换后至少检查：

```bash
rg '^\[mcp_servers\.' "$HOME/.codex/config.toml"
rg '^\[plugins\.' "$HOME/.codex/config.toml"
rg '^\[\[hooks\.' "$HOME/.codex/config.toml"
```

部分 CC Switch 版本可能无法完整保留嵌套工具审批块，例如：

```toml
[mcp_servers.some_server.tools.some_tool]
approval_mode = "approve"
```

这通常不影响工具存在，但可能恢复为每次询问。发现差异时，从备份恢复对应配置；不要把真实 Token 复制进文档或聊天。

## 7. 在 Codex 中切换和使用

### 7.1 使用 Kiro 模型和积分

1. 确认 Kiro-Go 健康：

   ```bash
   curl -fsS http://127.0.0.1:8080/health
   ```

2. 在 CC Switch 的 Codex 页签启用 `Kiro-Go Local`。
3. 完全退出 Codex（`Cmd+Q`），再重新打开。
4. 新建任务，不要续接 OpenAI Provider 下正在执行的任务。
5. 使用 `/model` 选择 `gpt-5.6-sol-kiro` 或其他 Kiro 映射模型。
6. 发送一个需要工具的验收任务，例如：

   ```text
   列出当前仓库根目录下的文件，然后只总结文件类型，不修改任何文件。
   ```

7. 在 Kiro-Go 管理后台确认请求数、tokens 或 credits 增加。

当前 Provider 是 `Kiro-Go Local` 时，该 Provider 下的所有模型都发往 Kiro-Go。即使选择的模型名字与 OpenAI 官方模型相似，也不会使用 OpenAI 官方额度。

### 7.2 切回 OpenAI 官方 Codex

1. 在 CC Switch 启用 `default` Provider。
2. 完全退出并重启 Codex。
3. 新建任务并选择官方模型。

决定请求去向的是当前 `model_provider` 和 `base_url`，不是菜单中的模型显示名。

### 7.3 会话历史说明

切换 Provider 后历史列表暂时“消失”，通常是 CC Switch 的 Provider 会话隔离，不代表历史被删除；切回 `default` 后会重新出现。

可以在 CC Switch 设置中开启“统一 Codex 会话历史”。即使历史列表统一，也建议不同 Provider 分别新建任务，因为模型能力、Responses 状态和上下文兼容性并不完全相同。

## 8. 验收清单

### Kiro-Go 通用项

- [ ] 当前代码来自 `Vino1016/Kiro-Go` 的 `main_vino`。
- [ ] 包含提交 `795b2ca`。
- [ ] `curl http://127.0.0.1:8080/health` 返回 `status: ok`。
- [ ] 管理密码已修改。
- [ ] Kiro 账号正常、启用且测试通过。
- [ ] 已开启 Kiro-Go API Key 验证。
- [ ] 未携带 Key 的 `/v1/responses` 返回 401。
- [ ] 携带 Key 的 `/v1/responses` 能返回模型结果。

### Docker 部署项

- [ ] 镜像由当前 `main_vino` 源码构建。
- [ ] `docker ps` 显示 `kiro-go-codex` 正常运行。
- [ ] 端口映射为 `127.0.0.1:8080:8080`。
- [ ] 重启 Docker 后容器能够恢复运行。

### macOS 原生部署项

- [ ] `go test ./...` 通过。
- [ ] `launchctl print` 显示服务正在运行。
- [ ] 关闭终端后服务仍在。
- [ ] 结束进程后 launchd 能自动拉起并恢复健康检查。
- [ ] 日志没有持续崩溃或端口占用错误。

### CC Switch / Codex

- [ ] `default` Provider 保留原 Codex 配置和登录。
- [ ] `Kiro-Go Local` 使用 `http://127.0.0.1:8080/v1`。
- [ ] API 格式是 Responses 原生，Local Routing 关闭。
- [ ] `/model` 能看到带 `-kiro` 的映射模型。
- [ ] Codex 能执行 Shell、文件和 MCP 工具，而不是只能聊天。
- [ ] Kiro-Go 后台请求计数或 credits 增加。
- [ ] 切回 `default` 后官方 Codex 历史和模型恢复。

## 9. 更新、备份与日志

更新前备份配置：

```bash
cp data/config.json "data/config.json.$(date +%Y%m%d-%H%M%S).bak"
chmod 600 data/config.json.*.bak
```

备份包含账号凭证，禁止上传或提交。

Docker 查看日志：

```bash
docker logs -f kiro-go-codex
```

macOS 原生查看日志：

```bash
tail -f logs/kiro-go.log
tail -f logs/kiro-go.error.log
```

需要临时开启调试日志时，优先在管理后台设置；排障结束后恢复 `info`。调试日志可能包含请求上下文，不应长期保留或外发。

## 10. 常见问题

### 10.1 `/model` 没有新模型

- 保存 CC Switch 模型映射；
- 完全退出并重启 Codex；
- 检查 `~/.codex/cc-switch-model-catalog.json` 是否存在；
- 确认当前启用的是 `Kiro-Go Local`。

### 10.2 模型说它没有 Shell、文件或 MCP 工具

在 Kiro-Go 仓库执行：

```bash
git branch --show-current
git merge-base --is-ancestor 795b2ca HEAD && echo OK
```

这通常意味着使用了未修复 Issue #151 的 Kiro-Go，或者 CC Switch 走了错误的 Chat 转换链路。本方案必须使用 `main_vino` 和 Responses 原生直连。

### 10.3 企业账号报 `upstream truncated response without stop reason`

确认使用 `main_vino` 且包含 `795b2ca`。这是 Issue #147 对应的问题。

### 10.4 返回 401

- CC Switch 中填写的必须是 Kiro-Go 管理后台生成的本地 API Key；
- 检查 Key 是否启用；
- 检查 Kiro-Go 是否已开启 API Key 验证；
- 不要把 Kiro 账号的 `ksk_...` Key 与 Kiro-Go 网关 Key 混用。

### 10.5 8080 端口被占用

```bash
lsof -nP -iTCP:8080 -sTCP:LISTEN
docker ps --filter name=kiro-go
```

Docker 和 macOS 原生版本不能同时监听 8080。保留一种运行方式，并停止另一种。

### 10.6 Kiro Provider 中选择“官方同名模型”，是否走 OpenAI

不会。当前 Provider 是 `Kiro-Go Local` 时，所有模型请求都发往 Kiro-Go。要使用 OpenAI 官方模型和额度，必须在 CC Switch 切回 `default`。

### 10.7 切换后 MCP、插件或 Hooks 丢失

确认 `Kiro-Go Local` 已应用 CC Switch 通用配置，并与安装前备份比较 `~/.codex/config.toml`。只恢复缺失配置，避免覆盖当前有效的 Provider 和认证设置。

## 11. 风险提示

- Codex Responses 或 Kiro 上游协议变化后可能需要再次适配。
- CC Switch 会改写用户级 Codex 配置，升级和切换后应检查 MCP、插件与 Hooks。
- kiro-GO只能在本机运行
- Kiro-Go 应限制为 `127.0.0.1:8080`，不要暴露到局域网或公网。

## 参考资料

- [架构与实现原理](./codex-kiro-go-architecture.md)
- [Vino1016/Kiro-Go `main_vino`](https://github.com/Vino1016/Kiro-Go/tree/main_vino)
- [修复提交 `795b2ca`](https://github.com/Vino1016/Kiro-Go/commit/795b2caf95d81da2bf350d8667859e3edf77708c)
- [Kiro-Go Issue #151](https://github.com/Quorinex/Kiro-Go/issues/151)
- [Kiro-Go Issue #147](https://github.com/Quorinex/Kiro-Go/issues/147)
- [CC Switch](https://github.com/farion1231/cc-switch)
- [CC Switch 用户手册](https://github.com/farion1231/cc-switch/tree/main/docs/user-manual)
- [OpenAI Codex 配置参考](https://learn.chatgpt.com/docs/config-file/config-reference)
