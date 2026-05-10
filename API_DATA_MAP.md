# 数据抓取与 API 参数映射（重构版）

> 目标：把“抓到什么包 -> 入什么表 -> 前端调什么 API”梳理清楚，方便后续扩展和排障。

## 1. 抓包层（TCP 8001）

- 入口：`internal/capture/capture.go`
- BPF 过滤：`tcp and src port 8001`
- 数据类型识别：`buf[12]`
  - `3`：zlib 压缩数据（主链路）
  - `5`：异或数据（`DecodeType5`）

### 关键协议号

- `3686`：主公簿激活包
  - 用途：绑定游戏 IP（src/dst）+ 初始化角色数据库
- `103`：同盟成员数据
  - 入库：`team_user`
- `92`：战报数据
  - 分支：
    - `NeedGetBattleData=false` -> 普通战报（`report`）
    - `NeedGetBattleData=true` -> 详细战报（`battle_report`）

## 2. 运行时状态

- 状态收口：`internal/runtime/state.go`
- 调试接口：`GET /v1/stzb/runtime/status`

返回字段：
- `packet_seen`：捕获包数量
- `parse_triggered`：触发解析数量
- `database_selected`：是否完成主公簿绑定
- `bound_src / bound_dst`：当前绑定 IP
- `cmd103_count`：收到同盟成员协议次数
- `cmd92_count`：收到战报协议次数
- `teamuser_save_count`：累计入库同盟成员条数
- `report_save_count`：累计入库普通战报条数
- `battle_save_count`：累计入库详细战报条数
- `cmd_observed_types`：当前会话观测到的协议号种类数
- `cmd_observed_top`：按出现次数排序的 Top 协议号列表（最多20条）

新增接口：
- `GET /v1/stzb/runtime/cmd/list`
  - 用途：返回抓包阶段观测到的全量协议号统计
  - 返回：
    - `total_types`：协议号种类数
    - `items[]`：`{ cmd_id, count }`（按 count desc, cmd_id asc）

- `GET /v1/stzb/leaderboard/union`
  - 用途：返回同盟排行榜看板数据（来源于 `cmd_id=700` 解析落库）
  - 参数：
    - `limit`（可选，默认 50，最大 500）
    - `name`（可选，同盟名模糊匹配）
  - 返回：
    - `items[]`：`rank/name/power/total_member/total_npc_city/region/refresh_time/union_id`
    - `count`

- `GET /v1/stzb/leaderboard/personal`
  - 用途：返回个人排行榜看板数据（来源于 `cmd_id=514` 解析落库）
  - 参数：
    - `limit`（可选，默认 100，最大 1000）
    - `event_id`（可选，按事件ID过滤）
    - `object_id`（可选，按对象ID过滤）
  - 返回：
    - `items[]`：`event_id/object_id/param_raw/param_a/param_b/extra_raw/flag/capture_time`
    - `count`

## 3. 数据到 API 对照

### 3.1 同盟成员
- 数据来源：协议 `103`
- 表：`team_user`
- API：
  - `GET /v1/getTeamUser`
    - 参数：`group`（可选）
  - `GET /v1/getTeamGroup`
  - `GET /v1/getGroupWu`

### 3.2 任务管理
- 表：`task`
- API：
  - `POST /v1/createTask`
    - 参数：
      - `taskname` string
      - `tasktime` int
      - `targetgroup` []string（repeat）
      - `taskpos` []string（repeat）
  - `GET /v1/getTaskList`
  - `GET /v1/getTask/:tid`
  - `GET /v1/deleteTask/:tid`

### 3.3 普通战报统计
- 数据来源：协议 `92`（普通分支）
- 表：`report`
- API：
  - `POST /v1/enable/getReport`
    - 参数：`pos` int
  - `GET /v1/disable/getReport`
  - `GET /v1/getReportNumByTaskId/:tid`
  - `GET /v1/statisticsReport/:tid`
  - `GET /v1/deleteTaskReport/:tid`

### 3.4 详细战报与队伍
- 数据来源：协议 `92`（详细分支）
- 表：`battle_report`
- API：
  - `GET /v1/enable/getBattleReport`
  - `GET /v1/disable/getBattleReport`
  - `GET /v1/stzb/report/list`
    - 返回：原始英文字段
    - 参数：
      - `nextid` (必填)
      - `atkname` (可选)
      - `atkunionname` (可选)
      - `atkhp` (可选)
      - `atklevel` (可选)
      - `atkstar` (可选)
      - `type` (可选, 1/2/3/4)
      - `nonpc` (可选, 1 表示过滤 NPC)
  - `GET /v1/stzb/report/list/cn`
    - 返回：中文字段（适合前端直出）
    - 额外解析：按 `hero_extra.json` 映射进攻/防守武将中文名
    - 参数：同 `report/list`
  - `GET /v1/stzb/player/team/get`
    - 参数：
      - `atkname` (可选)
      - `atkunionname` (可选)
      - `idu` (可选)

## 4. 排障 SOP（抓不到数据）

1) 确认程序以管理员权限运行（macOS 需要访问 `/dev/bpf*`）
2) 访问 `GET /v1/stzb/runtime/status`
   - `packet_seen=0`：抓包权限/网卡不对
   - `packet_seen>0 且 parse_triggered=0`：过滤条件命中但协议未进入解析
   - `parse_triggered>0 且 database_selected=false`：未触发 `3686` 主公簿绑定
3) 在游戏中打开主公簿，观察 `database_selected` 是否变 `true`
4) 再打开同盟成员页或战报页触发 `103/92` 数据

---

后续会继续补：
- 每个 API 的响应字段样例
- 前端页面到 API 的调用链图
- 定时任务（排行榜/周榜）输入输出字典
