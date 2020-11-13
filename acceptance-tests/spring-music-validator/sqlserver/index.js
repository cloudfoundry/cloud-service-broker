const sql = require('mssql')

// reject with error
// resolve with connection
function connectSqlserver(params) {
    return sql.connect(params)
}

function sqlserverShowTables(pool) {
    return pool.request().query("SELECT * FROM INFORMATION_SCHEMA.TABLES").then((result) => {
        return { tables: result, pool: pool }
    })
}

function sqlserverQuerySpringMusic(pool) {
    return pool.request().query("SELECT * from album").then((result) => {
        return { rows: result, pool: pool }
    })
}

module.exports = async function (credentials, runServer) {
    let content = ""

    if (credentials.hostname != credentials.sqlServerFullyQualifiedDomainName ||
        credentials.name != credentials.sqldbName) {
        console.error("hostname does not match sqlServerFullyQualifiedDomainName or name does not match sqldbName ")
        throw new Error("badly formed credentials")
    }
    return connectSqlserver({
        server: credentials.hostname,
        user: credentials.username,
        password: credentials.password,
        port: credentials.port,
        database: credentials.name
    }).then((pool) => {
        return sqlserverShowTables(pool)
    }).then((result) => {
        content += JSON.stringify(result.tables)
        return Promise.resolve(result.pool)
    }).then((pool) => {
        return sqlserverQuerySpringMusic(pool)
    }).then((result) => {
        content += JSON.stringify(result.rows)
        return Promise.resolve(result.connection)    
    }).then(() => {

        return connectSqlserver(credentials.uri)
    }).then((pool) => {
        return sqlserverShowTables(pool)
    }).then((result) => {
        content += JSON.stringify(result.tables)
        return Promise.resolve(result.pool)
    }).then((pool) => {
        return sqlserverQuerySpringMusic(pool)
    }).then((result) => {
        content += JSON.stringify(result.rows)
        return Promise.resolve(result.connection)
    }).then(() => {

        return connectSqlserver({
            server: credentials.sqlServerFullyQualifiedDomainName,
            user: credentials.databaseLogin,
            password: credentials.databaseLoginPassword,
            database: credentials.sqldbName
        })
    }).then((pool) => {
        return sqlserverShowTables(pool)
    }).then((result) => {
        content += JSON.stringify(result.tables)
        return Promise.resolve(result.pool)
    }).then((pool) => {
        return sqlserverQuerySpringMusic(pool)
    }).then((result) => {
        content += JSON.stringify(result.rows)
        return Promise.resolve(result.connection)
    }).then(() => {
        runServer(content)
    }).catch((error) => {
        console.error(error)
        throw new Error("mssql test failed", error)
    })        
}