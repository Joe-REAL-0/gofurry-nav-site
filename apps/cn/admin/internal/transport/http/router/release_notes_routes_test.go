package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	navadmin "github.com/gofurry/gofurry-admin/internal/app/navadmin/controller"
	"github.com/gofurry/gofurry-admin/internal/bootstrap"
	"net/http/httptest"
	"testing"
)

func TestReleaseNotePublicationRouteCapabilities(t *testing.T) {
	for _, tc := range []struct {
		name      string
		principal *authorization.Principal
		status    int
	}{
		{"writer", &authorization.Principal{Capabilities: []authorization.Capability{authorization.ContentWrite}}, 400},
		{"reader", &authorization.Principal{Capabilities: []authorization.Capability{authorization.ContentRead}}, 403},
		{"anonymous", nil, 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c fiber.Ctx) error {
				if tc.principal != nil {
					c.Locals(authorization.PrincipalContextKey, tc.principal)
				}
				return c.Next()
			})
			navRoutes(app.Group("/api/v1/nav"), &bootstrap.Runtime{NavAPI: navadmin.New(nil, nil)})
			for _, action := range []string{"publish", "unpublish"} {
				// Invalid ID reaches validation without storage; absent registration gives 404.
				response, err := app.Test(httptest.NewRequest("POST", "/api/v1/nav/update-notices/invalid/"+action, nil))
				if err != nil {
					t.Fatal(err)
				}
				response.Body.Close()
				if response.StatusCode != tc.status {
					t.Fatalf("%s: %d", action, response.StatusCode)
				}
			}
		})
	}
}
