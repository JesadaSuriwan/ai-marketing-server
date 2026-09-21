package repository

import (
	"strconv"

	"github.com/jmoiron/sqlx"
)

type usageRepositoryDB struct {
	db *sqlx.DB
}

func NewUsageRepositoryDB(db *sqlx.DB) UsageRepository {
	return usageRepositoryDB{db}
}

func (r usageRepositoryDB) Log(l UsageLog) error {
	_, err := r.db.Exec(
		`INSERT INTO api_usage_log (company_id, engine, purpose, model, input_tokens, output_tokens, cost_usd)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		l.CompanyId, l.Engine, l.Purpose, l.Model, l.InputTokens, l.OutputTokens, l.CostUsd,
	)
	return err
}

// GetBreakdownForCompany sums usage for one company. Access is checked by the
// route's company middleware, not here.
func (r usageRepositoryDB) GetBreakdownForCompany(companyId int, from, to string) ([]EngineBreakdown, error) {
	list := []EngineBreakdown{}
	query := `
		SELECT
			u.engine,
			u.purpose,
			COUNT(*)::INT AS calls,
			COALESCE(SUM(u.input_tokens), 0)::INT AS input_tokens,
			COALESCE(SUM(u.output_tokens), 0)::INT AS output_tokens,
			COALESCE(SUM(u.cost_usd), 0)::FLOAT AS cost_usd
		FROM api_usage_log u
		WHERE u.company_id = $1
	`
	args := []interface{}{companyId}
	if from != "" {
		args = append(args, from)
		query += ` AND u.created_at::date >= $` + strconv.Itoa(len(args))
	}
	if to != "" {
		args = append(args, to)
		query += ` AND u.created_at::date <= $` + strconv.Itoa(len(args))
	}
	query += ` GROUP BY u.engine, u.purpose ORDER BY cost_usd DESC`
	err := r.db.Select(&list, query, args...)
	return list, err
}
