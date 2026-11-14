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
	// CompanyName/CompanyWebsite are only populated by GetAllForUser's join
	// (the inbox is shared across every company the user can access, so each
	// row needs to say which one it's about) — zero-valued everywhere else.
	CompanyName    string `db:"company_name"`
	CompanyWebsite string `db:"company_website"`
}

type NotificationRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	// GetAllForUser returns every notification across every company userId
	// owns or is an accepted member of — the Notification Center and the
	// contract-reminder push dialog are both a single shared inbox, not
	// scoped to whichever company happens to be active.
	GetAllForUser(userId int) ([]Notification, error)
	Create(tx *sqlx.Tx, n Notification) (int, error)
	MarkRead(tx *sqlx.Tx, id int) error
	// MarkAllReadForUser marks read every notification across every company
	// userId can access (mirrors GetAllForUser's scope).
	MarkAllReadForUser(tx *sqlx.Tx, userId int) error
	Delete(tx *sqlx.Tx, id int) error
	BelongsToUser(id, userId int) (bool, error)
	// ExistsByCompanyTypeAndDescription dedupes recurring system-generated
	// notifications (e.g. the contract-expiry reminder) — description
	// embeds the exact date it's about, so a new one is only created once
	// that date actually changes (e.g. after a renewal).
	ExistsByCompanyTypeAndDescription(companyId int, nType, description string) (bool, error)
}
