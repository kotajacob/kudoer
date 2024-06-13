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

	Username string
	Email    string
}

// userView presents a user.
func (app *application) userView(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	user, err := app.users.Get(r.Context(), username)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.render(w, http.StatusOK, "userView.tmpl", userViewPage{
		Page:     app.newPage(r.Context()),
		Username: user.Username,
		Email:    user.Email,
	})
}

type userCreatePage struct {
	Page
	Form userCreateForm
}

// userCreate presents a web form to add a user.
func (app *application) userCreate(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "userCreate.tmpl", userCreatePage{
		Page: app.newPage(r.Context()),
		Form: userCreateForm{},
	})
}

type userCreateForm struct {
	Username string
	Email    string

	// NonFieldErrors stores errors which do not relate to a form field.
	NonFieldErrors []string
	// FieldErrors stores errors relating to specific form fields.
	FieldErrors map[string]string
}

// userCreatePost adds a user.
func (app *application) userCreatePost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := userCreateForm{
		Username:    r.PostForm.Get("username"),
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

	if len(form.Email) > 254 || !rxEmail.MatchString(form.Email) {
		// https://stackoverflow.com/questions/386294/what-is-the-maximum-length-of-a-valid-email-address
		form.FieldErrors["email"] = "Email appears to be invalid"
	}

	password := r.PostForm.Get("password")
	if strings.TrimSpace(password) == "" {
		form.FieldErrors["password"] = "Password cannot be blank"
	} else if len(password) > 72 {
		form.FieldErrors["password"] = "Password cannot be larger than 72 bytes as a limitation of bcrypt"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.users.Insert(
		r.Context(),
		form.Username,
		form.Email,
		string(hashedPassword),
	)
	if errors.Is(err, models.ErrUsernameExists) {
		form.FieldErrors["username"] = "Username is already taken"
	} else if errors.Is(err, models.ErrEmailExists) {
		form.FieldErrors["email"] = "Email is already in our system"
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	if len(form.FieldErrors) > 0 {
		app.render(w, http.StatusUnprocessableEntity, "userCreate.tmpl", userCreatePage{
			Page: app.newPage(r.Context()),
			Form: form,
		})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/user/view/%v", form.Username), http.StatusSeeOther)
}

type userLoginPage struct {
	Page
	Form userLoginForm
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "login.tmpl", userLoginPage{
		Page: app.newPage(r.Context()),
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

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
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
			Page: app.newPage(r.Context()),
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
				Page: app.newPage(r.Context()),
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

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUsername")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
