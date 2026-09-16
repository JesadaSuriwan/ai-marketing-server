// Package permission computes a user's effective role for a company and
// checks it against an allow-list. "Admin" is never stored in the database —
// it's always derived from companies.user_id (the creator) — while
// team_lead/specialist/customer live in company_members.role.
package permission

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

const (
	Admin      = "admin"
	TeamLead   = "team_lead"
	Specialist = "specialist"
	Customer   = "customer"
)

// EffectiveRole returns "admin" if userId created the company, otherwise
// their company_members.role if they have an active membership, otherwise ""
// (no relationship to this company at all).
func EffectiveRole(db *sqlx.DB, companyId, userId int) (string, error) {
	var isOwner bool
	if err := db.Get(&isOwner, `SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1 AND user_id = $2)`, companyId, userId); err != nil {
		return "", err
	}
	if isOwner {
		return Admin, nil
	}

	var role string
	err := db.Get(&role, `SELECT role FROM company_members WHERE company_id = $1 AND user_id = $2 AND status = 'active'`, companyId, userId)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return role, nil
}

// Allowed reports whether role matches one of the given allowed roles.
func Allowed(role string, allowed ...string) bool {
	for _, a := range allowed {
		if role == a {
			return true
		}
	}
	return false
}

// CanCreateWorkspace reports whether userId may create a brand new company:
// true if they already own at least one company (admin somewhere), true if
// they're team_lead on at least one active membership, or true for the
// bootstrap case (a user with zero relationship to any company yet). A user
// who only holds specialist/customer memberships is not allowed.
func CanCreateWorkspace(db *sqlx.DB, userId int) (bool, error) {
	var ownsAny bool
	if err := db.Get(&ownsAny, `SELECT EXISTS(SELECT 1 FROM companies WHERE user_id = $1)`, userId); err != nil {
		return false, err
	}
	if ownsAny {
		return true, nil
	}

	var isLeadSomewhere bool
	if err := db.Get(&isLeadSomewhere, `SELECT EXISTS(SELECT 1 FROM company_members WHERE user_id = $1 AND status = 'active' AND role = $2)`, userId, TeamLead); err != nil {
		return false, err
	}
	if isLeadSomewhere {
		return true, nil
	}

	var hasAnyMembership bool
	if err := db.Get(&hasAnyMembership, `SELECT EXISTS(SELECT 1 FROM company_members WHERE user_id = $1 AND status = 'active')`, userId); err != nil {
		return false, err
	}
	return !hasAnyMembership, nil
}
