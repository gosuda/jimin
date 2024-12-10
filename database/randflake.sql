-- name: TryCreateRandflakeRangeLease :one
INSERT INTO randflake_nodes (range_start, range_end, lease_holder, lease_start, lease_end)
    SELECT $1, $2, $3, $4, $5
        WHERE NOT EXISTS (
            SELECT * FROM randflake_nodes
                WHERE
                    (range_start <= $4 AND range_end >= $3) -- overlap
                    AND randflake_nodes.lease_end >= $6 -- not expired
        )
RETURNING *;

-- name: GetRandflakeRangeLease :many
SELECT range_start, range_end FROM randflake_nodes
    WHERE
        range_start >= $1
        AND lease_end >= $2
    ORDER BY range_start ASC LIMIT $3;

-- name: RenewRandflakeRangeLease :one
UPDATE randflake_nodes SET lease_end = $1 WHERE id = $2 AND lease_holder = $3 AND lease_end >= $4 RETURNING *;

-- name: ReleaseRandflakeRangeLease :exec
DELETE FROM randflake_nodes WHERE id = $1 AND lease_holder = $2;

-- name: GCRandflakeLeases :exec
DELETE FROM randflake_nodes WHERE lease_end < $1;