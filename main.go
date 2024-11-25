package main

import (
	"fmt"
	"internal/config"
	"os"
)

type state struct {
	cfg *config.Config
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
	if len(cmd.arguments) <= 0 {
		return fmt.Errorf("username required")
	}

	err := s.cfg.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}

	fmt.Printf("User %s has been set.\n", cmd.arguments[0])

	return nil
}

func main() {
	cmds := commands{
		mapping: make(map[string]func(*state, command) error),
	}

	cmds.register("login", loginHandler)

	args := os.Args

	if len(args) < 2 {
		fmt.Println("Error: Not enough arguments")
		os.Exit(1)
	}

	cmdName := args[1]
	cmdArgs := args[2:]

	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}

	s := state{
		cfg: &cfg,
	}

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
