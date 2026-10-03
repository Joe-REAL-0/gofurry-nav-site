package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	authmw "github.com/gofurry/gofurry-admin/internal/app/auth/middleware"
	"github.com/gofurry/gofurry-admin/internal/bootstrap"
)

func showcaseRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	api := runtime.GameAPI
	read, write := authmw.Require(authorization.ContentRead), authmw.Require(authorization.ContentWrite)
	root.Get("/composition", read, api.ShowcaseProxy)
	root.Get("/candidates", read, api.ShowcaseProxy)
	root.Get("/analytics/quality", read, api.ShowcaseQuality)
	root.Get("/campaigns/:id/stats.csv", read, api.ShowcaseStats)
	root.Get("/campaigns/:id/stats", read, api.ShowcaseStats)
	root.Get("/campaigns", read, api.ListShowcaseCampaigns)
	root.Post("/campaigns", write, api.CreateShowcaseCampaign)
	root.Get("/campaigns/:id", read, api.GetShowcaseCampaign)
	root.Put("/campaigns/:id/content", write, api.SaveShowcaseContent)
	root.Put("/campaigns/:id/schedule", write, api.SaveShowcaseSchedule)
	root.Post("/campaigns/:id/artwork/:variant", write, api.ShowcaseArtwork)
	root.Delete("/campaigns/:id/artwork/:variant", write, api.ShowcaseArtwork)
	root.Delete("/campaigns/:id", write, api.DeleteShowcaseDraft)
	for _, action := range []string{"publish", "pause", "resume", "archive"} {
		root.Post("/campaigns/:id/"+action, write, api.ShowcaseLifecycle)
	}
}
