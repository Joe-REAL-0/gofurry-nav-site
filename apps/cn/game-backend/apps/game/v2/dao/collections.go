package dao

import (
	"context"
	"errors"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
)

func (dao *ReadModelDAO) CountPublishedCollections(ctx context.Context) (int64, error) {
	if err := dao.ready(); err != nil {
		return 0, err
	}
	return dao.q.CountPublishedCollections(ctx)
}

func (dao *ReadModelDAO) ListPublishedCollections(ctx context.Context, limit, offset int64) ([]v2models.CollectionRecord, error) {
	if err := dao.ready(); err != nil {
		return nil, err
	}
	rows, err := dao.q.ListPublishedCollections(ctx, gamesqlc.ListPublishedCollectionsParams{PageSize: limit, PageOffset: offset})
	if err != nil {
		return nil, err
	}
	result := make([]v2models.CollectionRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, v2models.CollectionRecord{ID: row.ID, Code: row.Code, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, PublishedAt: row.PublishedAt.Time})
	}
	return result, nil
}

func (dao *ReadModelDAO) GetPublishedCollection(ctx context.Context, code string) (*v2models.CollectionRecord, error) {
	if err := dao.ready(); err != nil {
		return nil, err
	}
	row, err := dao.q.GetPublishedCollection(ctx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v2models.CollectionRecord{ID: row.ID, Code: row.Code, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, PublishedAt: row.PublishedAt.Time}, nil
}

func (dao *ReadModelDAO) ListPublishedCollectionHomeSlots(ctx context.Context) ([]v2models.CollectionRecord, error) {
	if err := dao.ready(); err != nil {
		return nil, err
	}
	rows, err := dao.q.ListPublishedCollectionHomeSlots(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]v2models.CollectionRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, v2models.CollectionRecord{ID: row.ID, Code: row.Code, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, PublishedAt: row.PublishedAt.Time, Slot: row.Slot})
	}
	return result, nil
}

func (dao *ReadModelDAO) LoadCollectionGames(ctx context.Context, collectionIDs []int64, lang string) (v2models.CollectionGames, error) {
	var result v2models.CollectionGames
	if err := dao.ready(); err != nil {
		return result, err
	}
	if len(collectionIDs) == 0 {
		return result, nil
	}
	rows, err := dao.q.BatchCollectionMemberships(ctx, collectionIDs)
	if err != nil {
		return result, err
	}
	gameIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		result.Memberships = append(result.Memberships, v2models.CollectionMembership{CollectionID: row.CollectionID, GameID: row.GameID})
		gameIDs = append(gameIDs, row.GameID)
	}
	// Deduplicate across the whole response before entering the existing V2
	// aggregate loader. newsLimit=0 keeps all reads bounded batches, not per game.
	// Its tag batch deliberately retains archived tags and their adult semantics.
	result.Games, err = dao.loadAggregatesByGameIDs(ctx, uniqueInt64s(gameIDs), lang, 0)
	return result, err
}
