-- name: CreateFeedFollow :one
with inserted as (INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5
)
RETURNING *)

SELECT
	inserted.*,
	feeds.name feed_name,
	users.name user_name
FROM
	inserted
	JOIN feeds ON inserted.feed_id = feeds.id
	JOIN users ON inserted.user_id = users.id;
	
-- name: GetFeedFollowsForUser :many
SELECT
	feed_follows.*,
	feeds.name feed_name
FROM
	feed_follows
	JOIN feeds on feed_follows.feed_id = feeds.id
WHERE
	feed_follows.user_id = $1;

-- name: DeleteFeedFollow :one
DELETE FROM feed_follows
WHERE
	feed_follows.user_id = $1
	and feed_follows.feed_id = $2
RETURNING *;
