"""
Gen AI Router - Python Example

This example demonstrates how to use the Gen AI Router with Python.
"""

import requests
import json
import os

# Gen server URL
BASE_URL = "http://localhost:2500"

def chat_completion(model="gpt-4", message="Hello!", stream=False):
    """Send a chat completion request to Gen."""
    
    url = f"{BASE_URL}/v1/chat/completions"
    
    payload = {
        "model": model,
        "messages": [
            {"role": "user", "content": message}
        ],
        "stream": stream
    }
    
    headers = {
        "Content-Type": "application/json"
    }
    
    if stream:
        # Streaming response
        response = requests.post(url, json=payload, headers=headers, stream=True)
        
        print(f"Model: {model}")
        print(f"Message: {message}")
        print("-" * 50)
        
        for line in response.iter_lines():
            if line:
                line = line.decode('utf-8')
                if line.startswith('data: '):
                    data = line[6:]
                    if data == '[DONE]':
                        print("\n[DONE]")
                        break
                    
                    try:
                        chunk = json.loads(data)
                        if 'choices' in chunk and len(chunk['choices']) > 0:
                            delta = chunk['choices'][0].get('delta', {})
                            if 'content' in delta:
                                print(delta['content'], end='', flush=True)
                    except json.JSONDecodeError:
                        pass
        
        print()
    else:
        # Non-streaming response
        response = requests.post(url, json=payload, headers=headers)
        
        if response.status_code == 200:
            data = response.json()
            
            print(f"Model: {data['model']}")
            print(f"Message: {message}")
            print("-" * 50)
            
            if 'choices' in data and len(data['choices']) > 0:
                print(data['choices'][0]['message']['content'])
            
            if 'usage' in data:
                usage = data['usage']
                print(f"\nTokens: {usage['total_tokens']} (prompt: {usage['prompt_tokens']}, completion: {usage['completion_tokens']})")
        else:
            print(f"Error: {response.status_code}")
            print(response.json())

def list_models():
    """List all available models."""
    
    url = f"{BASE_URL}/v1/models"
    
    response = requests.get(url)
    
    if response.status_code == 200:
        data = response.json()
        
        print("Available Models:")
        print("-" * 50)
        
        for model in data['data']:
            print(f"  {model['id']} (by {model['owned_by']})")
    else:
        print(f"Error: {response.status_code}")

def health_check():
    """Check server health."""
    
    url = f"{BASE_URL}/health"
    
    response = requests.get(url)
    
    if response.status_code == 200:
        data = response.json()
        
        print("Health Status:")
        print("-" * 50)
        print(f"  Status: {data['status']}")
        print(f"  Version: {data['version']}")
        print(f"  Uptime: {data['uptime']}")
        print(f"  Providers: {data['providers']}")
        print(f"  Port: {data['port']}")
    else:
        print(f"Error: {response.status_code}")

def provider_stats():
    """Get provider statistics."""
    
    url = f"{BASE_URL}/stats"
    
    response = requests.get(url)
    
    if response.status_code == 200:
        data = response.json()
        
        print("Provider Statistics:")
        print("-" * 50)
        
        for provider in data:
            status = "✓" if provider['enabled'] else "✗"
            cooldown = " (cooldown)" if provider['in_cooldown'] else ""
            print(f"  {status} {provider['name']} (priority: {provider['priority']}, failures: {provider['failures']}){cooldown}")
    else:
        print(f"Error: {response.status_code}")

if __name__ == "__main__":
    print("Gen AI Router - Python Example")
    print("=" * 50)
    print()
    
    # Health check
    health_check()
    print()
    
    # List models
    list_models()
    print()
    
    # Provider stats
    provider_stats()
    print()
    
    # Chat completion (non-streaming)
    print("Chat Completion (Non-streaming):")
    print("=" * 50)
    chat_completion(model="gpt-4", message="What is Go programming language?")
    print()
    
    # Chat completion (streaming)
    print("Chat Completion (Streaming):")
    print("=" * 50)
    chat_completion(model="gpt-4", message="Write a short poem about coding.", stream=True)