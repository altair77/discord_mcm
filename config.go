package main

import (
	"os"

	yaml "gopkg.in/yaml.v2"
)

const (
	CONFIG_FILE = "dmcm_config.yml"
)

type Config struct {
	Token         string     `yaml:"token"`
	ChannelID     string     `yaml:"channelId"`
	LaunchCommand string     `yaml:"launchCommand"`
	Prefix        string     `yaml:"prefix"`
	Schedules     []Schedule `yaml:"schedules"`
}

type Schedule struct {
	Type     string `yaml:"type"`
	Command  string `yaml:"command"`
	Datetime string `yaml:"datetime"`
}

func ExistsConfig() bool {
	_, err := os.Stat(CONFIG_FILE)
	return err == nil
}

func NewConfig() *Config {
	return &Config{
		Token:         "your token",
		ChannelID:     "your channel ID",
		LaunchCommand: "java -jar minecraft_server.jar",
		Prefix:        "m!",
		Schedules:     []Schedule{},
	}
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	data, err := os.ReadFile(CONFIG_FILE)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Config) Write() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if err := os.WriteFile(CONFIG_FILE, data, os.ModePerm); err != nil {
		return err
	}
	return nil
}
