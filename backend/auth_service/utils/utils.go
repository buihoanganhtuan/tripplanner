package utils

import (
	"fmt"
	"net/mail"
	"os"
)

// Manager type to manage environment variables
// Attempting to get a variable via Var() when the manager is in error state or
// without fetching the variable first via Fetch() will result in a panic
type EnvironmentVariableMap struct {
	varMap map[string]string
	err    error
}

func (ev *EnvironmentVariableMap) Fetch(names ...string) {
	if ev.err != nil {
		return
	}
	for _, name := range names {
		if _, exist := ev.varMap[name]; exist {
			continue
		}
		if val, ok := os.LookupEnv(name); ok {
			ev.varMap[name] = val
			continue
		}
		ev.err = fmt.Errorf("environment variable %v is unset", name)
	}
}

func (ev *EnvironmentVariableMap) Var(name string) string {
	if _, exist := ev.varMap[name]; ev.err != nil || !exist {
		if ev.err != nil {
			panic(ev.err)
		}
		panic(fmt.Errorf("environmental variable %v has not been fetched", name))
	}
	return ev.varMap[name]
}

func (ev *EnvironmentVariableMap) Err() error {
	return ev.err
}

func CheckPasswordStrength(passwd string) bool {
	noUpper, noDigit, noSpecial := true, true, true
	for _, c := range passwd {
		if isUpper(c) {
			noUpper = false
		}
		if isDigit(c) {
			noDigit = false
		}
		if !(isUpper(c) || isLower(c) || isDigit(c)) {
			noSpecial = false
		}
	}

	return !(noUpper || noDigit || noSpecial || len(passwd) < 8)
}

func CheckEmailFormat(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func isUpper(c rune) bool {
	return c >= 'A' && c <= 'Z'
}

func isLower(c rune) bool {
	return c >= 'a' && c <= 'z'
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}
