// SPDX-FileCopyrightText: 2026 彭旭
// SPDX-License-Identifier: AGPL-3.0-only
package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Storage   StorageConfig   `mapstructure:"storage"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	App       AppConfig       `mapstructure:"app"`
	Chat      ChatConfig      `mapstructure:"chat"`
	Qdrant    QdrantConfig    `mapstructure:"qdrant"`
	Embedding EmbeddingConfig `mapstructure:"embedding"`
	Surreal   SurrealConfig   `mapstructure:"surrealdb"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
	Mode string `mapstructure:"mode"`
}

type StorageConfig struct {
	DataDir string `mapstructure:"dataDir"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireDays int    `mapstructure:"expireDays"`
}

type AppConfig struct {
	Name       string `mapstructure:"name"`
	Version    string `mapstructure:"version"`
	DeployMode string `mapstructure:"deployMode"`
}

type ChatConfig struct {
	MergeWindowMs          int `mapstructure:"mergeWindowMs"`
	ContextWindowMaxRounds int `mapstructure:"contextWindowMaxRounds"`
}

type QdrantConfig struct {
	Host           string                      `mapstructure:"host"`
	Port           int                         `mapstructure:"port"`
	CollectionName string                      `mapstructure:"collectionName"`
	VectorDim      int                         `mapstructure:"vectorDim"`
	Limit          int                         `mapstructure:"limit"`
	Collections    map[string]CollectionConfig `mapstructure:"collections"`
}

type CollectionConfig struct {
	Name      string `mapstructure:"name"`
	VectorDim int    `mapstructure:"vectorDim"`
}

type EmbeddingConfig struct {
	ModelName string `mapstructure:"modelName"`
	BaseUrl   string `mapstructure:"baseUrl"`
	ApiKey    string `mapstructure:"apiKey"`
}

type SurrealConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	Database  string `mapstructure:"database"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	DataPath  string `mapstructure:"dataPath"`
}

var AppCfg *Config

func InitConfig(configPath string) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yml")
	v.AddConfigPath(configPath)
	v.AddConfigPath(".")

	v.SetDefault("server.port", 8900)
	v.SetDefault("server.host", "127.0.0.1")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("storage.dataDir", "../data")
	v.SetDefault("jwt.secret", "u-ai-secret-key-change-me")
	v.SetDefault("jwt.expireDays", 7)
	v.SetDefault("app.name", "U-Ai")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.deployMode", "desktop-local")
	v.SetDefault("chat.contextWindowMaxRounds", 20)
	v.SetDefault("chat.mergeWindowMs", 6000)
	v.SetDefault("qdrant.host", "127.0.0.1")
	v.SetDefault("qdrant.port", 9178)
	v.SetDefault("qdrant.collectionName", "memory_embeddings")
	v.SetDefault("qdrant.vectorDim", 2560)
	v.SetDefault("qdrant.limit", 10)
	v.SetDefault("qdrant.collections.memory_embeddings.name", "memory_embeddings")
	v.SetDefault("qdrant.collections.memory_embeddings.vectorDim", 2560)
	v.SetDefault("qdrant.collections.working_memory.name", "working_memory")
	v.SetDefault("qdrant.collections.working_memory.vectorDim", 2560)
	v.SetDefault("qdrant.collections.user_profiles.name", "user_profiles")
	v.SetDefault("qdrant.collections.user_profiles.vectorDim", 2560)
	v.SetDefault("qdrant.collections.episodic_memories.name", "episodic_memories")
	v.SetDefault("qdrant.collections.episodic_memories.vectorDim", 2560)
	v.SetDefault("embedding.modelName", "doubao-embedding-vision-251215")
	v.SetDefault("embedding.baseUrl", "")
	v.SetDefault("embedding.apiKey", "")
	v.SetDefault("surrealdb.host", "127.0.0.1")
	v.SetDefault("surrealdb.port", 8000)
	v.SetDefault("surrealdb.namespace", "uai")
	v.SetDefault("surrealdb.database", "memory_graph")
	v.SetDefault("surrealdb.username", "root")
	v.SetDefault("surrealdb.password", "root")
	v.SetDefault("surrealdb.dataPath", "data/graph.db")

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("[Config] 未找到配置文件，使用默认值: %v\n", err)
	} else {
		fmt.Printf("[Config] 已加载配置: %s\n", v.ConfigFileUsed())
	}

	AppCfg = &Config{}
	if err := v.Unmarshal(AppCfg); err != nil {
		log.Fatalf("配置解析失败: %v", err)
	}

	ensureJWTSecret()

	v.WatchConfig()
}

func ensureJWTSecret() {
	if AppCfg.JWT.Secret != "" && AppCfg.JWT.Secret != "u-ai-secret-key-change-me" {
		return
	}
	dataDir := AppCfg.Storage.DataDir
	if dataDir == "" {
		dataDir = "data"
	}
	keyPath := filepath.Join(dataDir, "jwt_secret.key")
	if data, err := os.ReadFile(keyPath); err == nil {
		if s := strings.TrimSpace(string(data)); s != "" {
			AppCfg.JWT.Secret = s
			fmt.Printf("[Config] 已加载持久化 JWT 密钥: %s\n", keyPath)
			return
		}
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		log.Fatalf("生成 JWT 密钥失败: %v", err)
	}
	secret := hex.EncodeToString(buf)
	if err := os.MkdirAll(dataDir, 0755); err == nil {
		if err := os.WriteFile(keyPath, []byte(secret), 0600); err != nil {
			fmt.Printf("[Config] JWT 密钥持久化失败，本次使用内存随机密钥: %v\n", err)
		} else {
			fmt.Printf("[Config] 已生成并持久化随机 JWT 密钥: %s\n", keyPath)
		}
	}
	AppCfg.JWT.Secret = secret
}

func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
