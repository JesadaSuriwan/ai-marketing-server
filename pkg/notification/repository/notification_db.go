package repository

import "github.com/jmoiron/sqlx"

type notificationRepositoryDB struct {
	db *sqlx.DB
}

func NewNotificationRepositoryDB(db *sqlx.DB) NotificationRepository {
	return notificationRepositoryDB{db}
}

func (r notificationRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r notificationRepositoryDB) GetAll(companyId int) ([]Notification, error) {
	list := []Notification{}
	err := r.db.Select(&list, `SELECT id, user_id, company_id, type, title, description, is_read, created_at FROM notifications WHERE company_id = $1 ORDER BY created_at DESC`, companyId)
	return list, err
}

func (r notificationRepositoryDB) Create(tx *sqlx.Tx, n Notification) (int, error) {
	var id int
	err := tx.QueryRowx(`INSERT INTO notifications (user_id, company_id, type, title, description) VALUES ($1,$2,$3,$4,$5) RETURNING id`, n.UserId, n.CompanyId, n.Type, n.Title, n.Description).Scan(&id)
	return id, err
}

func (r notificationRepositoryDB) MarkRead(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`UPDATE notifications SET is_read=true WHERE id=$1`, id)
	return err
}

func (r notificationRepositoryDB) MarkAllRead(tx *sqlx.Tx, companyId int) error {
	_, err := tx.Exec(`UPDATE notifications SET is_read=true WHERE company_id=$1`, companyId)
	return err
}

func (r notificationRepositoryDB) Delete(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`DELETE FROM notifications WHERE id=$1`, id)
	return err
}

func (r notificationRepositoryDB) BelongsToUser(id, userId int) (bool, error) {
	var exists bool
	err := r.db.QueryRowx(
		`SELECT EXISTS(SELECT 1 FROM notifications n JOIN companies c ON n.company_id = c.id WHERE n.id = $1 AND (
			c.user_id = $2
			OR EXISTS(SELECT 1 FROM company_members cm WHERE cm.company_id = c.id AND cm.user_id = $2 AND cm.status = 'active')
		))`,
		id, userId,
	).Scan(&exists)
	return exists, err
}
