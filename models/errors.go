// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package models

import "errors"

var ErrNoRecord = errors.New("model: no matching record found")
var ErrUsernameExists = errors.New("model: that username already exists")
var ErrAlreadyFollowing = errors.New("model: already following this user")
var ErrInvalidCredentials = errors.New("model: submitted credentials are invalid")
var ErrPWResetTokenInvalid = errors.New("model: password reset token missing or invalid")
