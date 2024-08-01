// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package application

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"git.sr.ht/~kota/kudoer/application/validator"
	"git.sr.ht/~kota/kudoer/db/models"
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
		app.authenticated(r),
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
		Username:    strings.TrimSpace(r.PostForm.Get("username")),
		DisplayName: strings.TrimSpace(r.PostForm.Get("displayname")),
		Email:       strings.TrimSpace(r.PostForm.Get("email")),
	}

	v := validator.New()
	v.Username(form.Username)
	v.Optional(form.DisplayName, v.DisplayName)
	v.Optional(form.Email, v.Email)

	password := r.PostForm.Get("password")
	confirmation := r.PostForm.Get("confirmation")
	v.Password(password, confirmation)

	// Set displayname to username if missing.
	if strings.TrimSpace(form.DisplayName) == "" {
		form.DisplayName = form.Username
	}

	validationError := func() {
		app.render(w, http.StatusUnprocessableEntity, "userRegister.tmpl", userRegisterPage{
			Page: app.newPage(
				r,
				"Register an account on Kudoer",
				"Register a new account on Kudoer where you can give kudos to your favorite things!",
			),
			Form: form,
		})
	}
	var valid bool
	if _, form.FieldErrors, valid = v.Valid(); !valid {
		validationError()
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
		v.AddFieldError("username", "Username is already taken")
		var valid bool
		if _, form.FieldErrors, valid = v.Valid(); !valid {
			validationError()
			return
		}
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.login(r, form.Username)
	if err != nil {
		app.serverError(w, err)
		return
	}

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

	v := validator.New()
	v.Username(form.Username)

	validationError := func() {
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
	var valid bool
	if form.NonFieldErrors, form.FieldErrors, valid = v.Valid(); !valid {
		validationError()
		return
	}

	password := r.PostForm.Get("password")
	err = app.users.Authenticate(r.Context(), form.Username, password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			v.AddNonFieldError("Email or password is incorrect")
			form.NonFieldErrors = v.NonFieldErrors
			validationError()
		} else {
			app.serverError(w, err)
		}
		return
	}

	err = app.login(r, form.Username)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type userForgotPage struct {
	Page
	Form userForgotForm
}

func (app *application) userForgotHandler(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "forgot.tmpl", userForgotPage{
		Page: app.newPage(
			r,
			"Forgot your password?",
			"Provide an email to reset your password",
		),
		Form: userForgotForm{},
	})
}

type userForgotForm struct {
	Username string
	Email    string

	// NonFieldErrors stores errors which do not relate to a form field.
	NonFieldErrors []string
	// FieldErrors stores errors relating to specific form fields.
	FieldErrors map[string]string
}

func (app *application) userForgotPostHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := userForgotForm{
		Username:       r.PostForm.Get("username"),
		Email:          r.PostForm.Get("email"),
		NonFieldErrors: []string{},
		FieldErrors:    map[string]string{},
	}

	v := validator.New()
	v.Username(form.Username)
	v.Email(form.Email)

	var valid bool
	if form.NonFieldErrors, form.FieldErrors, valid = v.Valid(); !valid {
		app.render(w, http.StatusUnprocessableEntity, "forgot.html", userForgotPage{
			Page: app.newPage(
				r,
				"Forgot your password?",
				"Provide an email to reset your password",
			),
			Form: form,
		})
	}

	email, err := app.users.GetEmail(r.Context(), form.Username)
	flashMsg := "If that email is in our system for your user instructions will be sent shortly"
	if email == "" {
		// Lie about it to prevent attackers from being able to "confirm" a
		// user's email address.
		app.flash(r, flashMsg)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	token, err := app.pwresets.New(r.Context(), form.Username)
	if err != nil {
		app.serverError(w, err)
		return
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				app.errLog.Println(err)
			}
		}()

		// Emails can be case sensitive; so we use the stored email rather than
		// the given email.
		err = app.mailer.PasswordReset(email, token)
		if err != nil {
			app.errLog.Println(err)
		}
	}()

	app.flash(r, flashMsg)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type userResetPage struct {
	Page
	Form userResetForm
}

func (app *application) userResetHandler(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "resetPassword.tmpl", userResetPage{
		Page: app.newPage(
			r,
			"Reset your password",
			"Enter a new password for your account",
		),
		Form: userResetForm{},
	})
}

type userResetForm struct {
	// NonFieldErrors stores errors which do not relate to a form field.
	NonFieldErrors []string
	// FieldErrors stores errors relating to specific form fields.
	FieldErrors map[string]string
}

func (app *application) userResetPostHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := userResetForm{}

	password := r.PostForm.Get("password")
	confirmation := r.PostForm.Get("confirmation")

	v := validator.New()
	v.Password(password, confirmation)

	// If the user is not logged in, they must provide a token which they
	// would've recieved in a password reset email.
	//
	// The Token is not part of the URL.
	//
	// This is to absolutely prevent the token from being leaked via the
	// referrer header.
	token := r.PostForm.Get("token")
	var username string
	if token == "" {
		username = app.authenticated(r)
		v.Username(username)
	} else {
		username, err = app.pwresets.Validate(r.Context(), token)
		if err != nil || username == "" {
			v.AddNonFieldError("Token is invalid")
		}
	}

	var valid bool
	if form.NonFieldErrors, form.FieldErrors, valid = v.Valid(); !valid {
		app.render(w, http.StatusUnprocessableEntity, "resetPassword.tmpl", userResetPage{
			Page: app.newPage(
				r,
				"Reset your password",
				"Enter a new password for your account",
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
	app.users.ChangePassword(r.Context(), username, string(hashedPassword))

	err = app.pwresets.DeleteAllUser(r.Context(), username)
	if err != nil {
		app.errLog.Println(err)
	}

	err = app.destroySessions(username)
	if err != nil {
		app.serverError(w, err)
	}

	err = app.login(r, username)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogoutPostHandler(w http.ResponseWriter, r *http.Request) {
	err := app.logout(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.flash(r, "You've been logged out successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type userSettingsPage struct {
	Page
	Username string
	Form     userSettingsForm
}

func (app *application) userSettingsHandler(w http.ResponseWriter, r *http.Request) {
	username := app.authenticated(r)
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
	err := r.ParseMultipartForm(1024 * 1024 * 5) // Ram cap, not total.
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	username := app.authenticated(r)
	form := userSettingsForm{
		DisplayName: r.PostForm.Get("displayname"),
		Email:       r.PostForm.Get("email"),
		Bio:         r.PostForm.Get("bio"),
		FieldErrors: map[string]string{},
	}

	v := validator.New()

	file, fileHeader, err := r.FormFile("pic")
	if err == nil {
		defer file.Close()

		if fileHeader.Size > (1024 * 1024 * 50) {
			v.AddFieldError("pic", "Profile picture must be less than 50MB")
		}

		// Store the profile picture variants.
		filename512, filename128, err := app.mediaStore.StorePic(file)
		if err != nil {
			app.serverError(w, err)
			return
		}

		// Update the user's pic.
		err = app.profilepics.Set(
			r.Context(),
			username,
			filename512, filename128,
		)
		if err != nil {
			app.serverError(w, err)
			return
		}
	} else if !errors.Is(err, http.ErrMissingFile) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	v.Optional(form.DisplayName, v.DisplayName)
	v.Optional(form.Email, v.Email)
	v.Optional(form.Bio, v.Bio)

	var valid bool
	if form.NonFieldErrors, form.FieldErrors, valid = v.Valid(); !valid {
		app.render(w, http.StatusUnprocessableEntity, "userSettings.tmpl", userSettingsPage{
			Page: app.newPage(
				r,
				"Editing your profile",
				"Change your profile settings",
			),
			Username: username,
			Form:     form,
		})
		return
	}

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

	username := app.authenticated(r)
	toFollow := r.PostForm.Get("follow")

	err = app.users.Follow(r.Context(), username, toFollow)
	if err != nil && !errors.Is(err, models.ErrAlreadyFollowing) {
		app.serverError(w, err)
		return
	}

	app.flash(r, "You're now following "+toFollow)
	http.Redirect(w, r, fmt.Sprintf("/user/view/%v", toFollow), http.StatusSeeOther)
}

func (app *application) userUnfollowPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	username := app.authenticated(r)
	toFollow := r.PostForm.Get("unfollow")

	err = app.users.Unfollow(r.Context(), username, toFollow)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.flash(r, "You're no longer following "+toFollow)
	http.Redirect(w, r, fmt.Sprintf("/user/view/%v", toFollow), http.StatusSeeOther)
}
