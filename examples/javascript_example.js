/**
 * Gen AI Router - JavaScript/Node.js Example
 * 
 * This example demonstrates how to use the Gen AI Router with JavaScript.
 */

const BASE_URL = 'http://localhost:2500';

async function chatCompletion(model = 'gpt-4', message = 'Hello!', stream = false) {
    const url = `${BASE_URL}/v1/chat/completions`;
    
    const payload = {
        model,
        messages: [
            { role: 'user', content: message }
        ],
        stream
    };
    
    const headers = {
        'Content-Type': 'application/json'
    };
    
    if (stream) {
        // Streaming response
        const response = await fetch(url, {
            method: 'POST',
            headers,
            body: JSON.stringify(payload)
        });
        
        console.log(`Model: ${model}`);
        console.log(`Message: ${message}`);
        console.log('-'.repeat(50));
        
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        
        while (true) {
            const { done, value } = await reader.read();
            
            if (done) {
                console.log('\n[DONE]');
                break;
            }
            
            const text = decoder.decode(value);
            const lines = text.split('\n');
            
            for (const line of lines) {
                if (line.startsWith('data: ')) {
                    const data = line.slice(6);
                    
                    if (data === '[DONE]') {
                        console.log('\n[DONE]');
                        return;
                    }
                    
                    try {
                        const chunk = JSON.parse(data);
                        if (chunk.choices && chunk.choices.length > 0) {
                            const delta = chunk.choices[0].delta || {};
                            if (delta.content) {
                                process.stdout.write(delta.content);
                            }
                        }
                    } catch (e) {
                        // Skip invalid JSON
                    }
                }
            }
        }
        
        console.log();
    } else {
        // Non-streaming response
        const response = await fetch(url, {
            method: 'POST',
            headers,
            body: JSON.stringify(payload)
        });
        
        if (response.ok) {
            const data = await response.json();
            
            console.log(`Model: ${data.model}`);
            console.log(`Message: ${message}`);
            console.log('-'.repeat(50));
            
            if (data.choices && data.choices.length > 0) {
                console.log(data.choices[0].message.content);
            }
            
            if (data.usage) {
                const { prompt_tokens, completion_tokens, total_tokens } = data.usage;
                console.log(`\nTokens: ${total_tokens} (prompt: ${prompt_tokens}, completion: ${completion_tokens})`);
            }
        } else {
            const error = await response.json();
            console.error(`Error: ${response.status}`);
            console.error(error);
        }
    }
}

async function listModels() {
    const url = `${BASE_URL}/v1/models`;
    
    const response = await fetch(url);
    
    if (response.ok) {
        const data = await response.json();
        
        console.log('Available Models:');
        console.log('-'.repeat(50));
        
        for (const model of data.data) {
            console.log(`  ${model.id} (by ${model.owned_by})`);
        }
    } else {
        console.error(`Error: ${response.status}`);
    }
}

async function healthCheck() {
    const url = `${BASE_URL}/health`;
    
    const response = await fetch(url);
    
    if (response.ok) {
        const data = await response.json();
        
        console.log('Health Status:');
        console.log('-'.repeat(50));
        console.log(`  Status: ${data.status}`);
        console.log(`  Version: ${data.version}`);
        console.log(`  Uptime: ${data.uptime}`);
        console.log(`  Providers: ${data.providers}`);
        console.log(`  Port: ${data.port}`);
    } else {
        console.error(`Error: ${response.status}`);
    }
}

async function providerStats() {
    const url = `${BASE_URL}/stats`;
    
    const response = await fetch(url);
    
    if (response.ok) {
        const data = await response.json();
        
        console.log('Provider Statistics:');
        console.log('-'.repeat(50));
        
        for (const provider of data) {
            const status = provider.enabled ? '✓' : '✗';
            const cooldown = provider.in_cooldown ? ' (cooldown)' : '';
            console.log(`  ${status} ${provider.name} (priority: ${provider.priority}, failures: ${provider.failures})${cooldown}`);
        }
    } else {
        console.error(`Error: ${response.status}`);
    }
}

async function main() {
    console.log('Gen AI Router - JavaScript Example');
    console.log('='.repeat(50));
    console.log();
    
    // Health check
    await healthCheck();
    console.log();
    
    // List models
    await listModels();
    console.log();
    
    // Provider stats
    await providerStats();
    console.log();
    
    // Chat completion (non-streaming)
    console.log('Chat Completion (Non-streaming):');
    console.log('='.repeat(50));
    await chatCompletion('gpt-4', 'What is Go programming language?', false);
    console.log();
    
    // Chat completion (streaming)
    console.log('Chat Completion (Streaming):');
    console.log('='.repeat(50));
    await chatCompletion('gpt-4', 'Write a short poem about coding.', true);
}

main().catch(console.error);