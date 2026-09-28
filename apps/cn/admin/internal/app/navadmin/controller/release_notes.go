package controller

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/navadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

var releaseCommitSHA = regexp.MustCompile(`^[0-9a-f]{7,64}$`)

func normalizeUpdateNoticePayload(req models.UpdateNoticePayload) (models.UpdateNoticePayload, time.Time, common.Error) {
	req.Title, req.TitleEn = strings.TrimSpace(req.Title), strings.TrimSpace(req.TitleEn)
	req.Body, req.BodyEn = strings.TrimSpace(req.Body), strings.TrimSpace(req.BodyEn)
	req.Summary, req.SummaryEn = strings.TrimSpace(req.Summary), strings.TrimSpace(req.SummaryEn)
	req.Version, req.CommitSHA = normalizeStringPtr(req.Version), normalizeStringPtr(req.CommitSHA)
	if req.Version != nil && utf8.RuneCountInString(*req.Version) > 64 {
		return req, time.Time{}, common.NewValidationError("version must be at most 64 characters")
	}
	if req.CommitSHA != nil {
		normalized := strings.ToLower(*req.CommitSHA)
		req.CommitSHA = &normalized
		if !releaseCommitSHA.MatchString(normalized) {
			return req, time.Time{}, common.NewValidationError("commit_sha must be 7 to 64 hexadecimal characters")
		}
	}
	publishedAt, err := parseReleasePublicationTime(req.PublishedAt)
	return req, publishedAt, err
}

// Release dates remain China-site wall timestamps, including offset-bearing input.
// Empty/null is valid for a draft and means Publish Now on the publish endpoint.
func parseReleasePublicationTime(value string) (time.Time, common.Error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Time{}, common.NewValidationError("publication timezone unavailable")
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02 15:04"} {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			return parsed.In(location).Truncate(time.Second), nil
		}
	}
	return time.Time{}, common.NewValidationError("invalid published_at")
}

func validateUpdateNoticePublication(title, body string) common.Error {
	if strings.TrimSpace(title) == "" || strings.TrimSpace(body) == "" {
		return common.NewValidationError("Chinese title and body are required to publish")
	}
	return nil
}

func (api *navAPI) PublishUpdateNotice(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req struct {
		PublishedAt string `json:"published_at"`
	}
	if len(c.Body()) > 0 {
		if err := adminutil.DecodeBody(c, &req); err != nil {
			return common.NewResponse(c).Error(err)
		}
	}
	at, parseErr := parseReleasePublicationTime(req.PublishedAt)
	if parseErr != nil {
		return common.NewResponse(c).Error(parseErr)
	}
	item, storeErr := api.store.publishUpdateNotice(c.Context(), audit.MetaFromFiber(c), id, at)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return common.NewResponse(c).SuccessWithData(item)
}

func (api *navAPI) UnpublishUpdateNotice(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	item, storeErr := api.store.unpublishUpdateNotice(c.Context(), audit.MetaFromFiber(c), id)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return common.NewResponse(c).SuccessWithData(item)
}

func (store *navStore) publishUpdateNotice(ctx context.Context, meta audit.Meta, id int64, at time.Time) (models.UpdateNotice, common.Error) {
	var result models.UpdateNotice
	err := store.mutate(ctx, meta, "publish", "gfn_nav_update_notice", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.LockUpdateNotice(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		if err := validateUpdateNoticePublication(before.Title, before.Body); err != nil {
			return id, nil, nil, err
		}
		after, err := q.PublishUpdateNotice(ctx, navsqlc.PublishUpdateNoticeParams{ID: id, PublishedAt: navTimestamp(at)})
		result = updateNoticeModel(after)
		return id, before, after, err
	})
	return result, err
}

func (store *navStore) unpublishUpdateNotice(ctx context.Context, meta audit.Meta, id int64) (models.UpdateNotice, common.Error) {
	var result models.UpdateNotice
	err := store.mutate(ctx, meta, "unpublish", "gfn_nav_update_notice", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, err := q.LockUpdateNotice(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		after, err := q.UnpublishUpdateNotice(ctx, id)
		result = updateNoticeModel(after)
		return id, before, after, err
	})
	return result, err
}
