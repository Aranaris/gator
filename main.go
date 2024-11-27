package main

import (
	"context"
	"database/sql"
	"fmt"
	"internal/config"
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

	fmt.Printf("DB URL: %s\n", cfg.DBurl)
	fmt.Printf("Username: %s\n", cfg.CurrentUser)
}
