const vcapServices = require('vcap_services');
const restify = require('restify');

function runServer(content) {
    const tracer = require('@google-cloud/trace-agent').get();
    const server = restify.createServer();
    server.get('/', (_, res, next) => {
        const customSpan = tracer.createChildSpan({ name: 'root-content' });
        res.send(content)
        customSpan.endSpan();
        next()
    });

    server.listen(process.env.PORT || 8080, function () {
        console.log('%s listening at %s', server.name, server.url);
    });
}

let credentials = vcapServices.findCredentials({ instance: { tags: 'tracing' } });

async function main() {
    require('@google-cloud/trace-agent').start({
        projectId: credentials.ProjectId,
        credentials: {
            client_email: credentials.Email,
            private_key: credentials.PrivateKeyData
        }
    })
    runServer("Should see a some trace data in GCP")
}

if (Object.keys(credentials).length > 0) {
    try {
        main()
    } catch (err) {
        console.error("Failure", err)
    }
} else {
    console.error("No stack driver trace creds in vcap_services")
}