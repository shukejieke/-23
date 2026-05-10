# stzbHelper 技术基础（重构后）

## 1) 语言与运行时
- Go 1.24+
- Node.js（用于前端构建）

## 2) 后端技术栈
- HTTP 框架：Gin
- ORM：GORM
- 本地库：SQLite（默认）
- 可选外部库：MySQL（用于 extra JSON 导入）
- 抓包：gopacket + pcap

## 3) 前端技术栈
- Vite
- Vue（当前前端实现）

## 4) 当前工程风格
- 分层：app / server / capture / parser / service / repo
- handler 薄化：只做 I/O，不做重业务
- SQL 安全：优先参数化，禁止字符串拼接入参
- 运行时状态：集中在 runtime 层读写

## 5) 构建与验证

### 后端
```bash
go mod tidy
go test ./...
```

### 前端
```bash
cd web
npm install
npm run build
```

## 6) 兼容策略
- repo.DB() 支持“注入连接优先，历史全局连接兜底”
- 在不一次性推翻旧代码的情况下，允许渐进式重构
