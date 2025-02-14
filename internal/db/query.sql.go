// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package db

import (
	"context"
)

const listMembers = `-- name: ListMembers :one
SELECT RANK() OVER (ORDER BY MAX(cq.completed_at) DESC) as rank, 
  t.team_name as name, tp.total_points as scores
FROM teams t
JOIN team_points tp ON t.team_id = tp.team_id
LEFT JOIN completed_questions cq ON t.team_id = cq.team_id
GROUP BY t.team_name, tp.total_points
ORDER BY rank
`

type ListMembersRow struct {
	Rank   int64
	Name   string
	Scores int32
}

func (q *Queries) ListMembers(ctx context.Context) ([]ListMembersRow, error) {
	rows, err := q.db.Query(ctx, listMembers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListMembersRow
	for rows.Next() {
		var i ListMembersRow
		if err := rows.Scan(&i.Rank, &i.Name, &i.Scores); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
