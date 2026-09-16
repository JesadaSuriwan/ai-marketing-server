package service

type NotificationData struct {
	Id          int    `json:"id"`
	UserId      int    `json:"user_id"`
	CompanyId   int    `json:"company_id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsRead      bool   `json:"is_read"`
	CreatedAt   string `json:"created_at"`
}

type CreateNotificationRequest struct {
	UserId      int    `json:"user_id,string" binding:"required"`
	CompanyId   int    `json:"company_id,string" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type NotificationListResponse struct {
	Status bool               `json:"status"`
	Desc   string             `json:"desc"`
	Data   []NotificationData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type NotificationService interface {
	GetAll(companyId int) (*NotificationListResponse, error)
	Create(req CreateNotificationRequest) (*SimpleResponse, error)
	MarkRead(id, userId int) (*SimpleResponse, error)
	MarkAllRead(companyId int) (*SimpleResponse, error)
	Delete(id, userId int) (*SimpleResponse, error)
}
