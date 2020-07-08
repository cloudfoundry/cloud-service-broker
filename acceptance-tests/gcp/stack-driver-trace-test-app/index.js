const vcapServices = require('vcap_services');

function runServer(content) {
    const tracer = require('@google-cloud/trace-agent').get();
    const server = require('restify').createServer();
    server.get('/', (_, res, next) => {
        const customSpan = tracer.createChildSpan({ name: 'gen-content' });
        res.send(`${JSON.stringify(tracer.getConfig())} ${tracer.getWriterProjectId()} ${content}`)
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
        logLevel: 4,
        enabled: true,
        projectId: credentials.ProjectId,
        bufferSize: 1,
        flushDelaySeconds: 0,
        credentials: JSON.parse(Buffer.from(credentials.PrivateKeyData, 'base64').toString("ascii"))
    })
    runServer(`${JSON.parse(Buffer.from(credentials.PrivateKeyData, 'base64').toString("ascii"))} `)
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