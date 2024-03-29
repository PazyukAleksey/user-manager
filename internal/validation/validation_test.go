package validation

import (
	"testing"
)

func TestNicknameValidation(t *testing.T) {
	tests := map[string]bool{
		"asdaASW":        true,
		"asdaASW12313!!": false,
	}

	for input, expected := range tests {
		result := NicknameValidation(input)
		if result != expected {
			t.Errorf("NicknameValidation(%s) = %v, expected %v", input, result, expected)
		}
	}
}

func TestNameValidation(t *testing.T) {
	tests := map[string]bool{
		"Oleksii":    true,
		"Oleksii123": false,
	}

	for input, expected := range tests {
		result := NameValidation(input)
		if result != expected {
			t.Errorf("NameValidation(%s) = %v, expected %v", input, result, expected)
		}
	}
}

func TestPasswordValidation(t *testing.T) {
	tests := map[string]bool{
		"Qwerty1123@#": true,
		"Aaaasssdd":    false,
	}

	for input, expected := range tests {
		result := PasswordValidation(input)
		if result != expected {
			t.Errorf("PasswordValidation(%s) = %v, expected %v", input, result, expected)
		}
	}
}

func TestEmailValidation(t *testing.T) {
	tests := map[string]bool{
		"validEmail@example.com": true,
		"invalidEmail@.com":      false,
	}

	for input, expected := range tests {
		result := EmailValidation(input)
		if result != expected {
			t.Errorf("EmailValidation(%s) = %v, expected %v", input, result, expected)
		}
	}
}
