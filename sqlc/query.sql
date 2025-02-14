-- name: ListMembers :many
SELECT RANK() OVER (ORDER BY MAX(cq.completed_at) DESC) as rank, t.team_name as name, tp.total_points as scores
FROM teams t
JOIN team_points tp ON t.team_id = tp.team_id
LEFT JOIN completed_questions cq ON t.team_id = cq.team_id
GROUP BY t.team_name, tp.total_points
ORDER BY rank