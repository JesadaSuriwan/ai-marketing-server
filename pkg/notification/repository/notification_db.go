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

func (r notificationRepositoryDB) GetAllForUser(userId int) ([]Notification, error) {
	list := []Notification{}
	query := `
		SELECT n.id, n.user_id, n.company_id, n.type, n.title, n.description, n.is_read, n.created_at,
			c.name AS company_name, c.website AS company_website
		FROM notifications n
		JOIN companies c ON c.id = n.company_id
		WHERE c.user_id = $1 OR EXISTS(SELECT 1 FROM company_members cm WHERE cm.company_id = c.id AND cm.user_id = $1 AND cm.status = 'active')
		ORDER BY n.created_at DESC
	`
	err := r.db.Select(&list, query, userId)
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

func (r notificationRepositoryDB) MarkAllReadForUser(tx *sqlx.Tx, userId int) error {
	_, err := tx.Exec(`
		UPDATE notifications n SET is_read = true
		FROM companies c
		WHERE n.company_id = c.id AND (
			c.user_id = $1
			OR EXISTS(SELECT 1 FROM company_members cm WHERE cm.company_id = c.id AND cm.user_id = $1 AND cm.status = 'active')
		)
	`, userId)
	return err
}

func (r notificationRepositoryDB) Delete(tx *sqlx.Tx, id int) error {
	_, err := tx.Exec(`DELETE FROM notifications WHERE id=$1`, id)
	return err
}

func (r notificationRepositoryDB) ExistsByCompanyTypeAndDescription(companyId int, nType, description string) (bool, error) {
	var exists bool
	err := r.db.QueryRowx(
		`SELECT EXISTS(SELECT 1 FROM notifications WHERE company_id = $1 AND type = $2 AND description = $3)`,
		companyId, nType, description,
	).Scan(&exists)
	return exists, err
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
