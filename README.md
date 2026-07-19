# Gen - AI Router (Go)

Lightweight AI router written in Go. Route AI API requests to multiple providers with automatic fallback.

## Features

- 🚀 **Single binary** - No dependencies, just download and run
- 🔄 **Auto fallback** - Automatically switches providers on failure
- 📡 **OpenAI-compatible** - Drop-in replacement for OpenAI API
- 🎯 **40+ providers** - Route to OpenAI, Groq, Together, OpenRouter, Gemini, and more
- 💾 **Lightweight** - <10MB binary, minimal memory footprint
- ⚡ **Streaming** - Full SSE streaming support
- 🛡️ **Cooldown** - Failed providers automatically cooled down

## Quick Start

```bash
# Download latest release
# Or build from source:
go build -o gen .

# Run (port 2500 hardcoded)
./gen

# Or cross-compile:
GOOS=linux GOARCH=amd64 go build -o gen .
GOOS=darwin GOARCH=arm64 go build -o gen .
```

## Configuration

Set API keys via environment variables:

```bash
export GEN_OPENAI_KEY=sk-xxx
export GEN_GROQ_KEY=gsk_xxx
export GEN_TOGETHER_KEY=xxx
export GEN_OPENROUTER_KEY=sk-or-xxx
export GEN_GEMINI_KEY=xxx
```

Disable a provider:
```bash
export GEN_OPENAI_DISABLE=1
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/chat/completions` | Chat completions |
| GET | `/v1/models` | List available models |
| GET | `/health` | Health check |
| GET | `/stats` | Provider statistics |

## Usage

```bash
curl http://localhost:2500/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## Port

**Hardcoded**: `2500`

## License

MIT