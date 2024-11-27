-- +goose Up
CREATE TABLE users (
	id bigserial primary key,
	created_at timestamp not null,
	updated_at timestamp not null,
	name text not null unique
	);

-- +goose Down
DROP TABLE users;
