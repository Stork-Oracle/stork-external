#!/usr/bin/env node

// Simple WebSocket test client for the self-serve chain pusher
// Requirements: npm install ws

const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:8080/ws');

ws.on('open', function() {
    console.log('Connected to WebSocket server');
    
    // Send a test price update message
    const message = {
        type: 'prices',
        data: [
            {
                t: Date.now() * 1000000, // Convert to nanoseconds
                a: 'BTCUSD',
                v: '50000.123456',
                m: {
                    source: 'test_client'
                }
            },
            {
                t: Date.now() * 1000000,
                a: 'ETHUSD', 
                v: '3000.789',
                m: {
                    source: 'test_client'
                }
            }
        ]
    };
    
    console.log('Sending price update:', JSON.stringify(message, null, 2));
    ws.send(JSON.stringify(message));
    
    // Send another message after 2 seconds with different prices to trigger delta
    setTimeout(() => {
        const message2 = {
            type: 'prices',
            data: [
                {
                    t: Date.now() * 1000000,
                    a: 'BTCUSD',
                    v: '51000.555', // 2% increase should trigger push
                    m: {
                        source: 'test_client_delta'
                    }
                }
            ]
        };
        
        console.log('Sending delta trigger update:', JSON.stringify(message2, null, 2));
        ws.send(JSON.stringify(message2));
        
        setTimeout(() => {
            console.log('Closing connection');
            ws.close();
        }, 1000);
    }, 2000);
});

ws.on('message', function(data) {
    console.log('Received:', data.toString());
});

ws.on('error', function(error) {
    console.error('WebSocket error:', error);
});

ws.on('close', function() {
    console.log('Connection closed');
    process.exit(0);
});

console.log('Connecting to WebSocket server at ws://localhost:8080/ws');