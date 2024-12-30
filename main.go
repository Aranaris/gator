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
	if len(cmd.arguments) > 0 {
		return fmt.Errorf("too many arguments")
	}

	rf, err := rss.FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		fmt.Printf("Error fetchin rss: %s", err)
		os.Exit(1)
	}
	fmt.Println(rf)
	return nil
}

func addFeedHandler(s *state, cmd command) error {
	if len(cmd.arguments) != 2 {
		return fmt.Errorf("not enough arguments (expected 2)")
	}
	
	id := uuid.New()
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUser)
	if err != nil {
		return err
	}

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

	fmt.Printf("Feed %s successfully added for user %s", feed.Name, user.Name)

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
	cmds.register("addfeed", addFeedHandler)
	cmds.register("feeds", feedsHandler)

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
