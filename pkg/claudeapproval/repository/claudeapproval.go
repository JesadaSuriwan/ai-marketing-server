package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type ClaudeApprovalRequest struct {
	Id                int        `db:"id"`
	CompanyId         int        `db:"company_id"`
	RequestedByUserId int        `db:"requested_by_user_id"`
	Status            string     `db:"status"`
	Token             string     `db:"token"`
	TokenExpiresAt    time.Time  `db:"token_expires_at"`
	ResolvedAt        *time.Time `db:"resolved_at"`
	CreatedAt         string     `db:"created_at"`
}

type ClaudeApprovalRepository interface {
	NewTransaction() (*sqlx.Tx, error)
	Create(tx *sqlx.Tx, r ClaudeApprovalRequest) (int, error)
	GetByToken(token string) (*ClaudeApprovalRequest, error)
	MarkResolved(tx *sqlx.Tx, id int, status string) error
}
