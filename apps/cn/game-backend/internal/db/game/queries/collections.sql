-- name: CountPublishedCollections :one
WITH visible_members AS (
    SELECT i.collection_id, g.name, g.name_en,
        (fa.game_id IS NOT NULL OR r.availability = 'available') AS released,
        (fa.game_id IS NULL AND r.availability = 'upcoming') AS upcoming
    FROM gfg_game_collection_item i
    JOIN gfg_game g ON g.id = i.game_id
    LEFT JOIN gfg_game_first_available fa ON fa.game_id = g.id
    LEFT JOIN gfg_game_release_state r ON r.game_id = g.id
    WHERE sqlc.arg(include_adult)::boolean OR NOT EXISTS (
        SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id = gt.tag_id
        WHERE gt.game_id = g.id AND t.code = 'adult'
    )
), membership AS (
    SELECT collection_id, count(*) AS visible_count,
        bool_or(released) AS has_released, bool_or(upcoming) AS has_upcoming,
        bool_or(strpos(lower(name), lower(sqlc.arg(keyword)::text)) > 0
            OR strpos(lower(name_en), lower(sqlc.arg(keyword)::text)) > 0) AS member_match
    FROM visible_members GROUP BY collection_id
), filtered AS (
    SELECT c.*, COALESCE(m.visible_count, 0)::bigint AS visible_count
    FROM gfg_game_collection c LEFT JOIN membership m ON m.collection_id = c.id
    WHERE c.status = 'published'
      AND (sqlc.arg(keyword)::text = '' OR m.member_match
        OR strpos(lower(c.code), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.name), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.name_en), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.info), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.info_en), lower(sqlc.arg(keyword)::text)) > 0)
      AND (sqlc.arg(phase)::text = 'all'
        OR (sqlc.arg(phase)::text = 'released' AND m.has_released)
        OR (sqlc.arg(phase)::text = 'upcoming' AND m.has_upcoming)
        OR (sqlc.arg(phase)::text = 'mixed' AND m.has_released AND m.has_upcoming))
)
SELECT count(*) FROM filtered;

-- name: ListPublishedCollections :many
WITH visible_members AS (
    SELECT i.collection_id, g.name, g.name_en,
        (fa.game_id IS NOT NULL OR r.availability = 'available') AS released,
        (fa.game_id IS NULL AND r.availability = 'upcoming') AS upcoming
    FROM gfg_game_collection_item i
    JOIN gfg_game g ON g.id = i.game_id
    LEFT JOIN gfg_game_first_available fa ON fa.game_id = g.id
    LEFT JOIN gfg_game_release_state r ON r.game_id = g.id
    WHERE sqlc.arg(include_adult)::boolean OR NOT EXISTS (
        SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id = gt.tag_id
        WHERE gt.game_id = g.id AND t.code = 'adult'
    )
), membership AS (
    SELECT collection_id, count(*) AS visible_count,
        bool_or(released) AS has_released, bool_or(upcoming) AS has_upcoming,
        bool_or(strpos(lower(name), lower(sqlc.arg(keyword)::text)) > 0
            OR strpos(lower(name_en), lower(sqlc.arg(keyword)::text)) > 0) AS member_match
    FROM visible_members GROUP BY collection_id
), filtered AS (
    SELECT c.*, COALESCE(m.visible_count, 0)::bigint AS visible_count
    FROM gfg_game_collection c LEFT JOIN membership m ON m.collection_id = c.id
    WHERE c.status = 'published'
      AND (sqlc.arg(keyword)::text = '' OR m.member_match
        OR strpos(lower(c.code), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.name), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.name_en), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.info), lower(sqlc.arg(keyword)::text)) > 0
        OR strpos(lower(c.info_en), lower(sqlc.arg(keyword)::text)) > 0)
      AND (sqlc.arg(phase)::text = 'all'
        OR (sqlc.arg(phase)::text = 'released' AND m.has_released)
        OR (sqlc.arg(phase)::text = 'upcoming' AND m.has_upcoming)
        OR (sqlc.arg(phase)::text = 'mixed' AND m.has_released AND m.has_upcoming))
)
SELECT id, code, name, name_en, info, info_en, published_at
FROM filtered
ORDER BY
    CASE WHEN sqlc.arg(sort)::text = 'count_desc' THEN visible_count END DESC,
    CASE WHEN sqlc.arg(sort)::text = 'count_asc' THEN visible_count END ASC,
    CASE WHEN sqlc.arg(sort)::text = 'name_asc' THEN
        CASE WHEN sqlc.arg(lang)::text = 'en' THEN CASE WHEN btrim(name_en) = '' THEN name ELSE name_en END
        ELSE CASE WHEN btrim(name) = '' THEN name_en ELSE name END END END ASC,
    CASE WHEN sqlc.arg(sort)::text = 'name_desc' THEN
        CASE WHEN sqlc.arg(lang)::text = 'en' THEN CASE WHEN btrim(name_en) = '' THEN name ELSE name_en END
        ELSE CASE WHEN btrim(name) = '' THEN name_en ELSE name END END END DESC,
    CASE WHEN sqlc.arg(sort)::text IN ('name_asc', 'name_desc') THEN id END ASC,
    published_at DESC, id DESC
LIMIT sqlc.arg(page_size)::bigint OFFSET sqlc.arg(page_offset)::bigint;

-- name: GetPublishedCollection :one
SELECT id, code, name, name_en, info, info_en, published_at
FROM gfg_game_collection WHERE code = sqlc.arg(code) AND status = 'published';

-- name: ListPublishedCollectionHomeSlots :many
SELECT s.slot, c.id, c.code, c.name, c.name_en, c.info, c.info_en, c.published_at,
    (SELECT count(*) FROM gfg_game_collection_item i
     WHERE i.collection_id = c.id
       AND (sqlc.arg(include_adult)::boolean OR NOT EXISTS (
           SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id = gt.tag_id
           WHERE gt.game_id = i.game_id AND t.code = 'adult'
       ))) AS visible_game_count
FROM gfg_game_collection_home_slot s
JOIN gfg_game_collection c ON c.id = s.collection_id
WHERE c.status = 'published'
ORDER BY s.slot;

-- name: BatchCollectionMemberships :many
SELECT collection_id, game_id FROM gfg_game_collection_item
WHERE collection_id = ANY(sqlc.arg(collection_ids)::bigint[])
ORDER BY collection_id, game_id;

-- name: BatchCollectionProjectionGames :many
SELECT g.id, g.name, g.name_en, g.info, g.info_en, g.header,
    COALESCE(d.name, '')::text AS detail_name, d.header_url AS detail_header,
    zh.game_id AS zh_id, zh.name AS zh_name, zh.short_description AS zh_summary,
    en.game_id AS en_id, en.name AS en_name, en.short_description AS en_summary,
    EXISTS (SELECT 1 FROM gfg_game_tag gt JOIN gfg_tag t ON t.id = gt.tag_id
        WHERE gt.game_id = g.id AND t.code = 'adult') AS adult
FROM gfg_game g
LEFT JOIN gfg_game_details d ON d.game_id = g.id
LEFT JOIN gfg_game_localized_details zh ON zh.game_id = g.id AND zh.lang = 'zh'
LEFT JOIN gfg_game_localized_details en ON en.game_id = g.id AND en.lang = 'en'
WHERE g.id = ANY(sqlc.arg(game_ids)::bigint[])
ORDER BY g.id;

-- name: BatchCollectionHeaderMedia :many
SELECT game_id, url FROM gfg_game_media
WHERE game_id = ANY(sqlc.arg(game_ids)::bigint[]) AND media_type = 'header'
ORDER BY game_id, media_type, sort_order, id;

-- name: BatchCollectionHeaderAssets :many
SELECT game_id, asset_type, lang, url, exists FROM gfg_game_assets
WHERE game_id = ANY(sqlc.arg(game_ids)::bigint[]) AND asset_type IN ('header', 'header_2x')
ORDER BY game_id, asset_family, sort_order, id;

-- name: CountCollectionBrowse :one
SELECT count(*) FROM gfg_game_collection WHERE status = 'published';

-- name: ListCollectionBrowse :many
SELECT id, code, name, name_en, info, info_en, published_at
FROM gfg_game_collection WHERE status = 'published'
ORDER BY published_at DESC, id DESC
LIMIT sqlc.arg(page_size)::bigint OFFSET sqlc.arg(page_offset)::bigint;
