const restify = require('restify');
const vcapServices = require('vcap_services');

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

async function testDataproc(credentials, runServer) {
    const dataproc = require('@google-cloud/dataproc');
    // Create a job client with the endpoint set to the desired cluster region
    const jobClient = new dataproc.v1.JobControllerClient({
        apiEndpoint: `${credentials.region}-dataproc.googleapis.com`,
        projectId: credentials.project_id,
        credentials: JSON.parse(Buffer.from(credentials.private_key, 'base64').toString("ascii"))
    });
    
    let content = ""
    return jobClient.listJobs({
        clusterName: credentials.cluster_name,
        projectId: credentials.project_id,
        region: credentials.region
    }).then((jobs) => {
        content += JSON.stringify(jobs)
    }).then(() => {
        runServer(content)
    }).catch((error) => {
        console.error(error)
        throw new Error("dataproc test failed", error) 
    })
}

let tests = [
    { tag: 'dataproc', testFunc: testDataproc }
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