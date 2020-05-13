const { Client } = require('pg')

// reject with error
// resolve with connection
function connectPostgres(params) {
    return new Promise((resolve) => {
        client = new Client(params)
        client.connect()
        resolve(client)
    })
}

function postgresShowTables(client) {
    return new Promise((resolve, reject) => {
        client.query("SELECT * FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema'", (err, result) => {
            if (err) {
                reject(err)
            } else {
                resolve({ tables: result, client: client })
            }
        })
    })
}

function postgresQuerySpringMusic(client) {
    console.log("postgresQuerySpringMusic")
    return new Promise((resolve, reject) => {
        client.query("SELECT * from album", (err, result) => {
            if (err) {
                reject(err)
            } else {
                resolve({ rows: result, connection: client })
            }
        })
    })
}

module.exports = async function (credentials, runServer) {
    let content = ""
    return connectPostgres({
        host: credentials.hostname,
        user: credentials.username,
        password: credentials.password,
        port: credentials.port,
        database: credentials.name,
        ssl: credentials.use_tls
    }).then((client) => {
        return postgresShowTables(client)
    }).then((result) => {
        content += JSON.stringify(result.tables)
        return Promise.resolve(result.client)
    }).then((client) => {
        return postgresQuerySpringMusic(client)
    }).then((result) => {
        content += JSON.stringify(result.rows)
        return Promise.resolve(result.connection)
    }).then(() => {
        return connectPostgres({
            connectionString: credentials.uri,
            ssl: credentials.use_tls
        })
    }).then((client) => {
        return postgresShowTables(client)
    }).then((result) => {
        content += JSON.stringify(result.tables)
        return Promise.resolve(result.client)
    }).then((client) => {
        return postgresQuerySpringMusic(client)
    }).then((result) => {
        content += JSON.stringify(result.rows)
        return Promise.resolve(result.connection)
    }).then(() => {
        runServer(content)
    }).catch((error) => {
        console.error(error)
        throw new Error("mysql test failed", error)
    })
}