const vcapServices = require('vcap_services');
const AWS = require('aws-sdk')
const restify = require('restify');

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

let credentials = vcapServices.findCredentials({ instance: { tags: 's3' } });

if (Object.keys(credentials).length > 0) {
    AWS.config.credentials = {
        accessKeyId: credentials.access_key_id,
        secretAccessKey: credentials.secret_access_key,
        region: credentials.region_name
    }
    s3 = new AWS.S3({ apiVersion: '2006-03-01' })
    s3.listObjects({ Bucket: credentials.bucket_name }, (err, data) => {
        if (err) {
            console.error("Failed listing bucket contents", err)
        } else {
            runServer(data)
        }
    })
} else {
    console.error("No S3 creds in vcap_services")
}
