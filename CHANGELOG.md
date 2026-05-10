# Changelog

> 规则：每次代码或配置有实质变更时，同步更新本文件与版本号。

## v0.0.4-refactor.33 (2026-03-19)

### Changed
- 修正“个人排行榜”展示口径，避免误导：
  - Tab 文案由“玩家外观快照(514)”统一为“个人榜原始事件(514)”。
  - 总览卡片文案改为“个人榜原始事件量(514)”与“原始事件最近抓取”。
  - 个人页新增显式提示：当前为原始事件流，未与游戏最终榜单字段一一对齐。
- 个人事件表格字段改为原始字段直出：
  - `原始参数A(param_a)`
  - `原始参数B(param_b)`
  - `原始扩展(extra_raw)`
  不再将其包装为“指标值/截止时间”等可能错误语义。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.33`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.32 (2026-03-19)

### Added
- CMD语义确认中心新增“仅看已忽略”筛选开关。
- `GET /v1/stzb/runtime/cmd/analysis` 新增查询参数 `only_ignored=1`，用于仅查看当前 ignore 列表中的 cmd_id。

### Changed
- `cmd/analysis` 响应新增回显字段 `only_ignored`。
- 现在支持三种视角：
  - 全量（默认）
  - 仅看未忽略（`only_not_ignored=1`）
  - 仅看已忽略（`only_ignored=1`）

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.32`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.31 (2026-03-19)

### Added
- CMD语义确认中心新增“仅看未忽略”筛选开关。
- `GET /v1/stzb/runtime/cmd/analysis` 新增查询参数 `only_not_ignored=1`，用于排除当前 ignore 列表中的 cmd_id。

### Changed
- `cmd/analysis` 响应新增回显字段 `only_not_ignored`，便于前后端一致性校验。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.31`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.30 (2026-03-19)

### Added
- CMD语义确认中心支持按行开关忽略 `cmd_id`（影响抓包统计与日志打印）。
- 新增忽略列表接口：
  - `GET /v1/stzb/runtime/cmd/ignore/list`
  - `POST /v1/stzb/runtime/cmd/ignore/set`（`{cmd_id, ignore}`）
- `GET /v1/stzb/runtime/cmd/analysis` 返回新增字段：
  - `ignored`：该 cmd_id 当前是否被忽略
  - `api_examples`：包含 method/route/req/resp 的 API 示例

### Changed
- `capture.shouldFilterCmd` 从硬编码 switch 改为可配置 ignore map（默认保留原已忽略项：`2100/90008/694/2200/90006`）。
- cmd语义推测输出增强：在页面中直接展示 API 示例，便于结合包样本+日志做功能研判。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.30`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.29 (2026-03-19)

### Added
- CMD语义确认页新增 `cmd_id` 精确筛选参数（支持“只填cmd_id”直接查询）。
- `GET /v1/stzb/runtime/cmd/analysis` 新增 `cmd_id` 查询参数。
- cmd分析结果新增 `apis` 字段，展示 cmd_id 关联的 API 请求路径。

### Changed
- 对照官方 `ParseData(cmdId,data)` 语义：明确旧链路核心仍是 `103(成员)` / `92(战报)`；新协议 `142/143/949` 标注为“新链路开发中”，避免误导为已完全替代旧链路。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.29`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.28 (2026-03-19)

### Fixed
- 清理历史错误兜底遗留影响：
  - `getTeamUser` / `getGroupWu` 查询统一增加 `id > 0` 条件，屏蔽此前错误写入的负ID占位数据。
  - 避免页面继续展示“十几条伪成员/伪武勋”导致的数据错觉。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.28`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.27 (2026-03-19)

### Fixed
- 回退错误兜底，恢复“原版语义优先”策略（避免假数据污染页面）：
  - 移除 `getTeamUser` 的组长视图兜底（不再用 `cmd142` 强行伪造成员列表）。
  - 移除 `getGroupWu` 的分组元数据兜底（不再将 `group_code` 映射为武勋）。
  - 停止 `cmd949 -> team_user` 的临时回填分支，避免字段语义错位写库。
- 保留 `cmd142` 元数据落库能力，仅用于后续独立新接口，不再污染旧接口语义。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.27`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.26 (2026-03-19)

### Fixed
- 推进协议漂移兼容（同盟成员/分组武勋）：
  - `parser.ParseData` 新增 `cmd_id=142` 与 `cmd_id=949` 解析入口。
  - 抓包分发扩展：
    - `data_type=5` 下 `cmd_id=142/949` 进入 `ParseData`
    - `data_type=3` 下 `cmd_id=100/143` 进入 `ParseData`
- 新增 `cmd142` 分组元数据持久化：
  - 新表 `union_group_meta`（group_id/group_name/group_code/leader_name/member_count/power/capture_time）
  - 支持 upsert，AutoMigrate 已接入。
- API 兜底恢复可用性：
  - `getGroupWu` 在 `team_user` 为空时，回退读取 `union_group_meta` 返回分组数据。
  - `getTeamUser` 在 `team_user` 为空时，回退输出组长视图（按分组元数据）。
- 新增 `cmd949` 到 `team_user` 的轻量回填：提取昵称与分组编码并尝试写入成员表。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.26`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.25 (2026-03-19)

### Fixed
- 修复“页面一直空数据”的启动依赖问题：
  - 以前必须先抓到 `cmd_id=3686` 才初始化 DB，导致重启后短时间内 `database_selected=false`，API 查询只能返回空。
  - 现在在抓包 `Start()` 启动时直接执行 `ensureDBReady()`，提前初始化 MySQL 连接并注入 repo，避免等待 3686。
- 日志补充：`[capture] 已初始化默认数据库连接（mysql-only）`。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.25`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.24 (2026-03-19)

### Fixed
- 修复“同盟成员仍为空”的协议漂移问题：
  - 在 `ParseData` 增加 `cmd_id=100` 兜底解析入口。
  - 新增 `parseTeamUserFromCmd100`：递归扫描 cmd100 JSON，提取候选成员二维数组并入库 `team_user`。
  - 新增诊断日志：`[teamuser-cmd100] 保存成功 count=...` / 未找到候选数组 / 无有效成员。
- 保持既有链路：`cmd_id=103` 继续作为主链路；`cmd_id=100` 作为新协议兼容兜底。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.24`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.23 (2026-03-19)

### Fixed
- 修复同盟成员仍无数据的兼容问题：
  - `parseTeamUser` 改为先解析为通用 JSON，再从多种形态中提取成员二维数组（支持根数组与对象包裹 `data/list/members/items`）。
  - 新增解析结果日志：`[teamuser] 保存成功 count=...`，便于现场确认是否入库。
- 修复 data_type=5 分发空解码保护：
  - 仅当 `DecodeType5` 非空时，才对 `cmd_id=103/92` 调用 `ParseData`，避免空数据干扰。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.23`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.22 (2026-03-19)

### Fixed
- 修复分组武勋查询在 MySQL `only_full_group_by` 下报错（1055）：
  - `internal/repo/team_repo.go` 中将 `IFNULL(sub.zero_wu_count,0)` 改为 `MAX(IFNULL(sub.zero_wu_count,0))`，避免非聚合列冲突。
- 修复同盟成员写库稳定性问题：
  - `internal/parser/parse.go`：`parseTeamUser` 增加 JSON 解析错误日志、空数组保护、无效记录过滤，避免异常输入导致“清空数据”。
  - `model/teamuser.go`：`ToTeamUser` 改为安全类型转换，避免类型断言 panic。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.22`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.21 (2026-03-19)

### Fixed
- 修复 data_type=5 的同盟成员/战报解析链路：
  - `internal/capture/capture.go` 中，当 `cmd_id=103/92` 且 `data_type=5` 时，新增 `ParseData` 分发。
  - 解决“抓包有信号但同盟成员、分组武勋页面无数据”的核心链路问题。
- 兼容协议形态切换：业务包从 `data_type=3` 切到 `data_type=5` 时，仍可正常入库。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.21`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.20 (2026-03-19)

### Changed
- 抓包噪声过滤增强：以下高频低业务价值协议不再进入抓包映射日志与 `cmd_id` 统计：
  - `90008`（心跳/确认包）
  - `694`（时间戳同步）
  - `2200`（名称/状态广播）
  - `90006`（高频短包）
  - 保留既有过滤：`2100`（同盟发言）
- 日志截断可观测性增强：
  - `[response-json]` 新增 `json_chars` 与 `truncated` 字段，明确是否被截断及原始长度。
  - `cmd_id=514/6314/6317/700` 的 JSON 预览上限从 1200 提升到 5000，便于字段对照。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.20`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.19 (2026-03-19)

### Changed
- 排行榜看板页面交互优化：新增「总览」Tab，避免依赖浏览器回退。
- 个人排行榜（514）展示改为更可读中文语义：
  - 列名调整为：`榜单ID / 对象ID / 指标值 / 截止时间 / 附加信息 / 抓取时间`
  - `param_a` 转为 `指标值`（千分位）
  - `param_b` 若为 10 位时间戳则转为可读时间，否则原样展示
  - 补充附加信息拼接展示（参数A/参数B/扩展）

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.19`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.18 (2026-03-19)

### Fixed
- 修复同盟排行榜接口 SQL 语法错误：
  - `GET /v1/stzb/leaderboard/union` 查询排序由 `rank asc` 改为 `` `rank` asc ``，避免 MySQL 8 下 `rank` 关键字冲突导致 500。
- 影响修复：排行榜看板默认同盟页签可正常返回数据。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.18`。

### Build
- 已执行 `go test`、后端构建、前端构建。

---

## v0.0.4-refactor.17 (2026-03-19)

### Added
- 新增个人排行榜落库表：`personal_leaderboard`（来源 `cmd_id=514`）
  - 字段：`event_id/object_id/param_raw/param_a/param_b/extra_raw/flag/source_cmd/capture_time`
- 新增后端接口：`GET /v1/stzb/leaderboard/personal`
  - 支持 `limit`、`event_id`、`object_id` 过滤
- 前端排行榜看板升级为双页签：
  - 同盟排行榜（700）
  - 个人排行榜（514）

### Changed
- `internal/capture/capture.go`：`cmd_id=514` 解码后自动入库 `personal_leaderboard`。
- `model/database.go`：AutoMigrate 增加 `PersonalLeaderboard`。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.17`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.16 (2026-03-19)

### Added
- 后端新增同盟排行榜接口：`GET /v1/stzb/leaderboard/union`
  - 支持参数：`limit`（默认50，最大500）、`name`（同盟名模糊过滤）
- 前端新增页面：`/leaderboard`（排行榜看板）
  - 展示字段：排名、同盟、势力值、成员、城池、区域、刷新时间
  - 首页新增“排行榜看板”入口卡片

### Changed
- `cmd_id=700` 解析结果在抓包时自动落库到新表 `union_leaderboard`，用于看板直接读取。
- `model/database.go` AutoMigrate 新增 `UnionLeaderboard`。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.16`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.15 (2026-03-19)

### Changed
- `model/database.go`：持久化主库从 SQLite 切换为本地 MySQL（GORM MySQL 驱动）。
  - 默认 DSN：`root:123456@tcp(127.0.0.1:3306)/stzb?charset=utf8mb4&parseTime=True&loc=Local`
  - 可通过环境变量 `STZB_MYSQL_DSN` 覆盖。
- 所有 `InitDB(...)` 调用统一接入 MySQL，后续持久化数据不再写入 sqlite 文件。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.15`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.14 (2026-03-19)

### Added
- `model/cmd_schema.go`：新增协议解析字段入库能力，自动沉淀“已映射协议”的结构化字段字典。
  - 表 `cmd_schema`：记录 `cmd_id/cmd_name/data_type/sample_text/seen_count/first_seen/updated_at`
  - 表 `cmd_field_schema`：记录 `cmd_id + field_path` 级别的字段类型、样例值、命中次数
- `model/database.go`：AutoMigrate 新增 `CmdSchema`、`CmdFieldSchema`。

### Changed
- `internal/capture/capture.go`：在成功解码响应后，对已映射协议自动调用 `SaveCmdSchema` 入库（异步）。
- `internal/capture/capture.go`：`cmdNameCN` 补充 `514 -> 个人排行榜`。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.14`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.13 (2026-03-19)

### Changed
- `internal/capture/capture.go`：协议中文映射补充
  - `700 -> 同盟排行榜`
  - `3758 -> 同盟通告`
- `internal/capture/capture.go`：`[response-kv]` 增强
  - 新增 `cmd_id=700` 的排行榜精简中文摘要（排名/联盟名/势力/成员/城池/union_id）
  - 新增 `cmd_id=3758` 的通告中文摘要（通告ID/发布者/内容/时间）

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.13`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.12 (2026-03-19)

### Changed
- `internal/capture/capture.go`：按用户要求过滤 `cmd_id=2100`（同盟发言）。
  - 该协议不再输出 `signal/response` 日志
  - 不再进入 `cmd_observed` 统计
- `cmdNameCN` 补充映射：`2100 -> 同盟发言`（用于代码可读性和后续扩展）。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.12`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.11 (2026-03-19)

### Added
- `internal/capture/capture.go`：新增 `[response-json]` 日志（格式化 JSON，截断输出）用于直接查看响应中文内容结构。
- `internal/capture/capture.go`：新增 `[response-kv]` 日志（结构化键值输出），首批支持 `cmd_id=514` 的事件对解析。

### Changed
- 响应解析链路拆分为 `decodeResponse`（解码）+ `responsePreview/responseJSONPreview/responseKVPreview`（多视图输出）。
- 保留原有 `[response-preview]`，同时补充更可读的 JSON/KV 视图，便于你直接做中文映射判断。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.11`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.10 (2026-03-19)

### Changed
- `internal/capture/capture.go`：新增帧重同步（resync）扫描，解决 PSH 边界错位导致的假 `cmd_id` 问题。
  - 新增 `pickFrame/frameMeta`：在 `offset 0..128` 范围内寻找合法帧头。
  - 仅当 `data_type in {3,5}` 且长度关系合法时认定为有效帧。
- 抓包日志分级：
  - `[signal-valid]`：offset=0 的合法帧
  - `[signal-resync]`：偏移修正后找到的合法帧
  - `[signal-noise]`：未找到合法帧
- 新增响应预览日志 `[response-preview]`：对有效帧尝试解压/解码后输出截断预览（便于快速判断排行榜返回内容）。
- `cmd_observed` 统计改为仅记录有效/重同步后的 `cmd_id`，减少噪声协议污染。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.10`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.9 (2026-03-19)

### Changed
- `internal/capture/capture.go`：`[signal]` 日志新增 `cmd_name` 中文字段。
  - 已映射：
    - `3686` -> `主公簿激活包`
    - `103` -> `同盟成员数据`
    - `92` -> `战报数据`
  - 未映射协议统一输出：`未映射`
- 目的：在高频日志场景下快速识别已知协议，减少人工筛选成本。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.9`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.8 (2026-03-19)

### Changed
- `internal/capture/capture.go`：新增更早期的抓包命中日志 ` [packet-hit] `，在协议解析与过滤分支前打印，字段包括：
  - `payload_len`
  - `src`
  - `dst`
  - `psh`
- 用途：用于快速区分“完全没抓到包”与“抓到包但未进入 `[signal]` 协议日志分支”。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.8`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.7 (2026-03-19)

### Changed
- `internal/capture/capture.go`：抓包链路新增全信号日志输出。每次识别到 `cmd_id` 后均打印一条统一日志：
  - `cmd_id`
  - `data_type`（`buf[12]`）
  - `payload_len`
  - `packet_buf_len`
  - `src/dst`
  - `PSH`
- 日志前缀：`[signal]`，便于你在终端/日志系统中直接过滤并人工映射功能。

### Version
- `VERSION` / `global/variable.go` / `README.md` / `web/package.json` 升级到 `v0.0.4-refactor.7`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.6 (2026-03-19)

### Added
- `internal/capture/capture.go`：新增全量协议号观测统计（`CmdStats`），按出现次数输出 `cmd_id/count`。
- `GET /v1/stzb/runtime/cmd/list`：新增全量协议号列表接口，便于排查“未知协议/未接入协议”。

### Changed
- `http/handle/api/runtime.go`：`/v1/stzb/runtime/status` 增加
  - `cmd_observed_types`（观测到的协议号种类）
  - `cmd_observed_top`（协议号 Top20）
- `docs/API_DATA_MAP.md`：补充全量协议号观测接口与字段说明。
- `VERSION` / `global/variable.go` / `README.md`：版本提升至 `v0.0.4-refactor.6`。

### Build
- 待执行后端测试与前端构建验证。

---

## v0.0.4-refactor.5 (2026-03-18)

### Added
- `internal/service/extra_dict.go`：新增 `hero_extra.json` 字典加载与缓存（ID -> 武将中文名）。
- `docs/DICT_EXTRA_TABLES.md`：新增扩展表字段字典（英文字段/中文释义/类型/索引）。

### Changed
- `http/handle/api/battle.go`：
  - `report/list`、`report/list/cn`、`player/team/get` 增加中文请求日志。
  - `report/list/cn` 增强：返回进攻/防守武将中文名、结果文本等中文语义字段。
- `docs/API_DATA_MAP.md`：补充中文战报接口的武将名映射说明。
- `VERSION` / `global/variable.go` / `web/package.json` / `README.md`：版本提升至 `v0.0.4-refactor.5`。

### Build
- 后端测试通过；前端构建待本轮完成后继续回归。

---

## v0.0.4-refactor.4 (2026-03-18)

### Added
- `GET /v1/stzb/report/list/cn`：新增中文字段战报接口，前端可直接展示。

### Changed
- `http/handle/api/battle.go`：新增中文映射 `reportToCN`，战报返回支持中文语义字段。
- `http/route/api/route.go`：注册中文战报路由。
- `docs/API_DATA_MAP.md`：补充中文战报接口说明与参数映射。
- `VERSION` / `global/variable.go` / `web/package.json` / `README.md`：版本提升至 `v0.0.4-refactor.4`。

### Ops
- MySQL 8.0（Docker 容器 `stzb`）已初始化，root 密码 `123456`。
- 已将 `skill_extra.json` / `hero_extra.json` / `gear_extra.json` 导入 `stzb` 库。

---

## v0.0.4-refactor.3 (2026-03-18)

### Added
- `internal/parser/parse.go`：新增协议计数/入库计数统计器（103/92/teamuser/report/battle）。

### Changed
- `http/handle/api/runtime.go`：`/v1/stzb/runtime/status` 增强，返回抓包、协议命中与入库计数。
- `docs/API_DATA_MAP.md`：补充运行态接口字段说明（字段级排障数据）。
- `VERSION` / `global/variable.go` / `web/package.json` / `README.md`：版本提升至 `v0.0.4-refactor.3`。

### Build
- 后端测试通过；前端构建通过。

---

## v0.0.4-refactor.2 (2026-03-18)

### Added
- `web/src/pages/Team.vue`：将 `teamweb` 队伍分析页并入 `web` 工程。

### Changed
- `web/src/routes.js`：新增 `/team` 路由，统一到单一 Vue 项目入口。
- `web/src/pages/Index.vue`：总览页新增“队伍深度分析(合并)”入口卡片。
- `VERSION` / `global/variable.go` / `web/package.json`：版本统一提升至 `v0.0.4-refactor.2`。

### Build
- `web` 前端打包通过（vite build）。

---

## v0.0.4-refactor.1 (2026-03-18)

### Added
- `internal/repo/db.go`：新增 `SetDB/DB`，支持依赖注入与全局兜底。
- `http/handle/api/runtime.go`：新增运行时状态排查接口 `GET /v1/stzb/runtime/status`。
- `docs/API_DATA_MAP.md`：新增“数据来源 -> API -> 参数 -> 排障 SOP”总览。

### Changed
- `internal/app/app.go`：启动时初始化默认库 `stzb_default`，降低首包前写库失败概率。
- `internal/capture/capture.go`：增加抓包/解析计数、绑定容错与中文日志。
- `http/route/api/route.go`：注册运行态排查路由。
- `web/src/pages/TeamUser.vue`：工具栏与列表卡片样式优化。
- `internal/server/httpserver/http.go`：清理无用 import（time）。

### Notes
- 当前仍在进行 `web + teamweb` 合并收口；未提交 commit 不视为稳定版本。
