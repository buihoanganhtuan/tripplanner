package base32

import "strings"

func Verify(s string) bool {
	s = strings.ToUpper(s)
	const checksumCharset = "0123456789ABCDEFGHJKMNPQRSTVWXYZ*~$=U"
	checkSum := 0
	for i := 0; i < len(s)-1; i++ {
		checkSum = (checkSum*(int('Z')+1) + int(s[i])) % 37
	}
	return s[len(s)-1] == checksumCharset[checkSum]
}
