-- +goose Up
CREATE TABLE feeds (
	id bigserial primary key,
	created_at timestamp not null,
	updated_at timestamp not null,
	name text not null,
	url text not null unique,
	user_id bigserial not null,
	CONSTRAINT fk_users_feeds
		FOREIGN KEY(user_id)
		REFERENCES users(id)
		ON DELETE CASCADE
	);

-- +goose Down
DROP TABLE feeds;
