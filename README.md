# GoBot-Laserbeak

A Discord LLM bot built in Go that listens to voice channels and responds with text. Built with Cobra, Viper, and Domain Driven Design.

## Features

- **Text Chat**: Respond to text commands with LLM-powered replies
- **Voice Listening**: Join voice channels and transcribe speech via OpenAI Whisper
- **Text Responses**: All responses are sent as text messages in Discord (no TTS)
- **Conversation Memory**: Per-channel conversation history with configurable limits
- **OpenAI Compatible**: Works with any OpenAI-compatible API (OpenAI, Ollama, etc.)

## Architecture (DDD)

```
internal/
├── domain/              # Domain layer — entities, value objects, interfaces
│   ├── conversation/    # Conversation aggregate (entity, message, repository)
│   └── bot/             # LLM and STT service port interfaces
├── application/         # Application layer — use cases / orchestration
│   ├── chat_service.go  # Text chat use case
│   └── voice_service.go # Voice-to-text-to-LLM pipeline
├── infrastructure/      # Infrastructure layer — external adapters
│   ├── discord/         # Discord bot + voice listener
│   ├── llm/             # OpenAI chat + Whisper STT clients
│   ├── audio/           # Opus decoder + WAV encoder
│   └── persistence/     # In-memory conversation repository
└── config/              # Viper configuration loading
cmd/                     # Interface layer — Cobra CLI commands
```

## Quick Start

1. **Copy config**:
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your tokens
   ```

2. **Build**:
   ```bash
   go build -o laserbeak .
   ```

3. **Run**:
   ```bash
   ./laserbeak serve
   ```

Or with environment variables:
```bash
LASERBEAK_DISCORD_TOKEN=... LASERBEAK_LLM_APIKEY=... ./laserbeak serve
```

## Commands

| Command | Description |
|---------|-------------|
| `!laser <message>` | Chat with the LLM |
| `!laser join` | Join your voice channel and start listening |
| `!laser leave` | Leave voice channel |
| `!laser clear` | Clear conversation history |
| `!laser help` | Show available commands |

## Configuration

Configuration is loaded from (in order of precedence):
1. CLI flags (`--discord-token`, `--llm-api-key`, etc.)
2. Environment variables (`LASERBEAK_DISCORD_TOKEN`, etc.)
3. Config file (`config.yaml`)

## Prerequisites

- Go 1.21+
- A Discord bot token with Message Content and Voice intents
- An OpenAI API key (or compatible API)
- libopus (for voice decoding): `apt install libopus-dev` / `brew install opus`
