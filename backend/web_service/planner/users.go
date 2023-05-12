package planner

type User struct {
	Id       UserId `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserService interface {
	GetUser(id UserId) (User, error)
	CreateUser(u User) (User, error)
	ListUsers() ([]User, error)
	UpdateUser(u User) (User, error)
	ReplaceUser(u User) (User, error)
	DeleteUser(id UserId) error
}

type UserId string
