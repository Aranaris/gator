-- +goose Up
CREATE TABLE feed_follows (
	id bigserial primary key,
	created_at timestamp not null,
	updated_at timestamp not null,
	user_id bigserial not null,
	feed_id bigserial not null,
	CONSTRAINT fk_user_feed_follows
		FOREIGN KEY(user_id)
		REFERENCES users(id)
		ON DELETE CASCADE,
	CONSTRAINT fk_feeds_feed_follows
		FOREIGN KEY(feed_id)
		REFERENCES feeds(id)
		ON DELETE CASCADE,
	unique (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;
