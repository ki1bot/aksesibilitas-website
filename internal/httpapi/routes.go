package httpapi

import "github.com/go-chi/chi/v5"

func (server *Server) apiRoutes(router chi.Router) {
	router.Get("/health", server.health)

	router.Route(
		"/auth",
		func(router chi.Router) {
			router.Post(
				"/register",
				server.register,
			)

			router.Post(
				"/login",
				server.login,
			)
		},
	)

	router.Group(server.protectedRoutes)
}

func (server *Server) protectedRoutes(router chi.Router) {
	router.Use(server.authenticate)
	router.Use(server.csrf)

	router.Get(
		"/auth/me",
		server.me,
	)

	router.Patch(
		"/auth/me",
		server.updateProfile,
	)

	router.Post(
		"/auth/logout",
		server.logout,
	)

	router.Get(
		"/projects",
		server.listProjects,
	)

	router.Post(
		"/projects",
		server.createProject,
	)

	router.Route(
		"/projects/{projectId}",
		func(router chi.Router) {
			router.Get(
				"/",
				server.getProject,
			)

			router.Patch(
				"/",
				server.updateProject,
			)

			router.Delete(
				"/",
				server.deleteProject,
			)
		},
	)

	router.Get(
		"/scans",
		server.listScans,
	)

	router.Post(
		"/scans",
		server.createScan,
	)

	router.Route(
		"/scans/{scanId}",
		func(router chi.Router) {
			router.Get(
				"/",
				server.getScan,
			)

			router.Delete(
				"/",
				server.deleteScan,
			)

			router.Post(
				"/cancel",
				server.cancelScan,
			)

			router.Post(
				"/retry",
				server.retryScan,
			)

			router.Get(
				"/violations",
				server.listViolations,
			)

			router.Get(
				"/manual-review",
				server.getManualReview,
			)

			router.Post(
				"/reports",
				server.createReport,
			)
		},
	)

	router.Route(
		"/violations/{violationId}",
		func(router chi.Router) {
			router.Get(
				"/",
				server.getViolation,
			)

			router.Patch(
				"/",
				server.updateViolation,
			)
		},
	)

	router.Patch(
		"/manual-review/items/{itemId}",
		server.updateManualReviewItem,
	)

	router.Get(
		"/reports/{reportId}",
		server.getReport,
	)

	router.Get(
		"/reports/{reportId}/download",
		server.downloadReport,
	)
}
