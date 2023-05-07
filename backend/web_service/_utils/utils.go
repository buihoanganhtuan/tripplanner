package utils

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

// unsafe manager type to manage environment variables
// Attempting to get a variable via Var() when the manager is in error state or
// without fetching the variable first via Fetch() will result in a panic
type EnvironmentVariableMap struct {
	varMap map[string]string
	err    error
}

type Set[T comparable] struct {
	m  map[T]bool
	sz int
}

type Queue[T any] struct {
	popStack  []T
	pushStack []T
	sz        int
}

func (q *Queue[T]) Push(val T) {
	q.pushStack = append(q.pushStack, val)
	q.sz++
}

func (q *Queue[T]) Pop() (T, bool) {
	ret, ok := q.Peek()
	if !ok {
		return ret, ok
	}
	q.popStack = q.popStack[:len(q.popStack)-1]
	q.sz--
	return ret, ok
}

func (q *Queue[T]) Peek() (T, bool) {
	var ret T
	if q.sz == 0 {
		return ret, false
	}
	if len(q.popStack) == 0 {
		for i := len(q.pushStack) - 1; i > -1; i-- {
			q.popStack = append(q.popStack, q.pushStack[i])
		}
		q.pushStack = []T{}
	}
	ret = q.popStack[len(q.popStack)-1]
	return ret, true
}

func (q *Queue[T]) IsEmpty() bool {
	return q.sz == 0
}

func (s *Set[T]) Add(val T) bool {
	_, ok := s.m[val]
	if ok {
		return false
	}
	s.m[val] = true
	s.sz++
	return true
}

func (s *Set[T]) AddAll(vals ...T) []bool {
	res := make([]bool, len(vals))
	for i, v := range vals {
		res[i] = s.Add(v)
	}
	return res
}

func (s *Set[T]) Remove(val T) bool {
	_, ok := s.m[val]
	if !ok {
		return false
	}
	delete(s.m, val)
	s.sz--
	return true
}

func (s *Set[T]) Contains(val T) bool {
	_, ok := s.m[val]
	return ok
}

func (s *Set[T]) Empty() bool {
	return s.sz == 0
}

func (s *Set[T]) Size() int {
	return s.sz
}

func (s *Set[T]) Values() []T {
	var res []T
	for k := range s.m {
		res = append(res, k)
	}
	return res
}

func (s *Set[T]) ToString(fmtFn func(T) string, sep string) string {
	tmp := make([]string, s.sz)
	var i int
	for k := range s.m {
		tmp[i] = fmtFn(k)
	}
	return strings.Join(tmp, sep)
}

func NewSet[T comparable](vals ...T) Set[T] {
	var s Set[T]
	for _, v := range vals {
		s.Add(v)
	}
	return s
}

func (env *EnvironmentVariableMap) Fetch(names ...string) {
	if env.err != nil {
		return
	}
	if len(env.varMap) == 0 {
		env.varMap = make(map[string]string)
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

func GetBase32RandomString(length int) string {
	const b32Charset = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	const checksumCharset = "0123456789ABCDEFGHJKMNPQRSTVWXYZ*~$=U"
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	b := make([]byte, length+1)
	checkSum := 0
	for i := 0; i < length; i++ {
		b[i] = b32Charset[rnd.Intn(32)]
		checkSum = (checkSum*(int('Z')+1) + int(b[i])) % 37
	}
	b[length] = checksumCharset[checkSum]
	return string(b)
}

func VerifyBase32String(s string) bool {
	s = strings.ToUpper(s)
	const checksumCharset = "0123456789ABCDEFGHJKMNPQRSTVWXYZ*~$=U"
	checkSum := 0
	for i := 0; i < len(s)-1; i++ {
		checkSum = (checkSum*(int('Z')+1) + int(s[i])) % 37
	}
	return s[len(s)-1] == checksumCharset[checkSum]
}

func (env *EnvironmentVariableMap) Err() error {
	return env.err
}

func ExtractClaims(rq *http.Request, pk *rsa.PublicKey) (jwt.MapClaims, error) {
	// check if there is any access token
	if rq.Header.Get("Authorization") == "" {
		return nil, fmt.Errorf("no access token")
	}

	// check access token integrity. Note that we don't support BasicAuth
	ts, ok := strings.CutPrefix(rq.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return nil, errors.New("invalid authorization header")
	}

	token, err := jwt.Parse(ts, func(token *jwt.Token) (interface{}, error) {
		return pk, nil
	}, jwt.WithValidMethods([]string{"RSA"}))

	if err != nil || !token.Valid {
		return nil, err
	}

	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claim datatype")
	}

	return mc, nil
}

func GetMapKeys[K comparable, V any](m map[K]V) []K {
	var ret []K
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}

func GetMapValues[K comparable, V any](m map[K]V) []V {
	var ret []V
	for _, v := range m {
		ret = append(ret, v)
	}
	return ret
}
