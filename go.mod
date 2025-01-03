module github.com/aranaris/gator

go 1.22.5

require internal/config v1.0.0
require internal/rss v1.0.0

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
)

replace internal/config => ./internal/config
replace internal/rss => ./internal/rss
