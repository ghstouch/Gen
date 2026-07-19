# Gen - AI Router (Go)

[![Build](https://github.com/ghstouch/Gen/actions/workflows/build.yml/badge.svg)](https://github.com/ghstouch/Gen/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ghstouch/Gen)](https://goreportcard.com/report/github.com/ghstouch/Gen)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Lightweight AI router written in Go. Route AI API requests to multiple providers with automatic fallback.

## ✨ Features

- 🚀 **Single binary** - No dependencies, just download and run
- 🔄 **Auto fallback** - Automatically switches providers on failure
- 📡 **OpenAI-compatible** - Drop-in replacement for OpenAI API
- 🎯 **40+ providers** - Route to OpenAI, Groq, Together, OpenRouter, Gemini, and more
- 💾 **Lightweight** - <10MB binary, minimal memory footprint
- ⚡ **Streaming** - Full SSE streaming support
- 🛡️ **Cooldown** - Failed providers automatically cooled down
- 🌍 **i18n** - Multi-language support (English, Vietnamese, Chinese)
- 📊 **Monitoring** - Health check and provider statistics endpoints
- 🔒 **CORS** - Built-in CORS support

## 🚀 Quick Start

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/ghstouch/Gen/releases).

### Build from Source

```bash
# Clone repository
git clone https://github.com/ghstouch/Gen.git
cd Gen

# Build
go build -o gen .

# Run (port 2500 hardcoded)
./gen
```

### Cross-compile

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o gen .

# macOS
GOOS=darwin GOARCH=arm64 go build -o gen .

# Windows
GOOS=windows GOARCH=amd64 go build -o gen.exe .
```

### Docker

```bash
# Build
docker build -t gen .

# Run
docker run -p 2500:2500 -e GEN_OPENAI_KEY=sk-xxx gen
```

## ⚙️ Configuration

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

## 📡 API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/chat/completions` | Chat completions |
| GET | `/v1/models` | List available models |
| GET | `/health` | Health check |
| GET | `/stats` | Provider statistics |

### Chat Completion

```bash
curl http://localhost:2500/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Streaming

```bash
curl http://localhost:2500/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": true
  }'
```

### List Models

```bash
curl http://localhost:2500/v1/models
```

### Health Check

```bash
curl http://localhost:2500/health
```

## 📚 Documentation

- [API Documentation](docs/API.md) - Complete API reference
- [Examples](examples/) - Code examples in Python, JavaScript, and Go

## 🏗️ Architecture

```
Gen/
├── main.go                      # Entry point
├── internal/
│   ├── config/config.go         # Configuration
│   ├── handler/handler.go       # HTTP handlers
│   ├── i18n/                    # Internationalization
│   │   ├── translator.go        # Translation engine
│   │   └── locales/             # Translation files
│   ├── models/models.go         # Data models
│   ├── proxy/proxy.go           # HTTP proxy client
│   └── router/router.go         # Routing & fallback
├── docs/                        # Documentation
├── examples/                    # Code examples
└── .github/workflows/           # CI/CD
```

## 🔄 How It Works

1. **Request arrives** at `/v1/chat/completions`
2. **Router selects** the highest-priority available provider
3. **Proxy forwards** the request to the provider
4. **If provider fails**, router tries the next provider (fallback)
5. **After 3 failures**, provider is cooled down for 60 seconds
6. **Response is streamed** back to the client

## 📊 Provider Priority

| Priority | Provider | Models |
|----------|----------|--------|
| 1 | OpenAI | gpt-4o, gpt-4o-mini, gpt-3.5-turbo |
| 2 | Groq | llama-3.3-70b-versatile, mixtral-8x7b-32768 |
| 3 | Together | Llama-3.3-70B-Instruct-Turbo |
| 4 | OpenRouter | gpt-4o, claude-3.5-sonnet |
| 5 | Gemini | gemini-2.0-flash, gemini-1.5-pro |

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/router/...

# Run with coverage
go test -cover ./...
```

## 📦 Port

**Hardcoded**: `2500`

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [OpenAI](https://openai.com/) for the API specification
- [Go](https://golang.org/) for the amazing language
- All the AI providers for their services