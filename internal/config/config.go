package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Discord  DiscordConfig
	LLM      LLMConfig
	STT      STTConfig
	Bot      BotConfig
}

// DiscordConfig holds Discord-specific settings.
type DiscordConfig struct {
	Token          string
	CommandPrefix  string
	GuildID        string // guild to operate in (required for auto-join)
	VoiceChannelID string // voice channel to auto-join on startup
	TextChannelID  string // text channel for voice command output
}

// LLMConfig holds LLM API settings.
type LLMConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// STTConfig holds speech-to-text API settings.
type STTConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// BotConfig holds general bot behavior settings.
type BotConfig struct {
	SystemPrompt string
	MaxHistory   int
	WakePhrase   string // wake phrase for voice commands (e.g. "hey m'bot")
}

// Load reads configuration from environment variables, config files, and flags.
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.gobot-laserbeak")
	viper.AddConfigPath("/etc/gobot-laserbeak")

	// Environment variable mapping
	viper.SetEnvPrefix("LASERBEAK")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("discord.commandprefix", "!laser")
	viper.SetDefault("llm.baseurl", "https://api.openai.com/v1")
	viper.SetDefault("llm.model", "gpt-4")
	viper.SetDefault("stt.baseurl", "https://api.openai.com/v1")
	viper.SetDefault("stt.model", "whisper-1")
	viper.SetDefault("bot.systemprompt", "You are Laserbeak, a helpful Discord assistant. Respond concisely and helpfully.")
	viper.SetDefault("bot.maxhistory", 50)
	viper.SetDefault("bot.wakephrase", "hey m'bot")

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	cfg := &Config{
		Discord: DiscordConfig{
			Token:          viper.GetString("discord.token"),
			CommandPrefix:  viper.GetString("discord.commandprefix"),
			GuildID:        viper.GetString("discord.guildid"),
			VoiceChannelID: viper.GetString("discord.voicechannelid"),
			TextChannelID:  viper.GetString("discord.textchannelid"),
		},
		LLM: LLMConfig{
			APIKey:  viper.GetString("llm.apikey"),
			BaseURL: viper.GetString("llm.baseurl"),
			Model:   viper.GetString("llm.model"),
		},
		STT: STTConfig{
			APIKey:  viper.GetString("stt.apikey"),
			BaseURL: viper.GetString("stt.baseurl"),
			Model:   viper.GetString("stt.model"),
		},
		Bot: BotConfig{
			SystemPrompt: viper.GetString("bot.systemprompt"),
			MaxHistory:   viper.GetInt("bot.maxhistory"),
			WakePhrase:   viper.GetString("bot.wakephrase"),
		},
	}

	if cfg.Discord.Token == "" {
		return nil, fmt.Errorf("discord.token is required (set LASERBEAK_DISCORD_TOKEN)")
	}
	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("llm.apikey is required (set LASERBEAK_LLM_APIKEY)")
	}

	return cfg, nil
}
