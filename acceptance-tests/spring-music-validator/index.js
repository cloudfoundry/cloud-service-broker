var restify = require('restify');
var vcapServices = require('vcap_services');
var mysql = require('mysql');

// reject with error
// resolve with connection
function connectMysql(params) {
    return new Promise((resolve, reject) => {
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

var credentials = vcapServices.findCredentials({ instance: { tags: 'mysql' } });

var content=""
if (credentials) {
    connectMysql({
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
        var server = restify.createServer();
        server.get('/', (_, res, next) => {
            res.send(content)
            next()
        });

        server.listen(process.env.PORT || 8080, function () {
            console.log('%s listening at %s', server.name, server.url);
        });
    }).catch(error => {
        console.error("mysql connection(s) failed", error)
    })
    
}
