-- +goose Up
CREATE TABLE posts (
	id bigserial primary key,
	created_at timestamp not null,
	updated_at timestamp not null,
	title text not null,
	url text not null unique,
	description text,
	published_at timestamp not null,
	feed_id bigserial not null,
	CONSTRAINT fk_posts_feeds
		FOREIGN KEY(feed_id)
		REFERENCES feeds(id)
		ON DELETE CASCADE
	);

-- +goose Down
DROP TABLE posts;
