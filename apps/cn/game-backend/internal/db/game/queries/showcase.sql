-- name: ListShowcaseManaged :many
SELECT sqlc.embed(c),sqlc.embed(l),
 (c.linked_game_id IS NULL OR EXISTS(SELECT 1 FROM gfg_game g WHERE g.id=c.linked_game_id))::boolean AS game_exists,
 NOT EXISTS(SELECT 1 FROM gfg_game_tag m JOIN gfg_tag t ON t.id=m.tag_id
   WHERE m.game_id=c.linked_game_id AND t.code='adult')::boolean AS sfw
FROM gfg_showcase_campaign c JOIN gfg_showcase_campaign_locale l ON l.campaign_id=c.id
WHERE c.state='published' AND l.enabled AND l.lang=sqlc.arg(lang)
 AND c.starts_at<=sqlc.arg(at_time)::timestamptz AND c.ends_at>sqlc.arg(at_time)::timestamptz ORDER BY c.id;

-- name: NextShowcaseBoundary :one
SELECT min(boundary)::timestamptz FROM (
 SELECT c.starts_at AS boundary FROM gfg_showcase_campaign c JOIN gfg_showcase_campaign_locale l ON l.campaign_id=c.id
 WHERE c.state='published' AND l.enabled AND l.lang=sqlc.arg(lang) AND c.starts_at>sqlc.arg(at_time)::timestamptz
 UNION ALL
 SELECT c.ends_at FROM gfg_showcase_campaign c JOIN gfg_showcase_campaign_locale l ON l.campaign_id=c.id
 WHERE c.state='published' AND l.enabled AND l.lang=sqlc.arg(lang) AND c.ends_at>sqlc.arg(at_time)::timestamptz
) boundaries;

-- name: ListShowcaseGames :many
SELECT g.id,g.appid,g.showcase_eligible,
 CASE WHEN sqlc.arg(lang)::text='en' THEN g.name_en ELSE g.name END::text AS title,
 CASE WHEN sqlc.arg(lang)::text='en' THEN g.info_en ELSE g.info END::text AS summary,
 COALESCE((SELECT a.url FROM gfg_game_assets a WHERE a.game_id=g.id AND a.appid=g.appid
   AND a.asset_type IN ('header','header_2x') AND btrim(a.url)<>'' AND a.exists IS DISTINCT FROM false
   ORDER BY CASE a.asset_type WHEN 'header' THEN 0 ELSE 1 END,
     CASE WHEN a.lang=sqlc.arg(lang)::text THEN 0 WHEN a.lang<>'' THEN 1 ELSE 2 END,a.sort_order,a.id LIMIT 1),
   (SELECT NULLIF(m.url,'') FROM gfg_game_media m WHERE m.game_id=g.id AND m.appid=g.appid AND m.media_type='header'
     ORDER BY m.sort_order DESC,m.id DESC LIMIT 1),NULLIF(d.header_url,''),NULLIF(g.header,''),'')::text AS artwork_url,
 NOT EXISTS(SELECT 1 FROM gfg_game_tag m JOIN gfg_tag t ON t.id=m.tag_id
   WHERE m.game_id=g.id AND t.code='adult')::boolean AS sfw,
 ARRAY(SELECT CASE WHEN sqlc.arg(lang)::text='en' THEN t.name_en ELSE t.name END::text
   FROM gfg_game_tag m JOIN gfg_tag t ON t.id=m.tag_id WHERE m.game_id=g.id
   ORDER BY CASE m.role WHEN 'primary' THEN 0 WHEN 'secondary' THEN 1 ELSE 2 END,t.id LIMIT 3)::text[] AS tags,
 r.availability,r.precision,r.exact_date,r.release_year,r.release_month,r.release_quarter,r.window_start,r.window_end,
 COALESCE(f.exact_date,f.window_start)::date AS first_available,
 p.id AS tracking_period_id
FROM gfg_game g
LEFT JOIN gfg_game_details d ON d.game_id=g.id AND d.appid=g.appid
LEFT JOIN gfg_game_release_state r ON r.game_id=g.id
LEFT JOIN gfg_game_first_available f ON f.game_id=g.id
LEFT JOIN gfg_game_tracking_periods p ON p.game_id=g.id AND p.appid=g.appid AND p.tracking_basis='explicit' AND p.tracked_until IS NULL
ORDER BY g.id;

-- name: ListShowcasePlayerFacts :many
SELECT d.tracking_period_id,d.game_id,d.fact_date,d.avg_players,d.successful_samples,d.expected_samples,d.quality_basis,
 h.processed_through
FROM gfg_fact_rollup_checkpoints h
JOIN gfg_game_tracking_periods p ON p.tracking_basis='explicit' AND p.tracked_until IS NULL
JOIN gfg_game g ON g.id=p.game_id AND g.appid=p.appid
JOIN gfg_game_player_daily d ON d.tracking_period_id=p.id AND d.game_id=g.id AND d.appid=g.appid
WHERE h.pipeline_key='game.player_facts' AND d.finalized_at IS NOT NULL
 AND d.fact_date BETWEEN h.processed_through-9 AND h.processed_through
 AND d.fact_date >= (p.tracked_from AT TIME ZONE 'UTC')::date
ORDER BY d.tracking_period_id,d.fact_date;

-- name: UpsertShowcaseDailyStat :exec
INSERT INTO gfg_showcase_daily_stat(stat_date,subject_key,subject_kind,campaign_id,game_id,reason,
 valid_impressions,qualified_clicks,click_artwork,click_title,click_primary,click_secondary,
 impression_position_1,impression_position_2,impression_position_3,impression_position_4,session_estimate)
VALUES(sqlc.arg(stat_date),sqlc.arg(subject_key),sqlc.arg(subject_kind),sqlc.narg(campaign_id),sqlc.narg(game_id),sqlc.arg(reason),
 sqlc.arg(valid_impressions),sqlc.arg(qualified_clicks),sqlc.arg(click_artwork),sqlc.arg(click_title),sqlc.arg(click_primary),sqlc.arg(click_secondary),
 sqlc.arg(impression_position_1),sqlc.arg(impression_position_2),sqlc.arg(impression_position_3),sqlc.arg(impression_position_4),sqlc.arg(session_estimate))
ON CONFLICT(stat_date,subject_key) DO UPDATE SET
 valid_impressions=EXCLUDED.valid_impressions,qualified_clicks=EXCLUDED.qualified_clicks,
 click_artwork=EXCLUDED.click_artwork,click_title=EXCLUDED.click_title,click_primary=EXCLUDED.click_primary,click_secondary=EXCLUDED.click_secondary,
 impression_position_1=EXCLUDED.impression_position_1,impression_position_2=EXCLUDED.impression_position_2,
 impression_position_3=EXCLUDED.impression_position_3,impression_position_4=EXCLUDED.impression_position_4,
 session_estimate=EXCLUDED.session_estimate,updated_at=now()
WHERE gfg_showcase_daily_stat.valid_impressions<=EXCLUDED.valid_impressions
 AND gfg_showcase_daily_stat.qualified_clicks<=EXCLUDED.qualified_clicks;

-- name: UpsertShowcaseQuality :exec
INSERT INTO gfg_showcase_analytics_daily_quality(stat_date,invalid_token_events,invalid_origin_events,filtered_user_agent_events,
 duplicate_impressions,duplicate_clicks,session_rate_limited,ip_rate_limited,malformed_events)
VALUES(sqlc.arg(stat_date),sqlc.arg(invalid_token_events),sqlc.arg(invalid_origin_events),sqlc.arg(filtered_user_agent_events),
 sqlc.arg(duplicate_impressions),sqlc.arg(duplicate_clicks),sqlc.arg(session_rate_limited),sqlc.arg(ip_rate_limited),sqlc.arg(malformed_events))
ON CONFLICT(stat_date) DO UPDATE SET invalid_token_events=GREATEST(gfg_showcase_analytics_daily_quality.invalid_token_events,EXCLUDED.invalid_token_events),
 invalid_origin_events=GREATEST(gfg_showcase_analytics_daily_quality.invalid_origin_events,EXCLUDED.invalid_origin_events),
 filtered_user_agent_events=GREATEST(gfg_showcase_analytics_daily_quality.filtered_user_agent_events,EXCLUDED.filtered_user_agent_events),
 duplicate_impressions=GREATEST(gfg_showcase_analytics_daily_quality.duplicate_impressions,EXCLUDED.duplicate_impressions),
 duplicate_clicks=GREATEST(gfg_showcase_analytics_daily_quality.duplicate_clicks,EXCLUDED.duplicate_clicks),
 session_rate_limited=GREATEST(gfg_showcase_analytics_daily_quality.session_rate_limited,EXCLUDED.session_rate_limited),
 ip_rate_limited=GREATEST(gfg_showcase_analytics_daily_quality.ip_rate_limited,EXCLUDED.ip_rate_limited),
 malformed_events=GREATEST(gfg_showcase_analytics_daily_quality.malformed_events,EXCLUDED.malformed_events),updated_at=now();
