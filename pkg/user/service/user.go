package service

type UserData struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Initials string `json:"initials"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"`
	Initials string `json:"initials"`
}

type UserResponse struct {
	Status bool     `json:"status"`
	Desc   string   `json:"desc"`
	Data   UserData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type UserService interface {
	GetById(id int) (*UserResponse, error)
	Update(id int, req UpdateUserRequest) (*SimpleResponse, error)
}
