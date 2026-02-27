package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "laserbeak",
	Short: "Laserbeak — a Discord LLM bot that listens to voice and chats in text",
	Long: `Laserbeak is a Discord bot built in Go that:
  - Responds to text commands with LLM-powered replies
  - Joins voice channels to listen to conversations
  - Transcribes voice to text via OpenAI Whisper
  - Sends LLM responses as text messages in Discord`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "config file path (default: ./config.yaml)")
	rootCmd.PersistentFlags().String("discord-token", "", "Discord bot token")
	rootCmd.PersistentFlags().String("llm-api-key", "", "LLM API key")
	rootCmd.PersistentFlags().String("llm-base-url", "", "LLM API base URL")
	rootCmd.PersistentFlags().String("llm-model", "", "LLM model name")
	rootCmd.PersistentFlags().String("stt-api-key", "", "STT API key")
	rootCmd.PersistentFlags().String("command-prefix", "", "Bot command prefix")

	viper.BindPFlag("discord.token", rootCmd.PersistentFlags().Lookup("discord-token"))
	viper.BindPFlag("llm.apikey", rootCmd.PersistentFlags().Lookup("llm-api-key"))
	viper.BindPFlag("llm.baseurl", rootCmd.PersistentFlags().Lookup("llm-base-url"))
	viper.BindPFlag("llm.model", rootCmd.PersistentFlags().Lookup("llm-model"))
	viper.BindPFlag("stt.apikey", rootCmd.PersistentFlags().Lookup("stt-api-key"))
	viper.BindPFlag("discord.commandprefix", rootCmd.PersistentFlags().Lookup("command-prefix"))
}
