const redis = require('redis')
const { promisify } = require("util")

function redisGetAllKeys(client) {
    console.log("redis get all keys")
    return new Promise((resolve, reject) => {
        client.keys('*', (err, keys) => {
            if (err) {
                console.error("redis get all keys failed", err)
                reject(err)
            } else {
                resolve({ keys: keys, client: client })
            }
        })
    })
}

module.exports = async function (credentials, runServer) {
    let content = ""    
    const client = redis.createClient({
        host: credentials.host,
        port: credentials.tls_port ? credentials.tls_port : credentials.port,
        password: credentials.password,
        retry_strategy: (options) => {
            if (options.error && options.error.code === "ECONNREFUSED") {
                // End reconnecting on a specific error and flush all commands with
                // a individual error
                return new Error("The server refused the connection");
            }
            console.log("redis client retry", options)
        },
        tls: credentials.tls_port ? {} : null,
    });

    return redisGetAllKeys(client).then((result) => {
        if (result.keys.length > 0) {
            content += JSON.stringify(result.keys)
            return Promise.resolve(result.client)
        }
        return Promise.reject("empty collection")
    }).then(() => {
        runServer(content)
    }).catch((error) => {
        console.error(error)
        throw new Error("redis test failed", error)
    })
}