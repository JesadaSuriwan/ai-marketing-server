package repository

import "github.com/jmoiron/sqlx"

type claudeApprovalRepositoryDB struct {
	db *sqlx.DB
}

func NewClaudeApprovalRepositoryDB(db *sqlx.DB) ClaudeApprovalRepository {
	return claudeApprovalRepositoryDB{db}
}

func (r claudeApprovalRepositoryDB) NewTransaction() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r claudeApprovalRepositoryDB) Create(tx *sqlx.Tx, req ClaudeApprovalRequest) (int, error) {
	var id int
	query := `INSERT INTO claude_approval_requests (company_id, requested_by_user_id, status, token, token_expires_at)
		VALUES ($1,$2,$3,$4,$5) RETURNING id`
	err := tx.QueryRowx(query, req.CompanyId, req.RequestedByUserId, req.Status, req.Token, req.TokenExpiresAt).Scan(&id)
	return id, err
}

func (r claudeApprovalRepositoryDB) GetByToken(token string) (*ClaudeApprovalRequest, error) {
	req := ClaudeApprovalRequest{}
	query := `SELECT id, company_id, requested_by_user_id, status, token, token_expires_at, resolved_at, created_at
		FROM claude_approval_requests WHERE token = $1`
	err := r.db.Get(&req, query, token)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r claudeApprovalRepositoryDB) MarkResolved(tx *sqlx.Tx, id int, status string) error {
	_, err := tx.Exec(`UPDATE claude_approval_requests SET status=$1, resolved_at=NOW() WHERE id=$2`, status, id)
	return err
}
