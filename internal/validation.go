package internal

import (
	"net/mail"
	"regexp"
)

const (
	nicknameRegex        = "^[a-zA-Z0-9_-]{4,12}$"
	nameRegex            = "^[a-zA-Z]{3,16}$"
	passwordRegexLength  = `^[\d\D\S]{6,16}$`
	passwordRegexUpper   = `[A-Z]`
	passwordRegexDigit   = `[\d]`
	passwordRegexSpecial = `[\W_]`
)

func NicknameValidation(s string) bool {
	regexpObj := regexp.MustCompile(nicknameRegex)
	if !regexpObj.MatchString(s) {
		return false
	}
	return true
}
func NameValidation(s string) bool {
	regexpObj := regexp.MustCompile(nameRegex)
	if !regexpObj.MatchString(s) {
		return false
	}
	return true
}
func PasswordValidation(s string) bool {
	regexPatterns := [4]string{passwordRegexLength, passwordRegexUpper, passwordRegexDigit, passwordRegexSpecial}
	for i := 0; i < cap(regexPatterns); i++ {
		regexPattern := regexPatterns[i]
		regexpObj := regexp.MustCompile(regexPattern)
		if !regexpObj.MatchString(s) {
			return false
		}
	}
	return true
}
func EmailValidation(s string) bool {
	_, err := mail.ParseAddress(s)
	if err != nil {
		return false
	}
	return true
}
