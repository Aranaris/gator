package main

import (
	"context"
	"database/sql"
	"fmt"
	"internal/config"
	"internal/rss"
	"os"
	"time"

	"github.com/aranaris/gator/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	db *database.Queries
}

type command struct {
	name string
	arguments []string
}

type commands struct {
	mapping map[string]func(*state, command) error
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	f := func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUser)
		if err != nil {
			return err
		}
		
		err = handler(s, cmd, user)
		if err != nil {
			return err
		}
		return nil
	}
	
	return f
}

func (c *commands) register(name string, f func(*state, command) error){
	c.mapping[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	err := c.mapping[cmd.name](s, cmd)
	if err != nil {
		return err
	}

	return nil
}

func loginHandler(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("username required")
	}

	_, err := s.db.GetUser(context.Background(),cmd.arguments[0])
	if err != nil {
		os.Exit(1)
	}

	err = s.cfg.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}

	fmt.Printf("User %s has been set.\n", cmd.arguments[0])

	return nil
}

func registerHandler(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return fmt.Errorf("name required")
	}

	_, err := s.db.GetUser(context.Background(),cmd.arguments[0])
	if err == nil {
		os.Exit(1)
	}
	if err != sql.ErrNoRows {
		return err
	}

	id := uuid.New()
	newUserParams := database.CreateUserParams{
		ID: int64(id.ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.arguments[0],
	}

	user, err := s.db.CreateUser(context.Background(), newUserParams)
	if err != nil {
		return err
	}

	err = s.cfg.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}

	fmt.Printf("User has been created in database: %v\n", user)

	return nil
}

func resetHandler(s *state, cmd command) error {
	if len(cmd.arguments) > 0 {
		return fmt.Errorf("too many arguments")
	}

	userCount, err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("%d users have been deleted from the database.", userCount)

	return nil
}

func usersHandler(s *state, cmd command) error {
	if len(cmd.arguments) > 0 {
		return fmt.Errorf("too many arguments")
	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for i := 0; i < len(users); i++ {
		if users[i].Name == s.cfg.CurrentUser {
			fmt.Printf("* %s (current)\n", users[i].Name)
		} else {
			fmt.Printf("* %s\n", users[i].Name)
		}
		
	}

	return nil
}

func aggHandler(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("incorrect number of arguments (expected 1)")
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return err
	}

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <- ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			return err
		}
	}
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	_, err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return err
	}

	rf, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("Error fetching rss: %s", err)
		return err
	}

	newFeedItems := rf.Channel.Item
	for i := range newFeedItems {
		postID := uuid.New()

		layout := "Mon, 02 Jan 2006 15:04:05 -0700"
		parsedTime, err := time.Parse(layout, newFeedItems[i].PubDate)
		if err != nil {
			return err
		}

		postParams := database.CreatePostParams{
			ID: int64(postID.ID()),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title: newFeedItems[i].Title,
			Url: newFeedItems[i].Link,
			Description: sql.NullString{
				String: newFeedItems[i].Description,
				Valid: true,
			},
			PublishedAt: parsedTime,
			FeedID: feed.ID,
		}

		_, err = s.db.CreatePost(context.Background(), postParams)
		if err, ok := err.(*pq.Error); ok {
			if err.Message != "duplicate key value violates unique constraint \"posts_url_key\"" {
				fmt.Println(err)
				return err
			}
		}
	}

	fmt.Printf("Posts from feed %s saved.\n", feed.Name)

	return nil
}

func addFeedHandler(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 2 {
		return fmt.Errorf("not enough arguments (expected 2)")
	}
	
	id := uuid.New()

	feedParams := database.CreateFeedParams{
		ID: int64(id.ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.arguments[0],
		Url: cmd.arguments[1],
		UserID: user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return err
	}

	followParams := database.CreateFeedFollowParams{
		ID: int64(id.ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	}

	_, err = s.db.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		return err
	}

	fmt.Printf("Feed %s successfully added for user %s\n", feed.Name, user.Name)

	return nil
}

func feedsHandler(s *state, cmd command) error {
	if len(cmd.arguments) > 0 {
		return fmt.Errorf("too many arguments")
	}

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for i := 0; i < len(feeds); i++ {
		user, err := s.db.GetUserByID(context.Background(), feeds[i].UserID)
		if err != nil {
			return err
		}

		fmt.Printf("* Name: %s || URL: %s || User: %s\n", feeds[i].Name, feeds[i].Url, user.Name)
	}

	return nil
}

func followHandler(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("incorrect number of arguments (expected 1)")
	}

	id := uuid.New()

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.arguments[0])
	if err != nil {
		return err
	}

	followParams := database.CreateFeedFollowParams{
		ID: int64(id.ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	}

	newFeedFollow, err := s.db.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		return err
	}

	fmt.Printf("%s has followed %s.\n", newFeedFollow.UserName, newFeedFollow.FeedName)

	return nil
}

func followingHandler (s *state, cmd command, user database.User) error {
	if len(cmd.arguments) > 0 {
		return fmt.Errorf("too many arguments")
	}

	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	fmt.Printf("User %s is following feeds: \n", s.cfg.CurrentUser)

	for i := 0; i < len(feedFollows); i++ {
		fmt.Printf("- %s\n", feedFollows[i].FeedName)
	}

	return nil
}

func unFollowHandler (s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("incorrect number of arguments (expected 1)")
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.arguments[0])
	if err != nil {
		return err
	}

	deleteParams := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	_ , err = s.db.DeleteFeedFollow(context.Background(), deleteParams)
	if err != nil {
		return err
	}

	fmt.Printf("%s has unfollowed %s\n", user.Name, feed.Url)
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
    fmt.Println("Error loading .env file")
  }
	
	dbURL := os.Getenv("DB_URL")
	sqldb, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println(err)
	}
	dbQueries := database.New(sqldb)

	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}

	s := state{
		cfg: &cfg,
		db: dbQueries,
	}

	cmds := commands{
		mapping: make(map[string]func(*state, command) error),
	}

	cmds.register("login", loginHandler)
	cmds.register("register", registerHandler)
	cmds.register("reset", resetHandler)
	cmds.register("users", usersHandler)
	cmds.register("agg", aggHandler)
	cmds.register("addfeed", middlewareLoggedIn(addFeedHandler))
	cmds.register("feeds", feedsHandler)
	cmds.register("follow", middlewareLoggedIn(followHandler))
	cmds.register("following", middlewareLoggedIn(followingHandler))
	cmds.register("unfollow", middlewareLoggedIn(unFollowHandler))

	args := os.Args

	if len(args) < 2 {
		fmt.Println("Error: Not enough arguments")
		os.Exit(1)
	}

	cmdName := args[1]
	cmdArgs := args[2:]

	cmd := command{
		name: cmdName,
		arguments: cmdArgs,
	}

	err = cmds.run(&s, cmd)
	if err != nil {
		fmt.Printf("Error running command: %s\n", err)
		os.Exit(1)
	}
}
