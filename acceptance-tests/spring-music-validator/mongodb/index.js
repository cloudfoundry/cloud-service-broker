const MongoClient = require('mongodb').MongoClient;

function mongoConnect(client) {
    return new Promise((resolve, reject) => {
        client.connect((err) => {
            if (err) {
                console.error("mongo connect failed", err)
                reject(err)
            } else {
                resolve(client)
            }
        })
    })
}

function fetchDocuments(client, dbName, collectionName) {
    return new Promise((resolve, reject) => {
        const db = client.db(dbName)
        const collection = db.collection(collectionName)
        collection.find({}).toArray((error, docs) => {
            if (error) {
                console.error("mongodb collection.find failed", err)
                reject(err)
            } else {
                resolve(docs)
            }
        })
    })
}

module.exports = async function (credentials, runServer) {
    let content = ""
    const client = new MongoClient(credentials.uri)

    return mongoConnect(client).then((client) => {
        return fetchDocuments(client, "musicdb", "album")
    }).then((docs) => {
        if (docs.length > 0) {
            content = JSON.stringify(docs)
            return Promise.resolve()
        }
        return Promise.reject("empty collection")
    }).then(() => {
        runServer(content)
    }).catch((error) => {
        console.error(error)
        throw new Error("mongo test failed", error)
    })
}