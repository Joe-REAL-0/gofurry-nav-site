package dao

import (
	"context"
	"errors"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
)

func (dao *ReadModelDAO) CountPublishedCollections(ctx context.Context, query v2models.CollectionQuery) (int64, error) {
	if err := dao.ready(); err != nil {
		return 0, err
	}
	if query.IsDefaultBrowse() {
		return dao.q.CountCollectionBrowse(ctx)
	}
	return dao.q.CountPublishedCollections(ctx, gamesqlc.CountPublishedCollectionsParams{IncludeAdult: query.Mode == "nsfw", Keyword: query.Q, Phase: query.Phase})
}

func (dao *ReadModelDAO) ListPublishedCollections(ctx context.Context, query v2models.CollectionQuery) ([]v2models.CollectionRecord, error) {
	if err := dao.ready(); err != nil {
		return nil, err
	}
	if query.IsDefaultBrowse() {
		rows, err := dao.q.ListCollectionBrowse(ctx, gamesqlc.ListCollectionBrowseParams{PageSize: query.PageSize, PageOffset: (query.Page - 1) * query.PageSize})
		if err != nil {
			return nil, err
		}
		result := make([]v2models.CollectionRecord, 0, len(rows))
		for _, row := range rows {
			result = append(result, v2models.CollectionRecord{ID: row.ID, Code: row.Code, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, PublishedAt: row.PublishedAt.Time})
		}
		return result, nil
	}
	rows, err := dao.q.ListPublishedCollections(ctx, gamesqlc.ListPublishedCollectionsParams{PageSize: query.PageSize, PageOffset: (query.Page - 1) * query.PageSize, IncludeAdult: query.Mode == "nsfw", Keyword: query.Q, Phase: query.Phase, Sort: query.Sort, Lang: query.Lang})
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

func (dao *ReadModelDAO) ListPublishedCollectionHomeSlots(ctx context.Context, mode string) ([]v2models.CollectionRecord, error) {
	if err := dao.ready(); err != nil {
		return nil, err
	}
	rows, err := dao.q.ListPublishedCollectionHomeSlots(ctx, mode == "nsfw")
	if err != nil {
		return nil, err
	}
	result := make([]v2models.CollectionRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, v2models.CollectionRecord{ID: row.ID, Code: row.Code, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, PublishedAt: row.PublishedAt.Time, Slot: row.Slot, VisibleGameCount: row.VisibleGameCount})
	}
	return result, nil
}

func (dao *ReadModelDAO) LoadCollectionProjectionGames(ctx context.Context, collectionIDs []int64, lang string) (v2models.CollectionGames, error) {
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
	gameIDs = uniqueInt64s(gameIDs)
	if len(gameIDs) == 0 {
		return result, nil
	}
	games, err := dao.q.BatchCollectionProjectionGames(ctx, gameIDs)
	if err != nil {
		return result, err
	}
	result.Games = make([]v2models.CollectionProjectionGame, 0, len(games))
	byID := make(map[int64]*v2models.CollectionProjectionGame, len(games))
	for _, row := range games {
		var zh, en *v2models.GfgGameV2LocalizedDetails
		if row.ZhID != nil {
			zh = &v2models.GfgGameV2LocalizedDetails{GameID: row.ID, Lang: "zh", Name: strPtrValue(row.ZhName), ShortDescription: row.ZhSummary}
		}
		if row.EnID != nil {
			en = &v2models.GfgGameV2LocalizedDetails{GameID: row.ID, Lang: "en", Name: strPtrValue(row.EnName), ShortDescription: row.EnSummary}
		}
		localized := mergeLocalizedDetails(zh, en)
		if lang == "en" {
			localized = mergeLocalizedDetails(en, zh)
		}
		result.Games = append(result.Games, v2models.CollectionProjectionGame{
			Site:      v2models.GameV2SiteRecord{ID: row.ID, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, Header: row.Header},
			Details:   &v2models.GfgGameV2Details{Name: row.DetailName, HeaderURL: row.DetailHeader},
			Localized: localized, Adult: row.Adult,
		})
		byID[row.ID] = &result.Games[len(result.Games)-1]
	}
	// Fixed batches, independent of collection/member count. Header reads retain
	// the aggregate loader's ordering, but omit every other media/asset type.
	media, err := dao.q.BatchCollectionHeaderMedia(ctx, gameIDs)
	if err != nil {
		return result, err
	}
	for _, row := range media {
		if g := byID[row.GameID]; g != nil {
			g.Media = append(g.Media, v2models.GfgGameV2Media{MediaType: "header", URL: &row.Url})
		}
	}
	assets, err := dao.q.BatchCollectionHeaderAssets(ctx, gameIDs)
	if err != nil {
		return result, err
	}
	for _, row := range assets {
		if g := byID[row.GameID]; g != nil {
			g.Assets = append(g.Assets, v2models.GfgGameV2Asset{AssetType: row.AssetType, Lang: row.Lang, URL: row.Url, Exists: row.Exists})
		}
	}
	releases, err := dao.q.BatchReleaseStatesByGames(ctx, gameIDs)
	if err != nil {
		return result, err
	}
	for _, row := range releases {
		if g := byID[row.GameID]; g != nil {
			value := releaseStateView(row)
			g.ReleaseState = &value
		}
	}
	facts, err := dao.q.BatchFirstAvailableByGames(ctx, gameIDs)
	if err != nil {
		return result, err
	}
	for _, row := range facts {
		if g := byID[row.GameID]; g != nil {
			value := firstAvailableView(row)
			g.FirstAvailable = &value
		}
	}
	return result, nil
}
