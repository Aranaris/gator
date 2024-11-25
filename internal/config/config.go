package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const ConfigFileName = ".gatorconfig.json"

type Config struct {
	DBurl string `json:"db_url"`
	CurrentUser string `json:"current_user_name"`
}

func Read() (Config, error) {
	fp, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	jsonFile, err := os.Open(fp)
	if err != nil {
		fmt.Printf("Error opening gatorconfig json file: %s", err)
		return Config{}, err
	}

	defer jsonFile.Close()
	data, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("Error reading gatorconfig json data: %s", err)
		return Config{}, err
	}

	var cfg Config
	
	json.Unmarshal(data, &cfg)

	return cfg, nil
}

func getConfigFilePath() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %s", err)
		return "", err
	}

	return h + "/" + ConfigFileName, nil
}


func (cfg *Config) SetUser(u string) error {
	cfg.CurrentUser = u
	err := write(*cfg)
	if err != nil {
		fmt.Printf("Error writing to config file: %s", err)
		return err
	}

	return nil
}

func write(cfg Config) error {
	fp, err := getConfigFilePath()
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(cfg)
	if err != nil {
		fmt.Printf("Error marshalling json: %s", err)
		return err
	}

	err = os.WriteFile(fp, jsonData, 0666)
	if err != nil {
		return err
	}

	return nil
} 
