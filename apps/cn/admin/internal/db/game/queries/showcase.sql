-- name: LockShowcaseDomain :exec
SELECT pg_advisory_xact_lock(hashtext('gfg.showcase')::bigint);

-- name: CreateShowcaseCampaign :one
INSERT INTO gfg_showcase_campaign(internal_name,content_type,sponsored,linked_game_id)
VALUES (sqlc.arg(internal_name),sqlc.arg(content_type),sqlc.arg(sponsored),sqlc.narg(linked_game_id)) RETURNING *;

-- name: GetShowcaseCampaign :one
SELECT * FROM gfg_showcase_campaign WHERE id=sqlc.arg(id);

-- name: LockShowcaseCampaign :one
SELECT * FROM gfg_showcase_campaign WHERE id=sqlc.arg(id) FOR UPDATE;

-- name: ListShowcaseCampaigns :many
SELECT c.*,count(*) OVER()::bigint AS total FROM gfg_showcase_campaign c
WHERE (sqlc.arg(state)::text='' OR c.state=sqlc.arg(state))
 AND (sqlc.narg(sponsored)::boolean IS NULL OR c.sponsored=sqlc.narg(sponsored))
 AND (sqlc.arg(keyword)::text='' OR c.internal_name ILIKE '%'||sqlc.arg(keyword)||'%' OR c.id::text=sqlc.arg(keyword))
 AND (sqlc.arg(derived_status)::text='' OR CASE WHEN c.state<>'published' THEN c.state
      WHEN sqlc.arg(at_time)::timestamptz<c.starts_at THEN 'scheduled'
      WHEN sqlc.arg(at_time)::timestamptz>=c.ends_at THEN 'ended' ELSE 'active' END=sqlc.arg(derived_status))
ORDER BY c.id DESC LIMIT sqlc.arg(row_limit)::integer OFFSET sqlc.arg(row_offset)::integer;

-- name: SaveShowcaseCampaign :one
UPDATE gfg_showcase_campaign SET internal_name=sqlc.arg(internal_name), content_type=sqlc.arg(content_type),
 sponsored=sqlc.arg(sponsored),linked_game_id=sqlc.narg(linked_game_id),state=sqlc.arg(state),
 starts_at=sqlc.narg(starts_at),ends_at=sqlc.narg(ends_at),weight=sqlc.arg(weight),pin_position=sqlc.narg(pin_position),
 desktop_object_key=sqlc.narg(desktop_object_key),mobile_object_key=sqlc.narg(mobile_object_key),
 focal_x=sqlc.arg(focal_x),focal_y=sqlc.arg(focal_y),primary_action_type=sqlc.narg(primary_action_type),
 primary_target=sqlc.narg(primary_target),secondary_action_type=sqlc.narg(secondary_action_type),
 secondary_target=sqlc.narg(secondary_target),updated_at=now()
WHERE id=sqlc.arg(id) RETURNING *;

-- name: ListShowcaseLocales :many
SELECT * FROM gfg_showcase_campaign_locale WHERE campaign_id=sqlc.arg(campaign_id) ORDER BY lang;

-- name: SaveShowcaseLocale :exec
INSERT INTO gfg_showcase_campaign_locale(campaign_id,lang,enabled,title,summary,tags,editorial_note)
VALUES(sqlc.arg(campaign_id),sqlc.arg(lang),sqlc.arg(enabled),sqlc.arg(title),sqlc.arg(summary),sqlc.arg(tags)::text[],sqlc.narg(editorial_note))
ON CONFLICT(campaign_id,lang) DO UPDATE SET enabled=EXCLUDED.enabled,title=EXCLUDED.title,
 summary=EXCLUDED.summary,tags=EXCLUDED.tags,editorial_note=EXCLUDED.editorial_note,updated_at=now();

-- name: DeleteShowcaseDraft :execrows
DELETE FROM gfg_showcase_campaign WHERE id=sqlc.arg(id) AND state='draft';

-- name: ShowcasePinConflict :one
SELECT EXISTS(SELECT 1 FROM gfg_showcase_campaign c
 JOIN gfg_showcase_campaign_locale l ON l.campaign_id=c.id AND l.enabled
 WHERE c.id<>sqlc.arg(id) AND c.state='published' AND c.pin_position IS NOT NULL
 AND c.starts_at<sqlc.arg(ends_at)::timestamptz AND c.ends_at>sqlc.arg(starts_at)::timestamptz
 AND l.lang=ANY(sqlc.arg(langs)::text[])
 AND (c.pin_position=sqlc.arg(pin_position)::smallint OR (c.sponsored AND sqlc.arg(sponsored)::boolean)
      OR c.linked_game_id=sqlc.narg(linked_game_id)::bigint))::boolean;

-- name: ShowcaseGameStatus :one
SELECT g.id,g.showcase_eligible,NOT EXISTS(SELECT 1 FROM gfg_game_tag m JOIN gfg_tag t ON t.id=m.tag_id
 WHERE m.game_id=g.id AND t.code='adult') AS sfw FROM gfg_game g WHERE g.id=sqlc.arg(id);

-- name: SetShowcaseEligibility :exec
UPDATE gfg_game SET showcase_eligible=sqlc.arg(showcase_eligible) WHERE id=sqlc.arg(id);

-- name: ListShowcaseCampaignStats :many
SELECT * FROM gfg_showcase_daily_stat WHERE campaign_id=sqlc.arg(campaign_id)
 AND stat_date BETWEEN sqlc.arg(from_date)::date AND sqlc.arg(to_date)::date ORDER BY stat_date;

-- name: ListShowcaseQuality :many
SELECT * FROM gfg_showcase_analytics_daily_quality
WHERE stat_date BETWEEN sqlc.arg(from_date)::date AND sqlc.arg(to_date)::date ORDER BY stat_date;

-- name: CountShowcaseCampaigns :one
SELECT count(*)::bigint FROM gfg_showcase_campaign c
WHERE (sqlc.arg(state)::text='' OR c.state=sqlc.arg(state))
 AND (sqlc.narg(sponsored)::boolean IS NULL OR c.sponsored=sqlc.narg(sponsored))
 AND (sqlc.arg(keyword)::text='' OR c.internal_name ILIKE '%'||sqlc.arg(keyword)||'%' OR c.id::text=sqlc.arg(keyword))
 AND (sqlc.arg(derived_status)::text='' OR CASE WHEN c.state<>'published' THEN c.state
      WHEN sqlc.arg(at_time)::timestamptz<c.starts_at THEN 'scheduled'
      WHEN sqlc.arg(at_time)::timestamptz>=c.ends_at THEN 'ended' ELSE 'active' END=sqlc.arg(derived_status))
;
