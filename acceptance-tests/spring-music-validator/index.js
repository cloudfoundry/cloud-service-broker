var restify = require('restify');
var vcapServices = require('vcap_services');
var testMysql = require('./mysql')
var testRedis = require('./redis')

function runServer(content) {
    var server = restify.createServer();
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
]

let credentials = vcapServices.findCredentials({ instance: { tags: tests[0].tag } });

async function runTest(credentials, testFunc) {
    try {
        await testFunc(credentials, runServer)
    } catch (e) {
        console.error(e)
    }
}

for (test of tests) {
    let credentials = vcapServices.findCredentials({ instance: { tags: test.tag } });

    if (Object.keys(credentials).length > 0) {
        console.log("testing %s", test.tag)
        runTest(credentials, test.testFunc)
    }
}

console.error('No services with tags matching any of:', tests.map((test) => { return test.tag }))
