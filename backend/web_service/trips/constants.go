package trips

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
	NoSuchTrip
	ParseError
	MarshallingError
	UnknownError
)

const (
	InvalidIdMessage                = "invalid user id"
	InvalidTokenMessge              = "invalid access token"
	InvalidClaimMessage             = "invalid claim %s"
	InvalidFieldValueMessage        = "invalid value for field %s"
	InsufficientPrivilegeMessage    = "user does not have privilege for action attempted on resource"
	InvalidRequestBodyMessage       = "invalid http request body"
	DatabaseConnectionErrorMessage  = "fail to connect to database"
	DatabaseTransactionErrorMessage = "fail to execute the action"
	DatabaseQueryErrorMessage       = "fail to query the database"
	NoSuchTripMessage               = "trip does not exist"
	ParseErrorMessage               = "fail to parse field %s"
	MarshallingErrorMessage         = "fail to construct http response body"
	UnknownErrorMessage             = "an unexpected error occurred"
)

const (
	IdLengthChar = 10
)
