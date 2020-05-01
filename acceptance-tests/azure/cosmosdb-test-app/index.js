const vcapServices = require('vcap_services');
const restify = require('restify');
const CosmosClient = require("@azure/cosmos").CosmosClient;

/*
// This script ensures that the database is setup and populated correctly
*/
async function create(client, databaseId, containerId) {
    const partitionKey = { kind: "Hash", paths: ["/category"] }

    /**
     * Create the database if it does not exist
     */
    const { database } = await client.databases.createIfNotExists({
        id: databaseId
    });
    console.log(`Created database:\n${database.id}\n`);

    /**
     * Create the container if it does not exist
     */
    const { container } = await client
        .database(databaseId)
        .containers.createIfNotExists(
            { id: containerId, partitionKey },
            { offerThroughput: 400 }
        );

    console.log(`Created container:\n${container.id}\n`);
}

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

let credentials = vcapServices.findCredentials({ instance: { tags: 'cosmosdb' } });

async function main() {
    const client = new CosmosClient({ endpoint: credentials.cosmosdb_host_endpoint, key: credentials.cosmosdb_master_key });

    //const database = client.database(credentials.cosmosdb_database_id);
    //const container = database.container("Items");

    // Make sure Tasks database is already setup. If not, create it.
    try {
        await create(client, credentials.cosmosdb_database_id, "Items");
        content = []
        runServer(content)
    } catch (err) {
        console.error("failed creating db", err)
    }
}

if (Object.keys(credentials).length > 0) {
    try {
        main()
    } catch (err) {
        console.error("Failed to list containers", err)
    }
} else {
    console.error("No cosmosdb creds in vcap_services")
}
