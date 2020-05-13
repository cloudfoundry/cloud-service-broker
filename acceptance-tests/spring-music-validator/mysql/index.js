const mysql = require('mysql');
const { URL } = require('url')

function getDatabaseConnectionConfig(uri) {
    // Parse a database url, and destructure the result.
    // The database url should be similar to this:
    // mysql://root:somepassword@127.0.0.1:3306/database-name
    const {
        username,
        password,
        port,
        hostname: host,
        pathname = ''
    } = new URL(uri)

    // Prepare connection configuration for mysql.
    return {
        user: unescape(username),
        password,
        host,
        port,
        database: pathname.replace('/', '')
    }
}

function connectionObject(params, use_tls) {
    if (use_tls) {
        return { ...params, ssl: {} }
    }
    return params
}

// reject with error
// resolve with connection
function connectMysql(params) {
    return new Promise((resolve, reject) => {
        console.log(params)
        var con = mysql.createConnection(params)
        con.connect(err => {
            if (err) {
                reject(err)
            } else {
                resolve(con)
            }
        })
    })
}

function mysqlShowTables(con) {
    return new Promise((resolve, reject) => {
        con.query("SHOW tables", (err, result, fields) => {
            if (err) {
                reject(err)
            } else {
                resolve({ tables: result, connection: con })
            }
        })
    })
}

function mysqlQuerySpringMusic(con) {
    return new Promise((resolve, reject) => {
        con.query("SELECT * from album", (err, result, fields) => {
            if (err) {
                reject(err)
            } else {
                resolve({ rows: result, connection: con })
            }
        })
    })
}

module.exports = async function (credentials, runServer) {
    let content = ""
    return connectMysql(connectionObject({
        host: credentials.hostname,
        user: credentials.username,
        password: credentials.password,
        port: credentials.port,
        database: credentials.name
    }, credentials.use_tls)).then((con) => {
        return mysqlShowTables(con)
    }).then((result) => {
        content += JSON.stringify(result.tables)
        return Promise.resolve(result.connection)
    }).then((con) => {
        return mysqlQuerySpringMusic(con)
    }).then((result) => {
        content += JSON.stringify(result.rows)
        return Promise.resolve(result.connection)
    }).then(() => {
        return connectMysql(connectionObject(getDatabaseConnectionConfig(credentials.uri), credentials.use_tls))
    }).then((con) => {
        return mysqlShowTables(con)
    }).then((result) => {
        content += JSON.stringify(result.tables)
        return Promise.resolve(result.connection)
    }).then((con) => {
        return mysqlQuerySpringMusic(con)
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