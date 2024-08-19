package application

import (
	"net/http"
)

type adminDashboardPage struct {
	Page

	UserCount       int
	KudoCount       int
	CustomItemCount int
}

// adminDashboardHandler displays the dashboard for admins.
func (app *application) adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	userCount, err := app.users.Count(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	kudoCount, err := app.kudos.Count(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	itemCount, err := app.items.CustomCount(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, http.StatusOK, "adminDashboard.tmpl", adminDashboardPage{
		Page: app.newPage(
			r,
			"Admin dashboard - Kudoer",
			"Manage accounts and admin features using this dashboard.",
		),
		UserCount:       userCount,
		KudoCount:       kudoCount,
		CustomItemCount: itemCount,
	})
}
