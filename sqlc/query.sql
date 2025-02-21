-- name: ListMembers :many
SELECT 
    ROW_NUMBER() OVER (ORDER BY tp.total_points DESC) as rank,
    t.team_name as name,
    tp.total_points as scores
FROM 
    teams t
JOIN 
    team_points tp ON t.team_id = tp.team_id
ORDER BY 
    tp.total_points DESC