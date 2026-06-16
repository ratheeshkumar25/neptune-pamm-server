// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/pkg/utilis/utilis.go
// Role: Utilis function help to validate Phone number

package utilis

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// PhoneNumberValidation represent is to validate phone
func PhoneNumberValidation(f1 validator.FieldLevel) bool {
	fieldVal := f1.Field().String()
	match, _ := regexp.MatchString("^[0-9+-]+$", fieldVal)
	return match
}

// EmailValidation helps to validte the email user enters
func EmailValidation(f1 validator.FieldLevel) bool {
	fieldVal := f1.Field().String()
	match, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, fieldVal)
	return match
}
