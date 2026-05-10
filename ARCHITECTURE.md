# stzbHelper 重构后架构说明

## 1. 分层目标
本轮重构的核心目标是：

- 把“抓包、解析、业务、数据访问、HTTP 入口”解耦
- 降低单文件复杂度（去巨石文件）
- 把 SQL 风险点（字符串拼接）改为参数化
- 为后续依赖注入（DI）和测试打基础

---

## 2. 目录结构（当前）

```text
main.go
internal/
  app/                 # 应用启动编排
  server/httpserver/   # HTTP 服务启动
  capture/             # 抓包与包级处理
  parser/              # 协议解析与数据落库
  runtime/             # 运行时状态集中管理（开关/IP绑定）
  repo/                # 数据访问层（统一 DB 入口）
  service/             # 业务层（组合 repo 能力）
http/
  route/               # 路由注册
  handle/              # Handler（尽量薄，只做参数/响应）
model/                 # 数据模型与数据库初始化
web/                   # 前端资源
```

---

## 3. 请求与数据流

### 3.1 HTTP 控制流

```text
Router -> Handler -> Service -> Repo -> DB
```

说明：
- Handler：参数解析 + 返回 JSON
- Service：业务编排、条件组装
- Repo：仅负责数据读写

### 3.2 抓包链路

```text
Capture -> Parser -> Model/Repo(DB)
```

说明：
- capture 只处理网络包边界、重组、过滤
- parser 只处理协议语义与对象转换
- DB 初始化后，通过 `repo.SetDB` 注入给 repo 层

---

## 4. 关键改造点

1. **入口瘦身**：`main.go` 仅保留 `app.Run()`
2. **巨石拆分**：`handle.go` 已拆为 `team/task/report/battle`
3. **参数化 SQL**：`GetPlayerTeam` 已改为参数化查询
4. **状态收口**：`internal/runtime/state.go` 管理报告开关与 IP 绑定
5. **DB 统一入口**：`internal/repo/db.go` 提供 `SetDB/DB`，支持注入

---

## 5. 后续建议

- 将剩余 `model.Conn` 直接访问继续迁移到 repo 层
- 在 app 层显式装配 repo/service 依赖，逐步去全局变量
- 为 service 层补充最小单测（优先报表/查询类函数）
