// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"git.sr.ht/~kota/kudoer/models"
	"golang.org/x/crypto/bcrypt"
)

type userViewPage struct {
	Page
	models.User

	// Is the logged in user following the user being viewed?
	Following bool

	// All kudos this user has given.
	Kudos []models.Kudo
}

// userViewHandler presents a user.
func (app *application) userViewHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	user, err := app.users.Info(r.Context(), username)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, err)
		}
		return
	}

	following, err := app.users.IsFollowing(
		r.Context(),
		app.sessionManager.GetString(r.Context(), "authenticatedUsername"),
		username,
	)
	if err != nil {
		app.serverError(w, err)
		return
	}

	kudos, err := app.kudos.User(r.Context(), username)
	if err != nil {
		app.serverError(w, err)
		return
	}

	title := user.DisplayName + " - Kudoer"
	desc := "Viewing " + user.DisplayName + " on Kudoer"
	app.render(w, http.StatusOK, "userView.tmpl", userViewPage{
		Page:      app.newPage(r, title, desc),
		User:      user,
		Following: following,
		Kudos:     kudos,
	})
}

type userRegisterPage struct {
	Page
	Form userRegisterForm
}

// userRegisterHandler presents a web form to add a user.
func (app *application) userRegisterHandler(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "userRegister.tmpl", userRegisterPage{
		Page: app.newPage(
			r,
			"Register an account on Kudoer",
			"Register a new account on Kudoer where you can give kudos to your favorite things!",
		),
		Form: userRegisterForm{},
	})
}

type userRegisterForm struct {
	Username    string
	DisplayName string
	Email       string

	// NonFieldErrors stores errors which do not relate to a form field.
	NonFieldErrors []string
	// FieldErrors stores errors relating to specific form fields.
	FieldErrors map[string]string
}

// userRegisterPostHandler adds a user.
func (app *application) userRegisterPostHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := userRegisterForm{
		Username:    r.PostForm.Get("username"),
		DisplayName: r.PostForm.Get("displayname"),
		Email:       r.PostForm.Get("email"),
		FieldErrors: map[string]string{},
	}

	if strings.TrimSpace(form.Username) == "" {
		form.FieldErrors["username"] = "Username cannot be blank"
	} else if utf8.RuneCountInString(form.Username) > 30 {
		form.FieldErrors["username"] = "Username cannot be longer than 30 characters"
	} else if !rxUsername.MatchString(form.Username) {
		form.FieldErrors["username"] = "Username may only contain lowercase letters, numbers, hyphen, and underscore"
	}

	if strings.TrimSpace(form.DisplayName) == "" {
		form.DisplayName = form.Username
	} else if utf8.RuneCountInString(form.DisplayName) > 30 {
		form.FieldErrors["displayname"] = "Display Name cannot be longer than 30 characters"
	}

	if form.Email != "" {
		if len(form.Email) > 254 || !rxEmail.MatchString(form.Email) {
			// https://stackoverflow.com/questions/386294/what-is-the-maximum-length-of-a-valid-email-address
			form.FieldErrors["email"] = "Email appears to be invalid"
		}
	}

	password := r.PostForm.Get("password")
	if strings.TrimSpace(password) == "" {
		form.FieldErrors["password"] = "Password cannot be blank"
	} else if len(password) > 72 {
		form.FieldErrors["password"] = "Password cannot be larger than 72 bytes as a limitation of bcrypt"
	}

	if len(form.FieldErrors) > 0 {
		app.render(w, http.StatusUnprocessableEntity, "userRegister.tmpl", userRegisterPage{
			Page: app.newPage(
				r,
				"Register an account on Kudoer",
				"Register a new account on Kudoer where you can give kudos to your favorite things!",
			),
			Form: form,
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.users.Register(
		r.Context(),
		form.Username,
		form.DisplayName,
		form.Email,
		string(hashedPassword),
	)
	if errors.Is(err, models.ErrUsernameExists) {
		form.FieldErrors["username"] = "Username is already taken"
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUsername", form.Username)
	http.Redirect(w, r, fmt.Sprintf("/user/view/%v", form.Username), http.StatusSeeOther)
}

type userLoginPage struct {
	Page
	Form userLoginForm
}

func (app *application) userLoginHandler(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "login.tmpl", userLoginPage{
		Page: app.newPage(
			r,
			"Login on Kudoer",
			"Provide your login details to access Kudoer",
		),
		Form: userLoginForm{},
	})
}

type userLoginForm struct {
	Username string

	// NonFieldErrors stores errors which do not relate to a form field.
	NonFieldErrors []string
	// FieldErrors stores errors relating to specific form fields.
	FieldErrors map[string]string
}

func (app *application) userLoginPostHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := userLoginForm{
		Username:       r.PostForm.Get("username"),
		NonFieldErrors: []string{},
		FieldErrors:    map[string]string{},
	}

	if strings.TrimSpace(form.Username) == "" {
		form.FieldErrors["username"] = "Username cannot be blank"
	} else if utf8.RuneCountInString(form.Username) > 30 {
		form.FieldErrors["username"] = "Username cannot be longer than 30 characters"
	} else if !rxUsername.MatchString(form.Username) {
		form.FieldErrors["username"] = "Username may only contain lowercase letters, numbers, hyphen, and underscore"
	}

	password := r.PostForm.Get("password")
	if strings.TrimSpace(password) == "" {
		form.FieldErrors["password"] = "Password cannot be blank"
	} else if len(password) > 72 {
		form.FieldErrors["password"] = "Password cannot be larger than 72 bytes as a limitation of bcrypt"
	}

	if len(form.FieldErrors) > 0 {
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl", userLoginPage{
			Page: app.newPage(
				r,
				"Login on Kudoer",
				"Provide your login details to access Kudoer",
			),
			Form: form,
		})
		return
	}

	err = app.users.Authenticate(r.Context(), form.Username, password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.NonFieldErrors = append(
				form.NonFieldErrors,
				"Email or password is incorrect",
			)
			app.render(w, http.StatusUnprocessableEntity, "login.tmpl", userLoginPage{
				Page: app.newPage(
					r,
					"Login on Kudoer",
					"Provide your login details to access Kudoer",
				),
				Form: form,
			})
		} else {
			app.serverError(w, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUsername", form.Username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogoutPostHandler(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUsername")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type userSettingsPage struct {
	Page
	Username string
	Form     userSettingsForm
}

func (app *application) userSettingsHandler(w http.ResponseWriter, r *http.Request) {
	username := app.sessionManager.GetString(r.Context(), "authenticatedUsername")
	user, err := app.users.Info(r.Context(), username)
	if err != nil {
		app.serverError(w, err)
		return
	}

	form := userSettingsForm{
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Bio:         user.Bio,
	}

	app.render(w, http.StatusOK, "userSettings.tmpl", userSettingsPage{
		Page: app.newPage(
			r,
			"Editing your profile",
			"Change your profile settings",
		),
		Username: user.Username,
		Form:     form,
	})
}

type userSettingsForm struct {
	DisplayName string
	Email       string
	Bio         string

	// NonFieldErrors stores errors which do not relate to a form field.
	NonFieldErrors []string
	// FieldErrors stores errors relating to specific form fields.
	FieldErrors map[string]string
}

func (app *application) userSettingsPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := userSettingsForm{
		DisplayName: r.PostForm.Get("displayname"),
		Email:       r.PostForm.Get("email"),
		Bio:         r.PostForm.Get("bio"),
		FieldErrors: map[string]string{},
	}

	if utf8.RuneCountInString(form.DisplayName) > 30 {
		form.FieldErrors["displayname"] = "Display Name cannot be longer than 30 characters"
	}

	if form.Email != "" {
		if len(form.Email) > 254 || !rxEmail.MatchString(form.Email) {
			// https://stackoverflow.com/questions/386294/what-is-the-maximum-length-of-a-valid-email-address
			form.FieldErrors["email"] = "Email appears to be invalid"
		}
	}

	if utf8.RuneCountInString(form.Bio) > 1000 {
		form.FieldErrors["bio"] = "Bio cannot be longer than 1000 characters"
	}

	password := r.PostForm.Get("password")
	if password != "" {
		if len(password) > 72 {
			form.FieldErrors["password"] = "Password cannot be larger than 72 bytes as a limitation of bcrypt"
		}
	}

	if len(form.FieldErrors) > 0 {
		app.render(w, http.StatusUnprocessableEntity, "userSettings.tmpl", userSettingsPage{
			Page: app.newPage(
				r,
				"Editing your profile",
				"Change your profile settings",
			),
			Form: form,
		})
		return
	}

	// Update non-password fields.
	username := app.sessionManager.GetString(r.Context(), "authenticatedUsername")
	err = app.users.UpdateProfile(
		r.Context(),
		username,
		form.DisplayName,
		form.Email,
		form.Bio,
	)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Update user's password ONLY if changed.
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			app.serverError(w, err)
			return
		}
		err = app.users.ChangePassword(
			r.Context(),
			username,
			string(hashedPassword),
		)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/user/view/%v", username), http.StatusSeeOther)
}

type userFollowersPage struct {
	Page
	models.User
	Users []models.User
}

func (app *application) userFollowersHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	user, err := app.users.Info(r.Context(), username)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, err)
		}
		return
	}

	users, err := app.users.Followers(r.Context(), username)
	if err != nil {
		app.serverError(w, err)
		return
	}

	title := user.DisplayName + " Followers - Kudoer"
	desc := "Followers of " + user.DisplayName + " on Kudoer"
	app.render(w, http.StatusOK, "userFollowers.tmpl", userFollowersPage{
		Page:  app.newPage(r, title, desc),
		User:  user,
		Users: users,
	})
}

type userFollowingPage struct {
	Page
	models.User
	Users []models.User
}

func (app *application) userFollowingHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	user, err := app.users.Info(r.Context(), username)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, err)
		}
		return
	}

	users, err := app.users.Following(r.Context(), username)
	if err != nil {
		app.serverError(w, err)
		return
	}

	title := user.DisplayName + " Following - Kudoer"
	desc := "Users " + user.DisplayName + " is following on Kudoer"
	app.render(w, http.StatusOK, "userFollowing.tmpl", userFollowingPage{
		Page:  app.newPage(r, title, desc),
		User:  user,
		Users: users,
	})
}

func (app *application) userFollowPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	username := app.sessionManager.GetString(r.Context(), "authenticatedUsername")
	toFollow := r.PostForm.Get("follow")

	err = app.users.Follow(r.Context(), username, toFollow)
	if err != nil && !errors.Is(err, models.ErrAlreadyFollowing) {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "You're now following "+toFollow)
	http.Redirect(w, r, fmt.Sprintf("/user/view/%v", toFollow), http.StatusSeeOther)
}

func (app *application) userUnfollowPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	username := app.sessionManager.GetString(r.Context(), "authenticatedUsername")
	toFollow := r.PostForm.Get("unfollow")

	err = app.users.Unfollow(r.Context(), username, toFollow)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "You're no longer following "+toFollow)
	http.Redirect(w, r, fmt.Sprintf("/user/view/%v", toFollow), http.StatusSeeOther)
}
