// Memory API Proxy Server
// This script acts as a proxy between the Windsurf integration and the Memory Client API

const http = require('http');
const crypto = require('crypto');

// Configuration
const config = {
    port: 10011,
    targetUrl: 'http://localhost:10010',
    debug: true
};

// Create the server
const server = http.createServer((req, res) => {
    // Set CORS headers
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
    
    // Handle preflight requests
    if (req.method === 'OPTIONS') {
        res.statusCode = 204;
        res.end();
        return;
    }
    
    // Only handle POST requests to /api/message
    if (req.method === 'POST' && req.url === '/api/message') {
        // Read request body
        let body = '';
        req.on('data', chunk => {
            body += chunk.toString();
        });
        
        req.on('end', () => {
            try {
                // Parse the request body
                const data = JSON.parse(body);
                console.log('Received message:', data);
                
                // Check if required fields are present
                if (!data.role || !data.content) {
                    res.statusCode = 400;
                    res.end('Missing required fields: role and content');
                    return;
                }
                
                // Only send the required fields to the Memory Client API
                const messageData = {
                    role: data.role,
                    content: data.content
                };
                
                // Forward the request to the Memory Client API
                const options = {
                    hostname: 'localhost',
                    port: 10010,
                    path: '/api/message',
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                };
                
                const proxyReq = http.request(options, (proxyRes) => {
                    let responseData = '';
                    
                    proxyRes.on('data', (chunk) => {
                        responseData += chunk;
                    });
                    
                    proxyRes.on('end', () => {
                        // Forward the response status and data
                        res.statusCode = proxyRes.statusCode;
                        
                        // Set response headers
                        Object.keys(proxyRes.headers).forEach(key => {
                            res.setHeader(key, proxyRes.headers[key]);
                        });
                        
                        // If successful, return a success message with the original ID
                        if (proxyRes.statusCode === 200) {
                            try {
                                const responseObj = JSON.parse(responseData);
                                // Include the original ID if it was provided
                                if (data.id) {
                                    responseObj.id = data.id;
                                }
                                res.end(JSON.stringify(responseObj));
                            } catch (e) {
                                // If we can't parse the response, just return it as-is
                                res.end(responseData);
                            }
                        } else {
                            // For error responses, try to make them more user-friendly
                            try {
                                // If the error is from Qdrant about missing ids, return a more helpful message
                                if (responseData.includes('missing field `ids`')) {
                                    console.error('Qdrant error about missing ids field. This should be fixed in the updated Memory Client.');
                                    res.statusCode = 500;
                                    res.end(JSON.stringify({
                                        success: false,
                                        message: 'Error from Qdrant about missing ids field. Please ensure the Memory Client has been updated.',
                                        error: responseData
                                    }));
                                } else {
                                    res.end(responseData);
                                }
                            } catch (e) {
                                res.end(responseData);
                            }
                        }
                        
                        console.log(`Response: ${proxyRes.statusCode}`);
                    });
                });
                
                proxyReq.on('error', (error) => {
                    console.error('Error forwarding request:', error);
                    res.statusCode = 500;
                    res.end('Error forwarding request: ' + error.message);
                });
                
                // Send the request data
                const requestData = JSON.stringify(messageData);
                proxyReq.write(requestData);
                proxyReq.end();
                
            } catch (error) {
                console.error('Error processing request:', error);
                res.statusCode = 400;
                res.end('Error processing request: ' + error.message);
            }
        });
    } else if (req.method === 'POST' && req.url === '/api/set-conversation-tag') {
        // Handle set tag requests
        let body = '';
        req.on('data', chunk => {
            body += chunk.toString();
        });
        
        req.on('end', () => {
            try {
                // Parse the request body
                const data = JSON.parse(body);
                console.log('Setting conversation tag:', data);
                
                // Forward the request to the Memory Client API
                const options = {
                    hostname: 'localhost',
                    port: 10010,
                    path: '/api/set-conversation-tag',
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                };
                
                const proxyReq = http.request(options, (proxyRes) => {
                    let responseData = '';
                    
                    proxyRes.on('data', (chunk) => {
                        responseData += chunk;
                    });
                    
                    proxyRes.on('end', () => {
                        // Forward the response status and data
                        res.statusCode = proxyRes.statusCode;
                        
                        // Set response headers
                        Object.keys(proxyRes.headers).forEach(key => {
                            res.setHeader(key, proxyRes.headers[key]);
                        });
                        
                        res.end(responseData);
                        console.log(`Response: ${proxyRes.statusCode}`);
                    });
                });
                
                proxyReq.on('error', (error) => {
                    console.error('Error forwarding request:', error);
                    res.statusCode = 500;
                    res.end('Error forwarding request: ' + error.message);
                });
                
                // Send the request data
                const requestData = JSON.stringify(data);
                proxyReq.write(requestData);
                proxyReq.end();
                
            } catch (error) {
                console.error('Error processing request:', error);
                res.statusCode = 400;
                res.end('Error processing request: ' + error.message);
            }
        });
    } else if (req.method === 'GET' && req.url === '/api/get-conversation-tag') {
        // Forward the request to the Memory Client API
        const options = {
            hostname: 'localhost',
            port: 10010,
            path: '/api/get-conversation-tag',
            method: 'GET',
            headers: req.headers
        };
        
        // Update the host header
        options.headers.host = 'localhost:10010';
        
        const proxyReq = http.request(options, (proxyRes) => {
            // Forward the response status and headers
            res.writeHead(proxyRes.statusCode, proxyRes.headers);
            
            // Pipe the response data
            proxyRes.pipe(res);
        });
        
        proxyReq.on('error', (error) => {
            console.error('Error forwarding request:', error);
            res.statusCode = 500;
            res.end('Error forwarding request: ' + error.message);
        });
        
        proxyReq.end();
    } else {
        // Forward all other requests as-is
        const options = {
            hostname: 'localhost',
            port: 10010,
            path: req.url,
            method: req.method,
            headers: req.headers
        };
        
        // Update the host header
        options.headers.host = 'localhost:10010';
        
        const proxyReq = http.request(options, (proxyRes) => {
            // Forward the response status and headers
            res.writeHead(proxyRes.statusCode, proxyRes.headers);
            
            // Pipe the response data
            proxyRes.pipe(res);
        });
        
        proxyReq.on('error', (error) => {
            console.error('Error forwarding request:', error);
            res.statusCode = 500;
            res.end('Error forwarding request: ' + error.message);
        });
        
        // Forward the request body if present
        if (req.method !== 'GET' && req.method !== 'HEAD') {
            req.pipe(proxyReq);
        } else {
            proxyReq.end();
        }
    }
});

// Generate a UUID for messages
function generateUUID() {
    return crypto.randomUUID();
}

// Start the server
server.listen(config.port, () => {
    console.log(`Memory API Proxy running at http://localhost:${config.port}`);
    console.log(`Forwarding requests to ${config.targetUrl}`);
});
