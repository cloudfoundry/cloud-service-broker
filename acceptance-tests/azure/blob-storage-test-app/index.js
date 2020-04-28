const vcapServices = require('vcap_services');
const restify = require('restify');
const { BlobServiceClient, StorageSharedKeyCredential } = require("@azure/storage-blob");

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

let credentials = vcapServices.findCredentials({ instance: { tags: 'storage' } });

async function main() {
    const sharedKeyCredential = new StorageSharedKeyCredential(credentials.storage_account_name, credentials.primary_access_key);
    const blobServiceClient = new BlobServiceClient(
        `https://${credentials.storage_account_name}.blob.core.windows.net`,
        sharedKeyCredential
    );
    content = []
    let iter = await blobServiceClient.listContainers();
    for await (const container of iter) {
        content.push(container)
    }
    runServer(content)
}

if (Object.keys(credentials).length > 0) {
    try {
        main()
    } catch (err) {
        console.error("Failed to list containers", err)
    }
} else {
    console.error("No S3 creds in vcap_services")
}