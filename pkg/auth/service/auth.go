package service

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Initials string `json:"initials"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserData struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Initials string `json:"initials"`
}

type AuthResponse struct {
	Status bool     `json:"status"`
	Desc   string   `json:"desc"`
	Data   UserData `json:"data"`
	Token  string   `json:"token,omitempty"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type AuthService interface {
	Register(req RegisterRequest) (*AuthResponse, string, error)
	Login(req LoginRequest) (*AuthResponse, string, error)
	Me(userId int) (*AuthResponse, error)
}
