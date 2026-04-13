const http = require('http');

// Test database creation
const testDatabaseCreation = () => {
    const data = JSON.stringify({ name: 'TestDB' });

    const options = {
        hostname: 'localhost',
        port: 8081,
        path: '/api/databases',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Content-Length': data.length
        }
    };

    const req = http.request(options, (res) => {
        let body = '';
        res.on('data', (chunk) => {
            body += chunk;
        });
        res.on('end', () => {
            console.log('Database Creation Response:', res.statusCode);
            console.log('Response:', body);
        });
    });

    req.on('error', (e) => {
        console.error('Error:', e);
    });

    req.write(data);
    req.end();
};

// Test using database
const testUseDatabase = () => {
    const data = JSON.stringify({ name: 'TestDB' });

    const options = {
        hostname: 'localhost',
        port: 8081,
        path: '/api/databases/use',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Content-Length': data.length
        }
    };

    const req = http.request(options, (res) => {
        let body = '';
        res.on('data', (chunk) => {
            body += chunk;
        });
        res.on('end', () => {
            console.log('Use Database Response:', res.statusCode);
            console.log('Response:', body);
        });
    });

    req.on('error', (e) => {
        console.error('Error:', e);
    });

    req.write(data);
    req.end();
};

// Run tests
console.log('Testing API endpoints...');
testDatabaseCreation();

setTimeout(() => {
    testUseDatabase();
}, 1000);