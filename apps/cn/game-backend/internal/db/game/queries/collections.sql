-- name: CountPublishedCollections :one
SELECT count(*) FROM gfg_game_collection WHERE status = 'published';

-- name: ListPublishedCollections :many
SELECT id, code, name, name_en, info, info_en, published_at
FROM gfg_game_collection WHERE status = 'published'
ORDER BY published_at DESC, id DESC
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
