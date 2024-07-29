// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package validator

import (
	"regexp"
	"strconv"
	"unicode/utf8"

	emojis "git.sr.ht/~kota/kudoer/application/emoji"
	"git.sr.ht/~kota/kudoer/application/frames"
)

var rxUsername = regexp.MustCompile("^[a-z0-9_-]+$")
var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	// NonFieldErrors stores errors which do not relate to a form field.
	NonFieldErrors []string
	// FieldErrors stores errors relating to specific form fields.
	FieldErrors map[string]string
}

// New returns a new Validator instance ready to be used.
func New() *Validator {
	return &Validator{
		NonFieldErrors: []string{},
		FieldErrors:    make(map[string]string),
	}
}

// Valid returns true if there were any errors along with the errors.
func (v *Validator) Valid() ([]string, map[string]string, bool) {
	valid := len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
	return v.NonFieldErrors, v.FieldErrors, valid
}

// AddFieldError adds an error message to a field manually.
func (v *Validator) AddFieldError(key, message string) {
	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// AddNonFieldError adds an error message which does not relate to a specific
// field.
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// Check a condition.
// If it fails (returns false), then add the given field error.
// If a blank key was given, the error is instead added to the overall non-field
// errors.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		if key != "" {
			v.AddFieldError(key, message)
		} else {
			v.AddNonFieldError(message)
		}
	}
}

// Optional runs checks only if the field is non-blank.
//
// For example, to validate emails only if the string is non-blank:
//
//	v.Optional(form.Email, v.Email)
func (v *Validator) Optional(value string, fn func(value string)) {
	if value != "" {
		fn(value)
	}
}

// Username runs validation on the "username" field.
func (v *Validator) Username(username string) {
	v.Check(username != "", "username", "Username cannot be blank")
	v.Check(
		utf8.RuneCountInString(username) <= 30,
		"username",
		"Username cannot be longer than 30 characters",
	)
	v.Check(
		rxUsername.MatchString(username),
		"username",
		"Username may only contain lowercase letters, numbers, hyphen, and underscore",
	)
}

// Email runs validation on the "email" field.
func (v *Validator) Email(email string) {
	v.Check(email != "", "email", "Email cannot be blank")
	// https://stackoverflow.com/questions/386294/what-is-the-maximum-length-of-a-valid-email-address
	v.Check(len(email) < 254, "email", "Email is too long")
	v.Check(rxEmail.MatchString(email), "email", "Email appears to be invalid")
}

// DisplayName runs validation on the "displayname" field.
func (v *Validator) DisplayName(displayname string) {
	v.Check(displayname != "", "displayname", "DisplayName cannot be blank")
	v.Check(
		utf8.RuneCountInString(displayname) <= 30,
		"displayname",
		"Display Name cannot be longer than 30 characters",
	)
}

// Password checks if a password is valid for use and that the confirmation
// password matches.
//
// IMPORTANT: This is not validating the password in the database. We're simply
// checking if a password is _able to be used at all_.
func (v *Validator) Password(password, confirmation string) {
	v.Check(password != "", "password", "Password cannot be blank")
	v.Check(
		len(password) <= 72,
		"password",
		"Password cannot be larger than 72 bytes as a limitation of bcrypt",
	)
	v.Check(
		password == confirmation,
		"password",
		"Password and confirmation do not match",
	)
}

// Bio runs validation on the bio field.
func (v *Validator) Bio(bio string) {
	v.Check(bio != "", "bio", "Bio cannot be blank")
	v.Check(
		utf8.RuneCountInString(bio) <= 1000,
		"bio",
		"Bio cannot be longer than 1000 characters",
	)
}

// ItemName runs validation on an item's name.
func (v *Validator) ItemName(name string) {
	v.Check(name != "", "name", "Name cannot be blank")
	v.Check(
		utf8.RuneCountInString(name) <= 100,
		"name",
		"Name cannot be longer than 100 characters",
	)
}

// ItemDescription runs validation on an item's description.
func (v *Validator) ItemDescription(desc string) {
	v.Check(desc != "", "description", "Description cannot be blank")
	v.Check(
		utf8.RuneCountInString(desc) <= 1000,
		"description",
		"Description cannot be longer than 1000 characters",
	)
}

// Kudo runs validation on all the kudo fields.
// If an error is found it is added as a "kudo" field error.
// Parsed fields are returned.
func (v *Validator) Kudo(emoji, frame, body string) (int, int, string) {
	e, err := strconv.Atoi(emoji)
	if err != nil {
		v.AddFieldError("kudo", "Invalid emoji payload")
	}
	v.Check(emojis.Validate(e), "kudo", "Invalid emoji selected")

	f, err := strconv.Atoi(frame)
	if err != nil {
		v.AddFieldError("kudo", "Invalid frame payload")
	}
	v.Check(frames.Validate(f), "kudo", "Invalid frame selected")

	v.Check(
		utf8.RuneCountInString(body) <= 5500,
		"kudo",
		"Body of kudo cannot be longer than 5000 characters",
	)
	return e, f, body
}
