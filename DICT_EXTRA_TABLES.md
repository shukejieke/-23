# DICT_EXTRA_TABLES

三张扩展表字段字典（英文名 / 中文释义 / 类型 / 是否索引）。

- 数据库：`stzb`
- 生成时间：2026-03-18

## skill_extra

| 英文字段 | 中文释义 | 类型 | 索引 |
|---|---|---|---|
| `isAppear` | 是否出现 | `varchar(255)` | 否 |
| `probability` | 概率 | `varchar(255)` | 否 |
| `zfQuality` | 战法品质 | `varchar(255)` | 否 |
| `skillCount` | 技能数量 | `bigint` | 否 |
| `studyDesc2` | 研究描述2 | `varchar(255)` | 否 |
| `id` | 主键ID | `bigint` | 是 |
| `studyDesc` | 研究描述 | `longtext` | 否 |
| `studyStar` | 研究星级 | `longtext` | 否 |
| `type` | 类型 | `varchar(255)` | 是 |
| `targetShow` | 目标展示 | `varchar(255)` | 否 |
| `notice` | 公告/提示 | `varchar(255)` | 否 |
| `targetType` | 目标类型 | `varchar(255)` | 是 |
| `soldierType` | 兵种类型 | `varchar(255)` | 否 |
| `effect` | 效果 | `varchar(255)` | 否 |
| `desc` | 描述 | `longtext` | 否 |
| `distance` | 距离/攻击距离 | `longtext` | 否 |
| `name` | 名称 | `varchar(255)` | 是 |
| `dismantling` | 拆解说明 | `longtext` | 否 |
| `disassembleHero1` | 拆解武将1 | `varchar(255)` | 否 |
| `disassembleHero2` | 拆解武将2 | `varchar(255)` | 否 |
| `disassembleHero3` | 拆解武将3 | `varchar(255)` | 否 |
| `desc(level1)` | 1级描述 | `longtext` | 否 |

## hero_extra

| 英文字段 | 中文释义 | 类型 | 索引 |
|---|---|---|---|
| `methodName` | 技能名 | `varchar(255)` | 否 |
| `share_desc` | 分享描述 | `varchar(255)` | 否 |
| `sex` | 性别 | `varchar(255)` | 否 |
| `methodName1` | 技能名1 | `varchar(255)` | 否 |
| `ruseGrow` | 谋略成长 | `varchar(255)` | 否 |
| `cost` | 统率消耗 | `varchar(255)` | 否 |
| `base_hero_id` | 基础武将ID | `longtext` | 否 |
| `speed` | 速度 | `varchar(255)` | 否 |
| `id` | 主键ID | `bigint` | 是 |
| `policyName` | 政策名 | `varchar(255)` | 否 |
| `policyDesc` | 政策描述 | `varchar(255)` | 否 |
| `group` | 分组 | `varchar(255)` | 是 |
| `methodDesc2` | 技能描述2 | `varchar(255)` | 否 |
| `methodDesc1` | 技能描述1 | `varchar(255)` | 否 |
| `quality` | 品质 | `varchar(255)` | 是 |
| `iconId` | 图标ID | `bigint` | 否 |
| `defGrow` | 防御成长 | `varchar(255)` | 否 |
| `groupName` | 分组名称 | `varchar(255)` | 否 |
| `attack` | 攻击 | `varchar(255)` | 否 |
| `methodId2` | 技能ID2 | `longtext` | 否 |
| `speedGrow` | 字段:speedGrow | `varchar(255)` | 否 |
| `methodDesc` | 技能描述 | `longtext` | 否 |
| `type` | 类型 | `varchar(255)` | 是 |
| `type_availible` | 可用类型 | `varchar(255)` | 否 |
| `policyId` | 政策ID | `longtext` | 否 |
| `methodId` | 技能ID | `bigint` | 否 |
| `siege` | 攻城值 | `varchar(255)` | 否 |
| `uniqueName` | 唯一名 | `varchar(255)` | 是 |
| `methodId1` | 技能ID1 | `bigint` | 否 |
| `groupId` | 分组ID | `varchar(255)` | 否 |
| `desc` | 描述 | `longtext` | 否 |
| `distance` | 距离/攻击距离 | `bigint` | 否 |
| `name` | 名称 | `varchar(255)` | 是 |
| `country` | 国家/阵营 | `varchar(255)` | 是 |
| `ruse` | 谋略 | `varchar(255)` | 否 |
| `methodName2` | 技能名2 | `varchar(255)` | 否 |
| `siegeGrow` | 攻城成长 | `varchar(255)` | 否 |
| `def` | 防御 | `varchar(255)` | 否 |
| `attGrow` | 攻击成长 | `varchar(255)` | 否 |

## gear_extra

| 英文字段 | 中文释义 | 类型 | 索引 |
|---|---|---|---|
| `skillDesc` | 装备技能描述 | `varchar(255)` | 否 |
| `skillName` | 装备技能名 | `varchar(255)` | 否 |
| `woodAdvance` | 木属性进阶 | `bigint` | 否 |
| `name` | 名称 | `varchar(255)` | 是 |
| `policyName` | 政策名 | `varchar(255)` | 否 |
| `type` | 类型 | `varchar(255)` | 是 |
| `obtain` | 获取方式 | `varchar(255)` | 否 |
| `ironAdvance` | 铁属性进阶 | `bigint` | 否 |
| `woodEnchant` | 木属性精炼 | `varchar(255)` | 否 |
| `ironForge` | 铁属性锻造 | `bigint` | 否 |
| `policyDesc` | 政策描述 | `varchar(255)` | 否 |
| `ironEnchant` | 铁属性精炼 | `varchar(255)` | 否 |
| `featureGroup` | 特性分组 | `bigint` | 否 |
| `woodForge` | 木属性锻造 | `bigint` | 否 |
| `quality` | 品质 | `varchar(255)` | 是 |
| `id` | 主键ID | `bigint` | 是 |
| `condition` | 条件 | `varchar(255)` | 否 |
| `desc` | 描述 | `varchar(255)` | 否 |
