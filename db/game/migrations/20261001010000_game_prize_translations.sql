-- +goose Up
ALTER TABLE public.gfg_prize
    ADD COLUMN title_en varchar(100) NOT NULL DEFAULT '',
    ADD COLUMN desc_en text NOT NULL DEFAULT '';

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.gfg_prize WHERE title_en <> '' OR desc_en <> '') THEN
        RAISE EXCEPTION 'prize translations exist; refusing lossy rollback';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.gfg_prize DROP COLUMN title_en, DROP COLUMN desc_en;
