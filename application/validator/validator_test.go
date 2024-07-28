package validator

import (
	"testing"
)

func TestUsername(t *testing.T) {
	type test struct {
		description string
		input       string
		valid       bool
		errMsg      string
	}

	tests := []test{
		{
			description: "Basic valid username",
			input:       "kota",
			valid:       true,
			errMsg:      "",
		},
		{
			description: "Blank",
			input:       "",
			valid:       false,
			errMsg:      "Username cannot be blank",
		},
		{
			description: "Weird username",
			input:       "a-_1",
			valid:       true,
			errMsg:      "",
		},
		{
			description: "Invalid",
			input:       "@",
			valid:       false,
			errMsg:      "Username may only contain lowercase letters, numbers, hyphen, and underscore",
		},
	}

	for _, tc := range tests {
		v := New()
		v.Username(tc.input)
		_, _, valid := v.Valid()

		var errMsg string
		if !valid {
			for _, e := range v.FieldErrors {
				errMsg = e
				break
			}
		}

		if valid != tc.valid {
			t.Fatalf(
				"%v: got: \"%v\" want: \"%v\" msg: \"%v\"\n",
				tc.description,
				valid,
				tc.valid,
				errMsg,
			)
		}
		if errMsg != tc.errMsg {
			t.Fatalf(
				"%v: msg: \"%v\" wanted msg: \"%v\"\n",
				tc.description,
				errMsg,
				tc.errMsg,
			)
		}
	}
}

func TestEmail(t *testing.T) {
	type test struct {
		description string
		input       string
		valid       bool
		errMsg      string
	}

	tests := []test{
		{
			description: "Basic valid email",
			input:       "a@gmail.com",
			valid:       true,
			errMsg:      "",
		},
		{
			description: "Blank",
			input:       "",
			valid:       false,
			errMsg:      "Email cannot be blank",
		},
		{
			description: "Invalid",
			input:       "@@@",
			valid:       false,
			errMsg:      "Email appears to be invalid",
		},
	}

	for _, tc := range tests {
		v := New()
		v.Email(tc.input)
		_, _, valid := v.Valid()

		var errMsg string
		if !valid {
			for _, e := range v.FieldErrors {
				errMsg = e
				break
			}
		}

		if valid != tc.valid {
			t.Fatalf(
				"%v: got: \"%v\" want: \"%v\" wantErr \"%v\"\n",
				tc.description,
				valid,
				tc.valid,
				tc.errMsg,
			)
		}
		if errMsg != tc.errMsg {
			t.Fatalf(
				"%v: msg: \"%v\" wanted msg: \"%v\"\n",
				tc.description,
				errMsg,
				tc.errMsg,
			)
		}
	}
}

func TestDisplayName(t *testing.T) {
	type test struct {
		description string
		input       string
		valid       bool
		errMsg      string
	}

	tests := []test{
		{
			description: "Basic valid DisplayName",
			input:       "Kota",
			valid:       true,
			errMsg:      "",
		},
		{
			description: "Too long DisplayName",
			input:       "1234567890123456789012345678901",
			valid:       false,
			errMsg:      "Display Name cannot be longer than 30 characters",
		},
	}

	for _, tc := range tests {
		v := New()
		v.DisplayName(tc.input)
		_, _, valid := v.Valid()

		var errMsg string
		if !valid {
			for _, e := range v.FieldErrors {
				errMsg = e
				break
			}
		}

		if valid != tc.valid {
			t.Fatalf(
				"%v: got: \"%v\" want: \"%v\" wantErr \"%v\"\n",
				tc.description,
				valid,
				tc.valid,
				tc.errMsg,
			)
		}
		if errMsg != tc.errMsg {
			t.Fatalf(
				"%v: msg: \"%v\" wanted msg: \"%v\"\n",
				tc.description,
				errMsg,
				tc.errMsg,
			)
		}
	}
}

func TestPassword(t *testing.T) {
	type test struct {
		description  string
		input        string
		confirmation string
		valid        bool
		errMsg       string
	}

	tests := []test{
		{
			description:  "Basic valid Password",
			input:        "hunter12",
			confirmation: "hunter12",
			valid:        true,
			errMsg:       "",
		},
		{
			description:  "Failed confirmation",
			input:        "hunter12",
			confirmation: "hunter21",
			valid:        false,
			errMsg:       "Password and confirmation do not match",
		},
		{
			description:  "Too long Password",
			input:        "1234567890123456789012345678901234567890123456789012345678901234567890123",
			confirmation: "1234567890123456789012345678901234567890123456789012345678901234567890123",
			valid:        false,
			errMsg:       "Password cannot be larger than 72 bytes as a limitation of bcrypt",
		},
	}

	for _, tc := range tests {
		v := New()
		v.Password(tc.input, tc.confirmation)
		_, _, valid := v.Valid()

		var errMsg string
		if !valid {
			for _, e := range v.FieldErrors {
				errMsg = e
				break
			}
		}

		if valid != tc.valid {
			t.Fatalf(
				"%v: got: \"%v\" want: \"%v\" wantErr \"%v\"\n",
				tc.description,
				valid,
				tc.valid,
				tc.errMsg,
			)
		}
		if errMsg != tc.errMsg {
			t.Fatalf(
				"%v: msg: \"%v\" wanted msg: \"%v\"\n",
				tc.description,
				errMsg,
				tc.errMsg,
			)
		}
	}
}

func TestBio(t *testing.T) {
	type test struct {
		description string
		input       string
		valid       bool
		errMsg      string
	}

	tests := []test{
		{
			description: "Basic valid Bio",
			input:       "Nalgene waterbottles",
			valid:       true,
			errMsg:      "",
		},
		{
			description: "Blank",
			input:       "",
			valid:       false,
			errMsg:      "Bio cannot be blank",
		},
	}

	for _, tc := range tests {
		v := New()
		v.Bio(tc.input)
		_, _, valid := v.Valid()

		var errMsg string
		if !valid {
			for _, e := range v.FieldErrors {
				errMsg = e
				break
			}
		}

		if valid != tc.valid {
			t.Fatalf(
				"%v: got: \"%v\" want: \"%v\" wantErr \"%v\"\n",
				tc.description,
				valid,
				tc.valid,
				tc.errMsg,
			)
		}
		if errMsg != tc.errMsg {
			t.Fatalf(
				"%v: msg: \"%v\" wanted msg: \"%v\"\n",
				tc.description,
				errMsg,
				tc.errMsg,
			)
		}
	}
}

func TestItemName(t *testing.T) {
	type test struct {
		description string
		input       string
		valid       bool
		errMsg      string
	}

	tests := []test{
		{
			description: "Basic valid ItemName",
			input:       "Nalgene",
			valid:       true,
			errMsg:      "",
		},
		{
			description: "Blank",
			input:       "",
			valid:       false,
			errMsg:      "Name cannot be blank",
		},
	}

	for _, tc := range tests {
		v := New()
		v.ItemName(tc.input)
		_, _, valid := v.Valid()

		var errMsg string
		if !valid {
			for _, e := range v.FieldErrors {
				errMsg = e
				break
			}
		}

		if valid != tc.valid {
			t.Fatalf(
				"%v: got: \"%v\" want: \"%v\" wantErr \"%v\"\n",
				tc.description,
				valid,
				tc.valid,
				tc.errMsg,
			)
		}
		if errMsg != tc.errMsg {
			t.Fatalf(
				"%v: msg: \"%v\" wanted msg: \"%v\"\n",
				tc.description,
				errMsg,
				tc.errMsg,
			)
		}
	}
}

func TestItemDescription(t *testing.T) {
	type test struct {
		description string
		input       string
		valid       bool
		errMsg      string
	}

	tests := []test{
		{
			description: "Basic valid ItemDescription",
			input:       "Nalgene",
			valid:       true,
			errMsg:      "",
		},
		{
			description: "Blank",
			input:       "",
			valid:       false,
			errMsg:      "Description cannot be blank",
		},
	}

	for _, tc := range tests {
		v := New()
		v.ItemDescription(tc.input)
		_, _, valid := v.Valid()

		var errMsg string
		if !valid {
			for _, e := range v.FieldErrors {
				errMsg = e
				break
			}
		}

		if valid != tc.valid {
			t.Fatalf(
				"%v: got: \"%v\" want: \"%v\" wantErr \"%v\"\n",
				tc.description,
				valid,
				tc.valid,
				tc.errMsg,
			)
		}
		if errMsg != tc.errMsg {
			t.Fatalf(
				"%v: msg: \"%v\" wanted msg: \"%v\"\n",
				tc.description,
				errMsg,
				tc.errMsg,
			)
		}
	}
}
