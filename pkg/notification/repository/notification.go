package repository

import "github.com/jmoiron/sqlx"

type Notification struct {
	Id          int    `db:"id"`
	UserId      int    `db:"user_id"`
	CompanyId   int    `db:"company_id"`
	Type        string `db:"type"`
	Title       string `db:"title"`
	Description string `db:"description"`
	IsRead      bool   `db:"is_read"`
	CreatedAt   string `db:"created_at"`
}

type NotificationRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	GetAll(companyId int) ([]Notification, error)
	Create(tx *sqlx.Tx, n Notification) (int, error)
	MarkRead(tx *sqlx.Tx, id int) error
	MarkAllRead(tx *sqlx.Tx, companyId int) error
	Delete(tx *sqlx.Tx, id int) error
	BelongsToUser(id, userId int) (bool, error)
}
