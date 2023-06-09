package utils

import (
	"fmt"
	"net/mail"
	"os"
)

// unsafe manager type to manage environment variables
// Attempting to get a variable via Var() when the manager is in error state or
// without fetching the variable first via Fetch() will result in a panic
type EnvironmentVariableMap struct {
	varMap map[string]string
	err    error
}

func (env *EnvironmentVariableMap) Fetch(names ...string) {
	if env.err != nil {
		return
	}
	for _, name := range names {
		if _, exist := env.varMap[name]; exist {
			continue
		}
		if val, ok := os.LookupEnv(name); ok {
			env.varMap[name] = val
			continue
		}
		env.err = fmt.Errorf("environment variable %v is unset", name)
	}
}

func (env *EnvironmentVariableMap) Var(name string) string {
	if _, exist := env.varMap[name]; env.err != nil || !exist {
		if env.err != nil {
			panic(env.err)
		}
		panic(fmt.Errorf("environmental variable %v has not been fetched", name))
	}
	return env.varMap[name]
}

func (env *EnvironmentVariableMap) Err() error {
	return env.err
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

func CheckUsername(username string) bool {
	for _, c := range username {
		if !isUpper(c) && !isLower(c) && !isDigit(c) {
			return false
		}
	}
	return len(username) > 0 && len(username) <= 30
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
