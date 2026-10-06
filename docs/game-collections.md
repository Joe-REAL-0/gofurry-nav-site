# Game Collection：数据域与公开时间线

Stage A（#140-A）建立人工策展游戏分区及 Game Backend 公开读模型。
人工只决定成员，公开顺序由 Canonical First Available / Release Facts 派生。
Collection 不是 Tag、Showcase、Recommendation，也不使用 `gfg_game.groups`。

## Stage C：Public Discovery 与 Timeline

Nav Web 提供 `/games/collections` 和 `/games/collections/:code`（以及 `/en` 对应路由）。
首页通过一次 `Promise.allSettled` 并行读取 Home、Showcase、Collections Home；两个可选 slice 各有 1 秒预算、无重试。
Collections Home 永远请求 SFW，失败时整个快捷入口 owner 缺席；合法空 slots 仍显示“全部分区”。
同一个快捷入口组件在 Desktop Sidebar 使用三列，在小于 xl 的内容区位于 Showcase 与最近发售之间，使用两列。
后端给出的非空 slots 按 slot 升序压缩展示，最后追加固定的全部分区入口。

Index 只提供简介、预览卡片和手动加载更多；0/1/2/3 张 preview 原样消费，不复制图片。
Detail 的单一时间线由后端排序，前端仅稳定分组：已发布、已发布但日期待考、NOW、未来、TBA、未知。
过去和未来各自复用 `serpentineSequence` 几何，900px 以上三列蛇形，以下单列；DOM 顺序始终不变。
日期待考、TBA、未知没有 connector，只有未知作品时不显示 NOW。NOW 取后端 `as_of_date`。
显示保留 day/month/quarter/year 与 inferred 精度，不暴露内部 First Available 来源，也不在前端过滤 Adult Tag。

两个分区页面 SSR 一律 SFW。挂载读取本地模式：SFW 零补请求，NSFW 恰好一次刷新。
页面实例独占 `useGameCollectionModeRefresh`，响应按 generation 接受，模式更新成功才整体替换内容；失败保留上次内容和局部重试。
Index 模式更新成功回到第一页，过期的 Load More 响应不能追加进新模式。模式不写 URL。
Detail 初始/刷新 404 使用真实 Nuxt 404；初始服务失败返回 HTTP 503 与可重试 surface，不伪装空分区。
adult-only 的 SFW Detail 仍为 200，零作品使用中性空态。所有图片复用 SteamAssetImage。

Stage C 不修改 Stage A 缓存、数据库或 Admin。Functional、固定环境 Visual 比对和维护者人工验收分别记录；
新增六张分区截图与首页四张变更须人工接受后才构成 Visual PASS，代码完成不代表 #140 已关闭。

2026-10-06 本地验证：Unit 366、Nuxt 64、focused Browser 76、受影响的既有消费者回归 38 项通过；
lint/stylelint/style:policy/typecheck/build、SEO/Insights guards 与 repository policy 通过。
固定 Playwright image 中六张新分区快照已生成并完成 6/6 复比，仍等待维护者接受。
首页四张旧快照未改，比较出现预期的快捷入口差异；最终 Public 视觉/模式人工验收与 Visual Closure 尚未完成。

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

Stage A 的交付包含 GFG migration 与 Game Backend；Collector 仅新增 generated schema model，不要求单独发布。
Stage A 不包含 Admin 与 Public UX。Stage B 运营合同见下文；Stage C 的 Nav Web 页面、模式与时间线合同及待完成人工验收状态见本文开头。

## Stage B：Admin 运营闭环

Admin 独立拥有 `/api/v1/game/collections`，与 `/api/v1/collection` 的采集控制面无关。
所有 GET 使用现有 `content.read`，所有写操作使用 `content.write`，不新增能力。

| 方法与路径（上述前缀内） | 行为 |
| --- | --- |
| GET / | 按 keyword（code/name/name_en）、status、home_eligible 筛选；page_num 默认 1、page_size 默认 50、上限 200；updated_at DESC、id DESC |
| POST / | 创建 version=1 的草稿；双语名称必填，简介及 0/1 个成员允许暂缺 |
| GET /:id | 完整 Admin DTO：基本内容、状态、版本、生命周期时间、总成员/SFW 数、home_slot |
| PUT /:id | 携带 version 更新双语名称/简介；Code 不可更改 |
| GET /:id/members | 同一只读快照内返回 collection_id、version、完整成员及 Adult 标记 |
| PUT /:id/members | 携带 version 与完整 game_ids 集合；正整数、去重、数值排序、批量检查存在性后替换 |
| POST /:id/publish、unpublish、archive、restore | 携带 version 执行显式生命周期转换 |
| GET /home-curation | 固定五位及 canonical placement revision；空位 collection=null |
| PUT /home-curation | 携带 revision，完整提交第 1–5 位各一次；非空分区不可重复 |

静态 `/home-curation` 先于 `/:id` 注册。Code 使用 1–64 位小写 kebab-case；重复返回 409。
名称最多 160、简介最多 500 个 Unicode 字符。未知写入字段被拒绝，不能变相写入 Code、排序或 NSFW。

### 版本、事务与生命周期

所有写操作先取得 GFG transaction advisory lock `gfg.game-collection-domain`，再锁定 Collection 行、比较 version、执行依赖修改。
内容、成员及生命周期成功后 version+1；真正相同的内容/成员集合不改变 version，也不写 Audit。
旧 version 返回 409：`此游戏分区已被其他操作修改，请重新加载后重试。`；客户端不得自动重试 mutation。

- Publish：draft→published；双语名称/简介非空、总成员至少两个；published_at=now。
- Unpublish：published→draft；保留 published_at，同事务移除首页入口。
- Archive：draft/published→archived；archived_at=now，保留 published_at/成员并移除入口。
- Restore：archived→draft；清空 archived_at，保留 published_at，不恢复发布或首页入口。
- Archived 拒绝内容及成员修改；Published 直接编辑仍须满足双语内容和至少两个成员的不变量。

Adult 唯一按 Tag code=`adult`，包括已归档 Tag 的现有关系；ID 1014 不具有特殊意义。
成人限定分区可发布，但首页资格必须是 published 且 sfw_member_count>0。
Published 成员保存移除最后一个 SFW 游戏时，同一 GFG 事务自动移除 Home Slot；分区仍可维持 Published。

### 首页与 Audit

revision 只对第 1–5 位 Collection ID/null 的 canonical placement 计算 SHA-256；名称、version、count 不影响 revision。
完整替换在领域锁下校验旧 revision 和当前资格，原子删除/插入；显式首页编排不增加 Collection.version。
第六个“全部分区”是固定产品入口，只在 Admin 说明，不进入 API 或数据库。

Collection Audit resource 为 `gfg_game_collection`，action 为 create/update/members_update/publish/unpublish/archive/restore。
快照仅含有界业务字段，成员保存附带 canonical game ID 集合；自动撤下入口体现在同条记录的 home_slot before→null。
首页显式编排使用 `gfg_game_collection_home_slot` / `home_curation_update`，记录五个 placement。
沿用现有 GFG business transaction + 独立 GFA Audit；审计失败阻止 GFG 提交，但这不是跨库 ACID。

### React Workspace 与刷新

路由：`/game/collections`、`/new`、`/:id`、`/home-curation`（后三者均在相同前缀下）。
列表搜索使用 IME-safe helper 并将 keyword/status/page_num 保存在 URL。
单页 Workspace 包含基本内容、收录游戏、公开刷新说明；成员只做添加/移除，不提供顺序编辑或 Timeline Preview。
Game options 与 Home eligibility 搜索复用 RemoteSelect，不消费 IME 确认键。

内容与成员草稿共享 baseVersion；本页保存成功可推进版本并保留另一份草稿，背景刷新不得覆盖脏草稿或提升其版本。
409 保留草稿并要求显式重新加载。草稿通过 useUnsavedChanges 保护导航，生命周期操作在 dirty 时禁用且全部要求确认。
只读用户可查看完整内容；归档后仅 Restore 可写。Audit 入口仅对 audit.read 显示，使用 resource 过滤而不假装支持 target_id。
Home 草稿同样绑定原 revision，背景刷新不会覆盖它。

Stage B 没有新迁移，也不变更 Stage A cache：不 purge Redis、不新增内部失效接口或 Pub/Sub。
成功提示为“已保存；公开页面将在最多约5分钟内刷新。”；自动撤下首页入口时另行说明。

验证包含 Controller/route 单测、`TestAdminGameCollectionThreeDatabase` 的隔离 PG18 并发/约束/Audit 集成测试，
以及 React 列表 IME、共享版本、冲突保留、生命周期、只读、五位完整编排和未保存保护测试。
集成测试接入既有 postgres-integration gate，不对共享开发或生产库执行 Goose。
