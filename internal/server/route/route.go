package route

import (
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/middlewares"
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/server/service"
	"github.com/go-chi/chi"
)

func New(service service.Servicer, middleware *middlewares.Middlewares) *chi.Mux {

	// New router
	route := chi.NewRouter()

	// Handlers
	route.Get("/api/user_banner", middleware.UserAuthorization(service.GetUserBanner))    // Getting user banner
	route.Post("/api/banner", middleware.AdminAuthorization(service.CreateBanner))        // Create new banner
	route.Patch("/api/banner/{id}", middleware.AdminAuthorization(service.UpdateBanner))  // Update banner exists
	route.Get("/api/banner", middleware.AdminAuthorization(service.GetBanners))           // Get banners
	route.Delete("/api/banner/{id}", middleware.AdminAuthorization(service.DeleteBanner)) // Delete banner

	route.Get("/api/history_banner/{id}", middleware.AdminAuthorization(service.GetHistoryBanner))
	route.Post("/api/version_banner", middleware.AdminAuthorization(service.UpdateVersion))
	return route
}
