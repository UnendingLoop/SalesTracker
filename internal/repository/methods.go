package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/UnendingLoop/SalesTracker/internal/model"
	"github.com/wb-go/wbf/dbpg"
)

type PostgresRepo struct {
	db *dbpg.DB
}

func (pr *PostgresRepo) Create(ctx context.Context, op *model.Operation) error {
	query := `INSERT INTO operations (amount, actor_id, category_id, type, operation_at, description)
	VALUES (
    $1,
    (SELECT id FROM family_members WHERE fam_member = $2),
    (SELECT id FROM category WHERE cat_name = $3),
    $4,
    $5,
    $6);`

	_, err := pr.db.ExecContext(ctx, query, op.Amount, op.Actor, op.Category, op.Type, op.OperationAt, op.Description)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "null value in column"):
			return model.ErrUnknownActorOrCategory
		default:
			return err
		}
	}

	return nil
}

func (pr *PostgresRepo) Get(ctx context.Context, id int) (*model.Operation, error) {
	query := `SELECT o.id, o.amount, f.fam_member, c.cat_name, o.type, o.operation_at, o.created_at, o.description 
	FROM operations o 
	LEFT JOIN category c ON c.id = o.category_id 
	LEFT JOIN family_members f ON f.id = o.actor_id 
	WHERE o.id = $1`

	var result model.Operation
	if err := pr.db.QueryRowContext(ctx, query, id).Scan(
		&result.ID,
		&result.Amount,
		&result.Actor,
		&result.Category,
		&result.Type,
		&result.OperationAt,
		&result.CreatedAt,
		&result.Description); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, model.ErrOperationIDNotFound
		default:
			return nil, err
		}
	}

	return &result, nil
}

func (pr *PostgresRepo) List(ctx context.Context, f *model.RequestParamOperations) ([]model.Operation, error) {
	periodExpr := definePeriodExpr(f.StartTime, f.EndTime)
	limofExpr := defineLimitOffsetExpr(f.Limit, f.Page)
	orderExpr, err := defineOrderExpr(f.OrderBy, f.ASC, f.DESC)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`SELECT o.id, o.amount, f.fam_member, c.cat_name, o.type, o.operation_at, o.created_at, o.description 
	FROM operations o 
	LEFT JOIN category c ON c.id = o.category_id 
	LEFT JOIN family_members f ON f.id = o.actor_id
	%s
	%s
	%s`, periodExpr, orderExpr, limofExpr)

	rows, err := pr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.Operation, 0)
	for rows.Next() {
		item := model.Operation{}
		if err := rows.Scan(
			&item.ID,
			&item.Amount,
			&item.Actor,
			&item.Category,
			&item.Type,
			&item.OperationAt,
			&item.CreatedAt,
			&item.Description); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return result, nil
}

func (pr *PostgresRepo) Update(ctx context.Context, op *model.Operation) error {
	query := `UPDATE operations SET 
	amount = $2, 
	actor_id = (SELECT id FROM family_members WHERE fam_member = $3), 
	category_id = (SELECT id FROM category WHERE cat_name = $4), 
	type = $5, 
	operation_at = $6, 
	description = $7 
	WHERE id = $1;`

	row, err := pr.db.ExecContext(ctx,
		query,
		op.ID,
		op.Amount,
		op.Actor,
		op.Category,
		op.Type,
		op.OperationAt,
		op.Description)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "null value"):
			return model.ErrUnknownActorOrCategory
		default:
			return err
		}
	}

	rows, _ := row.RowsAffected()
	if rows == 0 {
		return model.ErrOperationIDNotFound
	}

	return nil
}

func (pr *PostgresRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM operations WHERE id = $1`

	row, err := pr.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, _ := row.RowsAffected()
	if rows == 0 {
		return model.ErrOperationIDNotFound
	}

	return nil
}

func (pr *PostgresRepo) AnalyticsGroup(ctx context.Context, f *model.RequestParamAnalytics) ([]model.AnalyticsQuantum, error) {
	groupExpr, err := defineGroupExpr(f.GroupBy)
	if err != nil {
		return nil, err
	}
	limitOffsetExpr := defineLimitOffsetExpr(f.Limit, f.Page)
	periodExpr := definePeriodExpr(f.StartTime, f.EndTime)

	query := fmt.Sprintf(`SELECT %s AS group_key,
       SUM(amount)::float8/100,
       AVG(amount)::float8/100,
       COUNT(*),
	   percentile_cont(0.5) WITHIN GROUP (ORDER BY amount)::float8 / 100,
	   percentile_cont(0.9) WITHIN GROUP (ORDER BY amount)::float8 / 100
	   FROM operations o 
	   LEFT JOIN category c ON c.id = o.category_id 
	   LEFT JOIN family_members f ON f.id = o.actor_id
	   %s
	   GROUP BY %s
	   ORDER BY %s
	   %s`, groupExpr, periodExpr, groupExpr, groupExpr, limitOffsetExpr)

	rows, err := pr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.AnalyticsQuantum, 0)
	for rows.Next() {
		var item model.AnalyticsQuantum
		if err := rows.Scan(&item.Key, &item.Sum, &item.Avg, &item.Count, &item.Median, &item.P90); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func (pr *PostgresRepo) AnalyticsSummary(ctx context.Context, f *model.RequestParamAnalytics) (*model.AnalyticsSummary, error) {
	periodExpr := definePeriodExpr(f.StartTime, f.EndTime)
	query := fmt.Sprintf(`SELECT
       SUM(amount)::float8/100,
       AVG(amount)::float8/100,
       COUNT(*),
	   percentile_cont(0.5) WITHIN GROUP (ORDER BY amount)::float8 / 100,
	   percentile_cont(0.9) WITHIN GROUP (ORDER BY amount)::float8 / 100
	   FROM operations
	   %s`, periodExpr)

	var result model.AnalyticsSummary
	if err := pr.db.QueryRowContext(ctx, query).Scan(
		&result.Sum,
		&result.Avg,
		&result.Count,
		&result.Median,
		&result.P90); err != nil {
		return nil, err
	}

	return &result, nil
}

func defineGroupExpr(input *string) (string, error) {
	if input == nil {
		return "", model.ErrInvalidGroupBy
	}

	switch *input {
	case model.GroupByDay:
		return "date_trunc('day', operation_at)", nil
	case model.GroupByWeek:
		return "date_trunc('week', operation_at)", nil
	case model.GroupByMonth:
		return "date_trunc('month', operation_at)", nil
	case model.GroupByYear:
		return "date_trunc('year', operation_at)", nil
	case model.GroupByActor:
		return "f.fam_member", nil
	case model.GroupByCategory:
		return "c.cat_name", nil
	case model.GroupByOpType:
		return "o.type", nil
	default:
		return "", model.ErrInvalidGroupBy
	}
}

func defineLimitOffsetExpr(lim, p *int) string {
	if lim == nil && p == nil { // оба значения нил - вообще не применяем их к квери
		return ""
	}

	// избавляемся от указателей
	var limit, page int
	if lim != nil {
		limit = *lim
	}
	if p != nil {
		page = *p
	}

	if limit <= 0 { // задаем значение по умолчанию если лимит пуст/некорректен
		limit = 20
	}

	if limit > 1000 { // защита от слишком больших значений
		limit = 1000
	}

	if page <= 0 { // если страница имеет некорректное значение - ставим 1
		page = 1
	}

	// оба значения корректны - добавляем в квери
	offset := limit * (page - 1)
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

func defineOrderExpr(orderBy *string, asc, desc bool) (string, error) {
	if orderBy == nil {
		return "", nil
	}

	direction := ""
	switch {
	case asc == desc:
		direction = "DESC" // значение по умолчанию
	case asc:
		direction = "ASC"
	default:
		direction = "DESC"
	}

	switch *orderBy {
	case model.OrderByActor:
		return fmt.Sprintf("ORDER BY f.fam_member %s ", direction), nil
	case model.OrderByAmount:
		return fmt.Sprintf("ORDER BY o.amount %s ", direction), nil
	case model.OrderByCategory:
		return fmt.Sprintf("ORDER BY c.cat_name %s ", direction), nil
	case model.OrderByOpDate:
		return fmt.Sprintf("ORDER BY o.operation_at %s ", direction), nil
	case model.OrderByOpID:
		return fmt.Sprintf("ORDER BY o.id %s ", direction), nil
	case model.OrderByType:
		return fmt.Sprintf("ORDER BY o.type %s ", direction), nil
	default:
		return "", model.ErrInvalidOrderBy
	}
}

func definePeriodExpr(start, end *time.Time) string {
	switch {
	case start != nil && end != nil:
		return fmt.Sprintf("WHERE operation_at BETWEEN '%s' AND '%s'", start.Format(time.RFC3339), end.Format(time.RFC3339))
	case start != nil:
		return fmt.Sprintf("WHERE operation_at > '%s'", start.Format(time.RFC3339))
	case end != nil:
		return fmt.Sprintf("WHERE operation_at < '%s'", end.Format(time.RFC3339))
	default:
		return ""
	}
}
