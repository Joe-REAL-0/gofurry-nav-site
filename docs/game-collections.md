# Game Collection：数据域与公开时间线

Stage A（#140-A）建立人工策展游戏分区及 Game Backend 公开读模型。
人工只决定成员，公开顺序由 Canonical First Available / Release Facts 派生。
Collection 不是 Tag、Showcase、Recommendation，也不使用 `gfg_game.groups`。

## Schema ownership

Goose migration `db/game/migrations/20261006020000_game_collection_foundation.sql`
独占以下 GFG 表；同一迁移提供所有 Table/Column 的中文 COMMENT：

| 表 | 事实与约束 |
| --- | --- |
| `gfg_game_collection` | 内部 identity 主键；唯一 1..64 字符 kebab-case `code`；zh/en 名称与简介；`status`、乐观锁 `version`、发布/归档/维护时间 |
| `gfg_game_collection_item` | `(collection_id, game_id)` 唯一人工成员关系；双向 FK 级联清理关系，反向查询索引 `(game_id, collection_id)` |
| `gfg_game_collection_home_slot` | 1..5 的唯一槽位和唯一 Collection 引用；删除分区时级联清理槽位 |

Item 不含 position、weight、sort_order、note、nsfw。Collection 不含 NSFW、排序或素材字段。
首页第六个“全部分区”入口不入库。本迁移不创建任何 Collection 或 Membership seed。
Down 明确拒绝删除策展数据；恢复需经过验证的备份或重建隔离数据库。

生命周期冻结为 draft → published → draft、draft/published → archived、archived → draft。
`published_at` 是最近一次成功发布时刻，draft 可保留；只有 archived 必须有 `archived_at`。
`version` 从 1 开始。Stage B 才实现版本校验、code 创建后不可变及这些写入转换；
Stage A 没有 Admin mutation、Membership editor、Home curation 或 Audit 功能。

## Public API

三个 GET 均使用现有 `{code: 1, data: ...}` 成功响应 envelope，与 `/api/v2/game/home` 独立：

| 路径 | data |
| --- | --- |
| `/api/v2/game/collections/home` | 公共 metadata、`slots: [{slot, collection: summary}]`；slot 升序，最多五项 |
| `/api/v2/game/collections` | 公共 metadata、`page`、`page_size`、`total`、`has_more`、`items: summary[]` |
| `/api/v2/game/collections/:code` | 公共 metadata、`collection: info`、`items: timelineItem[]`；完整一维时间线，不分页 |

公共 metadata 为 `schema_version: 1`、UTC `generated_at`、UTC 日历日 `as_of_date`。
`lang=zh|en`，缺失或非法默认 zh；`mode=sfw|nsfw`，缺失或非法默认 sfw。
List 默认 `page=1,page_size=24`，page_size 上限 60；非整数、非正数回安全默认值。
无法安全计算 bigint offset 的 page 回到 1。越界页为 `items=[]` 并保留 total。
List 按 `published_at DESC,id DESC` 排序，total 是已发布分区数量。

仅 published 公开。非法 code、draft、archived 和不存在的 Detail 均为 HTTP 404，
同一 `Collection not found` 错误；SQL/连接故障为 503 `Collections unavailable`，不透传内部错误。
空 Index、空 Home 和零可见成员 Detail 都是 200；数组使用 `[]`。
HTTP `Cache-Control: no-store`，避免在服务端五分钟缓存之外叠加浏览器缓存。

`info` 投影：`code,name,info,visible_game_count,published_at`。
`summary` 在 info 上增加 `preview_games: [{game_id,name,header_url}]`。
`timelineItem`：`game_id,name,summary,header_url,phase,chronology`，game_id 是十进制字符串。
不返回 Collection 内部 ID、status、version、archived_at、原始成员数或隐藏成人数。
分区名称/简介逐字段优先请求语言，空值回退另一语言；游戏文案和图片复用现有 V2 优先级。

## 成人过滤与 Preview

唯一成人事实是 `gfg_game_tag → gfg_tag.code='adult'`。
即使标签已归档，只要关系仍存在就保持成人语义；数字 Tag ID 1014 和名称不参与判断。
sfw 在 Backend 先过滤成人成员，再解析时间线、计数和选 Preview；nsfw 返回完整有效成员。
成人专属分区在 sfw Index/Detail 仍存在，`visible_game_count=0,preview_games/items=[]`。
Home 对当前 mode 下零可见成员的槽位直接省略，也省略非 published 分区。

Preview 从过滤后的完整有序时间线选取：0 项为 []，1 项取首项，2 项取首尾，
3 项及以上取 `[0, floor(n/2), n-1]`。最多三个不同游戏，不维护 Banner 或手工排序。

## Chronology

Resolver 接收显式 UTC as-of 日期；不解析 release raw text、不在纯函数内读取实时时钟。
First Available 是最高优先级，即使 current release 变回 upcoming 也保持 released。

| 顺序 / phase | 条件 | chronology |
| --- | --- | --- |
| 1 `released` | 存在 First Available；否则 current available 且有可信 window | `source=first_available` 或 `release` |
| 2 `released_unknown` | current available，无可信 window | null |
| 3 `upcoming_overdue` | current upcoming 且 window_end **严格早于** UTC as_of_date | release window |
| 4 `upcoming` | current upcoming，window_end 当天或未来 | release window |
| 5 `upcoming_tba` | current upcoming，无可信 window | null |
| 6 `unknown` | current unknown/missing 且无 First Available | null |

有 window 的 phase 按 `window_start ASC,window_end ASC,game_id ASC`；其余按数值 game_id ASC。
`window_end == as_of_date` 不算 overdue。
chronology 只含 `source,precision,window_start,window_end,inferred`；保留 day/month/quarter/year 精度。
window 是排序区间，不是合成的精确发布日期；没有 synthetic display date。
First Available 的 inferred 原样保留，内部 legacy_manual/steam_backfill/observed_transition 不公开。

## Batch 与 Cache

固定 sqlc 查询负责公开 Collection、count、page、slots 和批量 Membership。
每个响应收集 unique game IDs 后调用现有 `loadAggregatesByGameIDs` / `loadAggregatesBySites`
批量加载 canonical release、first available、tags、localized details 和 media。
newsLimit 固定为 0，不逐游戏加载新闻；不调用内部 `/game/info` HTTP。
游戏显示投影复用 `buildListItem`，不复制 header resolver。

独立 Redis namespace，TTL **5 minutes**：

```text
game:v2:collections:v1:home:{asOfDate}:{lang}:{mode}
game:v2:collections:v1:list:{asOfDate}:{lang}:{mode}:{page}:{pageSize}
game:v2:collections:v1:detail:{asOfDate}:{lang}:{mode}:{code}
```

UTC 日期切换立即换 key。有效缓存命中直接返回；miss/error/malformed 回 DB，成功后 best-effort 写缓存。
数据库错误和 404 不缓存。每个 Service 实例用 singleflight 合并同 key 的并发 miss；不同 key 独立。
缓存序列化结果按调用方解码，避免共享可变 slices。单个调用方取消不终止其他等待者，构建有 8 秒总预算。
Redis 读写使用 200ms context 和共享连接池的 read/write timeout clone，不改全局 Redis 配置。
Stage A 不做跨服务精确失效或 stale fallback；后续运营修改最多约五分钟最终一致。

## 验证与交付边界

- Service 单测覆盖完整 chronology、UTC 边界、稳定排序、preview、mode/locale、cache TTL/分键/故障及 singleflight。
- Controller 单测覆盖 query normalization、DTO envelope、状态码和错误脱敏。
- `TestPostgresReadModelSemantics/collections` 接入既有 postgres-integration gate，验证 populated read model、实际 HTTP、成人归档语义、固定批次数及 FK/check/unique/cascade。
- `TestPostgresFreshAndBaselineAdoption` 验证 PG18 fresh/adoption/drift/readability；`expected-final/gfg.json` 从实际库生成。
- `task check:db-readability`、`task check:sqlc`、`task check:policy` 保持静态治理。

本阶段未实现 Admin Collection API/UI、Nav Web Collection UI、SEO、Timeline Vue、SSR mode hydration、Visual Golden 或 Games Home 5+1。
生产审核后需执行待应用的 GFG migration 并部署 Game Backend；同窗口更新 Admin 可同步 DataOps migration inventory。
Collector 仅新增 generated schema model 不要求单独发布。本阶段验证仅使用 disposable PostgreSQL，不迁移共享开发或生产。
