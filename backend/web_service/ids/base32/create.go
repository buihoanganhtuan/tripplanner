package base32

import (
	"math/rand"
	"time"
)

func Create(length int) string {
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
