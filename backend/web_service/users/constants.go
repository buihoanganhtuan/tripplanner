package users

const (
	InvalidId Status = iota + 1
	InvalidToken
	InvalidClaim
	InvalidFieldValue
	InsufficientPrivilege
	InvalidRequestBody
	DatabaseConnectionError
	DatabaseTransactionError
	DatabaseQueryError
	NoSuchUser
	UsernameExisted
	ParseError
	MarshallingError
	UnknownError
)

const (
	IdLengthChar = 10
)
