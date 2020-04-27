const restify = require('restify');
const vcapServices = require('vcap_services');
const testMysql = require('./mysql')
const testRedis = require('./redis')
const testMongodb = require('./mongodb')
const testSqlserver = require('./sqlserver')
const testPostgres = require('./postgres')

function runServer(content) {
    const server = restify.createServer();
    server.get('/', (_, res, next) => {
        res.send(content)
        next()
    });

    server.listen(process.env.PORT || 8080, function () {
        console.log('%s listening at %s', server.name, server.url);
    });
}

let tests = [
    { tag: 'mysql', testFunc: testMysql },
    { tag: 'redis', testFunc: testRedis },
    { tag: 'mongodb', testFunc: testMongodb },
    { tag: 'sqlserver', testFunc: testSqlserver },
    { tag: 'postgres', testFunc: testPostgres }
]

async function runTest(credentials, testFunc) {
    try {
        await testFunc(credentials, runServer)
    } catch (e) {
        console.error(e)
    }
}

let testPromises = []

for (test of tests) {
    let credentials = vcapServices.findCredentials({ instance: { tags: test.tag } });

    console.log(test.tag, credentials)
    if (Object.keys(credentials).length > 0) {
        console.log("testing %s", test.tag)
        testPromises.push(runTest(credentials, test.testFunc))
    }
}

if (testPromises.length > 0) {
    Promise.all(testPromises).then(() => {
        console.log('Success')
    }).catch((err) => {
        console.error('Test failure:', err)
    })
} else {
    console.error('No services with tags matching any of:', tests.map((test) => { return test.tag }))
}
