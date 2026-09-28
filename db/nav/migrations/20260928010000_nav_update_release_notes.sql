-- +goose Up
ALTER TABLE gfn_nav_update_notice
    ADD COLUMN version varchar(64),
    ADD COLUMN commit_sha varchar(64),
    ADD COLUMN summary text NOT NULL DEFAULT '',
    ADD COLUMN summary_en text NOT NULL DEFAULT '',
    ADD COLUMN publication_state varchar(16);

-- Legacy notices were public. Preserve their content and publication timestamps.
UPDATE gfn_nav_update_notice SET publication_state = 'published';

ALTER TABLE gfn_nav_update_notice
    ALTER COLUMN publication_state SET NOT NULL,
    ALTER COLUMN publication_state SET DEFAULT 'draft',
    ALTER COLUMN published_at DROP NOT NULL,
    ADD CONSTRAINT gfn_nav_update_notice_publication_state_check
        CHECK (publication_state IN ('draft', 'published')),
    ADD CONSTRAINT gfn_nav_update_notice_version_check
        CHECK (version IS NULL OR (version = btrim(version) AND version <> '')),
    ADD CONSTRAINT gfn_nav_update_notice_commit_sha_check
        CHECK (commit_sha IS NULL OR commit_sha ~ '^[0-9a-f]{7,64}$');

-- +goose Down
-- Refuse a lossy rollback: draft timestamps must be resolved by an operator.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM gfn_nav_update_notice WHERE published_at IS NULL) THEN
        RAISE EXCEPTION 'Cannot roll back Release Notes while published_at is NULL';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE gfn_nav_update_notice
    ALTER COLUMN published_at SET NOT NULL,
    DROP COLUMN publication_state,
    DROP COLUMN summary_en,
    DROP COLUMN summary,
    DROP COLUMN commit_sha,
    DROP COLUMN version;
