# stzbHelper

> 当前版本：`v0.0.4-refactor.33`（详见 `VERSION` 与 `docs/CHANGELOG.md`）

率土之滨攻城考勤助手。通过抓包获取战报与同盟成员数据，提供任务考勤、武勋统计、队伍查询等能力。

---

## 功能概览

- 攻城任务考勤（主力/拆迁队伍数量与攻城次数）
- 分组周武勋统计
- 队伍查询（基于战报回放）
- 额外 JSON 数据导入（skill/hero/gear）

---

## 重构后架构（重点）

本项目已从单体巨石文件改为分层结构：

```text
main.go -> internal/app
             ├─ internal/server/httpserver   (HTTP 启动)
             ├─ internal/capture             (抓包)
             ├─ internal/parser              (协议解析)
             ├─ internal/service             (业务编排)
             ├─ internal/repo                (数据访问)
             └─ internal/runtime             (运行时状态)

http/router -> http/handle(api) -> internal/service -> internal/repo -> DB
```

详细说明见：
- `docs/ARCHITECTURE.md`
- `docs/TECH_BASELINE.md`

---

## 运行依赖

- Go 1.24+
- Node.js（前端构建）
- 抓包能力（Windows 常用 Npcap；macOS 需 root + /dev/bpf* 权限）

---

## 构建方式

### 1) 安装依赖
```bash
go mod tidy
```

```bash
cd web
npm install
npm run build
cd ..
```

### 2) 编译

Windows（示例）：
```bash
go build -tags="nomsgpack" -ldflags="-s -w" -o dist/stzbHelper-windows-amd64.exe .
```

macOS arm64（示例）：
```bash
CGO_ENABLED=1 go build -tags="nomsgpack" -ldflags="-s -w" -o dist/stzbHelper-darwin-arm64 .
```

---

## 启动

### 方式 A：直接运行编译产物（推荐）
```bash
cd /Users/xiasiyi/GolandProjects/stzbHelper
./dist/stzbHelper-darwin-arm64
```

如需抓包权限（macOS 常见）：
```bash
cd /Users/xiasiyi/GolandProjects/stzbHelper
sudo ./dist/stzbHelper-darwin-arm64
```

### 方式 B：开发模式启动后端
```bash
cd /Users/xiasiyi/GolandProjects/stzbHelper
CGO_ENABLED=1 go run .
```

启动后访问：
- `http://127.0.0.1:9527`

> 若前端独立调试：
```bash
cd /Users/xiasiyi/GolandProjects/stzbHelper/web
npm install
npm run dev
```

---

## 开发说明

- 协议解析入口：`internal/parser/parse.go`
- 抓包入口：`internal/capture/capture.go`
- 路由入口：`http/route/api/route.go`
- 运行时状态：`internal/runtime/state.go`
- 数据访问入口：`internal/repo/db.go`（支持注入）

当前规则：
- 新增业务优先走 `handler -> service -> repo`
- 避免在 handler 中直接写 SQL
- 查询必须参数化，避免拼接用户输入
