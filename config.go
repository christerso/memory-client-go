package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	QdrantURL      string
	CollectionName string
	EmbeddingSize  int
}

func LoadConfig() *Config {
	// Set default config locations
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Look in current directory and home directory
	viper.AddConfigPath(".")

	// Get user home directory for config
	home, err := os.UserHomeDir()
	if err == nil {
		viper.AddConfigPath(filepath.Join(home, ".config", "memory-client"))
	}

	// Enable environment variables
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("QDRANT_URL", "http://localhost:6333")
	viper.SetDefault("COLLECTION_NAME", "conversation_memory")
	viper.SetDefault("EMBEDDING_SIZE", 384)

	// Try to read config file, but don't fail if not found
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config: %v\n", err)
		}
	}

	return &Config{
		QdrantURL:      viper.GetString("QDRANT_URL"),
		CollectionName: viper.GetString("COLLECTION_NAME"),
		EmbeddingSize:  viper.GetInt("EMBEDDING_SIZE"),
	}
}
