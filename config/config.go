package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type URL struct {
	URL           string `yaml:"url"`
	StatusCode    *int   `yaml:"status_code"`
	SlackToken    string `yaml:"slack_token"`
	TelegramToken string `yaml:"telegram_token"`
	Scheduler     string `yaml:"scheduler"`
}

type ConfigAmanda struct {
	Urls []URL `yaml:"urls"`
}

//Load Configuration
func (c *ConfigAmanda) LoadConfig() (*ConfigAmanda, error) {

	err := ValidatePath("config.yml")

	if err != nil {
		log.Println("failed load config.yml")
		return nil, err
	}

	file, err := os.Open(filepath.Clean("config.yml"))

	if err != nil {
		return nil, err
	}

	defer file.Close()

	ymlFile := yaml.NewDecoder(file)

	if err := ymlFile.Decode(&c); err != nil {
		return nil, err
	}

	return c, nil
}

func ValidatePath(path string) error {
	// Check path if exists
	s, err := os.Stat(path)

	if err != nil {
		return err
	}

	// Check is directory

	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory", path)
	}

	return nil
}
