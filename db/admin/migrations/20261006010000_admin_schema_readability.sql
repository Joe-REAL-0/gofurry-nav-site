-- +goose Up
-- 当前业务语义的中文元数据；仅修改 COMMENT，不修改结构、数据或函数行为。
-- 后续纠错通过新迁移前滚，不提供恢复缺失或旧语义注释的 Down。

COMMENT ON TABLE public."gfa_admin_account" IS 'Admin 本地账号身份、密码凭证与会话撤销版本；权限映射由编译代码持有';
COMMENT ON COLUMN public."gfa_admin_account"."id" IS 'Admin 本地账号身份、密码凭证与会话撤销版本的行身份；用于稳定引用该记录';
COMMENT ON COLUMN public."gfa_admin_account"."password_hash" IS '登录密码的单向哈希凭证，不保存明文密码';
COMMENT ON COLUMN public."gfa_admin_account"."session_version" IS '该账号的会话撤销版本；变更后旧版本登录令牌失效';
COMMENT ON COLUMN public."gfa_admin_account"."created_at" IS 'Admin 本地账号身份、密码凭证与会话撤销版本的记录创建时间';
COMMENT ON COLUMN public."gfa_admin_account"."updated_at" IS 'Admin 本地账号身份、密码凭证与会话撤销版本的记录最近更新时间';
COMMENT ON COLUMN public."gfa_admin_account"."password_updated_at" IS '最近设置或重置密码的时间；历史账号可能为 NULL';
COMMENT ON COLUMN public."gfa_admin_account"."username" IS '规范化小写登录名，去除首尾空白且在账号域内唯一';
COMMENT ON COLUMN public."gfa_admin_account"."display_name" IS '后台显示名称；审计记录保存操作发生时的名称快照';
COMMENT ON COLUMN public."gfa_admin_account"."role" IS '固定角色 owner、developer 或 operator；能力集合由后端策略推导';
COMMENT ON COLUMN public."gfa_admin_account"."status" IS '账号可用状态：active 可登录，disabled 禁止登录';
COMMENT ON COLUMN public."gfa_admin_account"."last_login_at" IS '最后一次成功登录时间；从未登录时为 NULL';

COMMENT ON TABLE public."gfa_admin_audit_log" IS 'Admin 操作的持久审计记录，保存操作者身份快照及变更前后内容';
COMMENT ON COLUMN public."gfa_admin_audit_log"."id" IS 'Admin 操作的持久审计记录的行身份；用于稳定引用该记录';
COMMENT ON COLUMN public."gfa_admin_audit_log"."action" IS '发生的业务操作代码，用于检索和解释审计事件';
COMMENT ON COLUMN public."gfa_admin_audit_log"."resource" IS '被操作的资源类型代码；不是数据库连接名';
COMMENT ON COLUMN public."gfa_admin_audit_log"."target_id" IS '被操作资源的文本身份；无单一目标的操作可为 NULL';
COMMENT ON COLUMN public."gfa_admin_audit_log"."operator" IS '历史兼容的操作者文本；新身份另由账号、名称和角色快照保存';
COMMENT ON COLUMN public."gfa_admin_audit_log"."session_version" IS '操作发生时的会话版本；系统或旧记录可为零';
COMMENT ON COLUMN public."gfa_admin_audit_log"."request_id" IS '关联本次 HTTP 请求的追踪标识；未提供时为 NULL';
COMMENT ON COLUMN public."gfa_admin_audit_log"."ip_address" IS '审计请求来源地址；未取得时为 NULL，不用于重建角色权限';
COMMENT ON COLUMN public."gfa_admin_audit_log"."user_agent" IS '审计请求的 User-Agent；未取得时为 NULL';
COMMENT ON COLUMN public."gfa_admin_audit_log"."before_data" IS '操作前快照的序列化文本；敏感字段由调用方按审计策略脱敏，无快照时为空串或历史 NULL';
COMMENT ON COLUMN public."gfa_admin_audit_log"."after_data" IS '操作后快照的序列化文本；敏感字段由调用方按审计策略脱敏，无快照时为空串或历史 NULL';
COMMENT ON COLUMN public."gfa_admin_audit_log"."created_at" IS 'Admin 操作的持久审计记录的记录创建时间';
COMMENT ON COLUMN public."gfa_admin_audit_log"."operator_account_id" IS '实际操作者账号；历史或系统操作可为 NULL，账号删除受外键限制';
COMMENT ON COLUMN public."gfa_admin_audit_log"."operator_name" IS '操作发生时的显示名称快照，不随账号改名回写';
COMMENT ON COLUMN public."gfa_admin_audit_log"."operator_role" IS '操作发生时的角色快照；系统操作使用 system';

COMMENT ON TABLE public."gfa_collaboration_board_edge" IS '共享画布节点间的连线；节点删除时级联删除关联连线';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."id" IS '共享画布节点间的连线的行身份；用于稳定引用该记录';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."source_id" IS '起点节点身份，引用同一共享画布的节点';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."target_id" IS '终点节点身份；不得与起点相同';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."source_handle" IS '起点连接方位 top、right、bottom 或 left';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."target_handle" IS '终点连接方位 top、right、bottom 或 left';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."routing" IS '连线路径形态：curve 曲线或 step 折线';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."label" IS '连线上的说明文本，不超过 200 字符';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."color" IS '画布语义色：sand、blue、green、rose 或 slate，不存任意 CSS';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."arrow" IS '是否显示连线方向箭头';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."created_by_account_id" IS '创建该协作对象的 Admin 账号身份，引用 gfa_admin_account';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."updated_by_account_id" IS '最近修改该协作对象的 Admin 账号身份，引用 gfa_admin_account';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."version" IS '乐观并发版本；修改时必须匹配旧版本并递增';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."created_at" IS '共享画布节点间的连线的记录创建时间';
COMMENT ON COLUMN public."gfa_collaboration_board_edge"."updated_at" IS '共享画布节点间的连线的记录最近更新时间';

COMMENT ON TABLE public."gfa_collaboration_board_node" IS '协作中心唯一共享画布的节点；位置和尺寸更新也受版本检查';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."id" IS '协作中心唯一共享画布的节点的行身份；用于稳定引用该记录';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."body" IS '节点正文文本；不作为正式 Game/Nav 内容发布';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."x" IS '节点在共享画布坐标系中的横向位置';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."y" IS '节点在共享画布坐标系中的纵向位置';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."width" IS '节点在画布中的宽度，范围 48..1600';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."height" IS '节点在画布中的高度，范围 40..1600';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."z_index" IS '节点叠放顺序，数值越大越靠前';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."created_by_account_id" IS '创建该协作对象的 Admin 账号身份，引用 gfa_admin_account';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."updated_by_account_id" IS '最近修改该协作对象的 Admin 账号身份，引用 gfa_admin_account';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."version" IS '乐观并发版本；修改时必须匹配旧版本并递增';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."created_at" IS '协作中心唯一共享画布的节点的记录创建时间';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."updated_at" IS '协作中心唯一共享画布的节点的记录最近更新时间';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."kind" IS '节点形态：note、card、text、rectangle、ellipse 或 arrow';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."title" IS '节点标题文本；图形节点允许无标题';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."color" IS '画布语义色：sand、blue、green、rose 或 slate，不存任意 CSS';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."rotation" IS '节点旋转角度，仅允许 0、90、180、270 度';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."reference_kind" IS '卡片引用类型 idea、game 或 site；无引用时与 reference_id 同为 NULL';
COMMENT ON COLUMN public."gfa_collaboration_board_node"."reference_id" IS '卡片关联对象身份；正式内容为跨库逻辑引用，不复制其业务数据';

COMMENT ON TABLE public."gfa_content_idea" IS '协作中心内容想法池；正式内容仍由 Game/Nav API 创建，落地关联不是跨库事务';
COMMENT ON COLUMN public."gfa_content_idea"."id" IS '协作中心内容想法池的行身份；用于稳定引用该记录';
COMMENT ON COLUMN public."gfa_content_idea"."kind" IS '想法类型：game 游戏、site 网站或 other 其它';
COMMENT ON COLUMN public."gfa_content_idea"."title" IS '运营填写的想法标题；可缺省，但标题与来源不能同时为空';
COMMENT ON COLUMN public."gfa_content_idea"."source" IS '运营记录的原始来源或线索文本；未提供时为 NULL';
COMMENT ON COLUMN public."gfa_content_idea"."source_key" IS '来源归一键，如 steam:AppID 或 host:域名；仅用于重复提示，不保证唯一';
COMMENT ON COLUMN public."gfa_content_idea"."note" IS '调研与协作备注文本，不直接发布到正式内容';
COMMENT ON COLUMN public."gfa_content_idea"."priority" IS '处理优先级 normal 或 high，不代表正式内容排序';
COMMENT ON COLUMN public."gfa_content_idea"."status" IS '流转阶段：idea 待研究、researching 研究中、landed 已落地、shelved 已搁置';
COMMENT ON COLUMN public."gfa_content_idea"."created_by_account_id" IS '创建该协作对象的 Admin 账号身份，引用 gfa_admin_account';
COMMENT ON COLUMN public."gfa_content_idea"."researching_by_account_id" IS '研究中阶段的认领账号；其它阶段为 NULL，不是租约';
COMMENT ON COLUMN public."gfa_content_idea"."linked_kind" IS '落地资源类型 game 或 site；未关联时为 NULL';
COMMENT ON COLUMN public."gfa_content_idea"."linked_resource_id" IS '正式 GFG/GFN 内容的逻辑身份；跨库不建外键，重开想法清除此关联';
COMMENT ON COLUMN public."gfa_content_idea"."version" IS '乐观并发版本；修改时必须匹配旧版本并递增';
COMMENT ON COLUMN public."gfa_content_idea"."created_at" IS '协作中心内容想法池的记录创建时间';
COMMENT ON COLUMN public."gfa_content_idea"."updated_at" IS '协作中心内容想法池的记录最近更新时间';
COMMENT ON COLUMN public."gfa_content_idea"."researching_at" IS '进入当前研究中阶段的时间；其它阶段为 NULL';
COMMENT ON COLUMN public."gfa_content_idea"."landed_at" IS '本次完成落地的时间；非 landed 阶段为 NULL';
