package models

import (
	"time"

	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
)

type Response struct {
	SchemaVersion int                `json:"schema_version"`
	GeneratedAt   time.Time          `json:"generated_at"`
	State         string             `json:"state"`
	SiteID        int64              `json:"site_id"`
	Items         []navmodels.SiteVo `json:"items"`
}
