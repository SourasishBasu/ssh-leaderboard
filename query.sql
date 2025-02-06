-- name: ListMembers :many
SELECT 
    RANK() OVER (ORDER BY (COALESCE(game1, 0) + COALESCE(game2, 0) + COALESCE(game3, 0)) DESC) AS rank,
    name, 
    COALESCE(game1, 0) + COALESCE(game2, 0) + COALESCE(game3, 0) AS scores
FROM participants;
