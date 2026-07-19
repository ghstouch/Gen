# Gen API Documentation

## Overview

Gen is a lightweight AI router that provides OpenAI-compatible API endpoints. It automatically routes requests to multiple AI providers with fallback support.

**Base URL**: `http://localhost:2500`

## Authentication

Set API keys via environment variables:

```bash
export GEN_OPENAI_KEY=sk-xxx
export GEN_GROQ_KEY=gsk_xxx
export GEN_TOGETHER_KEY=xxx
export GEN_OPENROUTER_KEY=sk-or-xxx
export GEN_GEMINI_KEY=xxx
```

## Endpoints

### Chat Completions

**POST** `/v1/chat/completions`

Create a chat completion.

#### Request Body

```json
{
  "model": "gpt-4",
  "messages": [
    {
      "role": "user",
      "content": "Hello!"
    }
  ],
  "stream": false,
  "max_tokens": 1024,
  "temperature": 0.7
}
```

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| model | string | Yes | Model identifier |
| messages | array | Yes | Array of message objects |
| stream | boolean | No | Enable streaming (default: false) |
| max_tokens | integer | No | Maximum tokens to generate |
| temperature | float | No | Sampling temperature (0-2) |
| top_p | float | No | Nucleus sampling parameter |
| n | integer | No | Number of completions to generate |
| stop | array | No | Stop sequences |
| presence_penalty | float | No | Presence penalty (-2 to 2) |
| frequency_penalty | float | No | Frequency penalty (-2 to 2) |
| user | string | No | User identifier |

#### Message Object

```json
{
  "role": "user",
  "content": "Hello!"
}
```

| Field | Type | Description |
|-------|------|-------------|
| role | string | "system", "user", or "assistant" |
| content | string | Message content |

#### Response (Non-streaming)

```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1700000000,
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 15,
    "total_tokens": 25
  }
}
```

#### Response (Streaming)

When `stream: true`, response is Server-Sent Events (SSE):

```
data: {"id":"chatcmpl-123","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

#### Error Response

```json
{
  "error": {
    "message": "Invalid API key",
    "type": "authentication_error",
    "code": "401"
  }
}
```

### List Models

**GET** `/v1/models`

List all available models.

#### Response

```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4",
      "object": "model",
      "created": 1700000000,
      "owned_by": "openai"
    },
    {
      "id": "llama-3.3-70b-versatile",
      "object": "model",
      "created": 1700000000,
      "owned_by": "groq"
    }
  ]
}
```

### Health Check

**GET** `/health`

Check server health and status.

#### Response

```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": "5m30s",
  "providers": 3,
  "port": 2500
}
```

### Provider Statistics

**GET** `/stats`

Get provider statistics and status.

#### Response

```json
[
  {
    "name": "openai",
    "priority": 1,
    "enabled": true,
    "failures": 0,
    "in_cooldown": false
  },
  {
    "name": "groq",
    "priority": 2,
    "enabled": true,
    "failures": 2,
    "in_cooldown": false
  }
]
```

## Error Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 400 | Bad Request - Invalid JSON or missing required fields |
| 401 | Unauthorized - Invalid or missing API key |
| 403 | Forbidden - Access denied |
| 404 | Not Found |
| 405 | Method Not Allowed |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error |
| 503 | Service Unavailable - All providers failed |

## Rate Limiting

Gen implements automatic cooldown for failed providers:

- After 3 consecutive failures, a provider is cooled down for 60 seconds
- During cooldown, requests are routed to other available providers
- After cooldown period, the provider is automatically re-enabled

## CORS

CORS is enabled by default with the following configuration:

- **Access-Control-Allow-Origin**: `*`
- **Access-Control-Allow-Methods**: `GET, POST, OPTIONS`
- **Access-Control-Allow-Headers**: `Content-Type, Authorization, x-request-id`

## Environment Variables

| Variable | Description |
|----------|-------------|
| GEN_OPENAI_KEY | OpenAI API key |
| GEN_GROQ_KEY | Groq API key |
| GEN_TOGETHER_KEY | Together API key |
| GEN_OPENROUTER_KEY | OpenRouter API key |
| GEN_GEMINI_KEY | Google Gemini API key |
| GEN_*_DISABLE | Disable provider (set to 1) |

## Examples

### cURL

```bash
curl http://localhost:2500/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Python

```python
import requests

response = requests.post(
    "http://localhost:2500/v1/chat/completions",
    json={
        "model": "gpt-4",
        "messages": [{"role": "user", "content": "Hello!"}]
    }
)

print(response.json())
```

### JavaScript

```javascript
const response = await fetch("http://localhost:2500/v1/chat/completions", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    model: "gpt-4",
    messages: [{ role: "user", content: "Hello!" }]
  })
});

const data = await response.json();
console.log(data);
```