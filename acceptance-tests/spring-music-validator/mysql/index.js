var mysql = require('mysql');

// reject with error
// resolve with connection
function connectMysql(params) {
    return new Promise((resolve, reject) => {
        let con = mysql.createConnection(params)
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
    return connectMysql({
        host: credentials.hostname,
        user: credentials.username,
        password: credentials.password,
        port: credentials.port,
        database: credentials.name
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
        return connectMysql(credentials.uri)
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