## RSS Feed Aggregator

This is a commandline tool for aggregating RSS feeds.

Go and Postgres are required for this program to run.

### Installation

After cloning this repo, set up a new db connection in Postgres for this tool and save the connection string.

Add a .env file to the root of the repository with the `DB_URL` variable. The connection string should look something like `postgres://<username>:@<connection>/<db_name>`

Install the CLI via `go install`.

### Configuration

This tool references a config file to keep track of the logged in user. In your `$HOME` directory, add a `.gatorconfig.json` file with the following structure:

```
{
	"db_url":<CONNECTION_STRING_GOES_HERE>?sslmode=disable,
	"current_username":""
}
```

### Using the tool

Once installed and configured, you can run the gator cli and various commands.

`register <username>` will add a new logged in user.

`addfeed <feed_name> <feed_url>` save a new RSS feed to the db.

`agg <time_interval>` starts a ticker that will continuously retrieve new posts from a user's followed feeds after every time interval

`follow <feed_url>` adds a feed to a user's follow list

`browse <limit(2)>` shows the X most recent posts for the logged in user's feeds (default 2)
