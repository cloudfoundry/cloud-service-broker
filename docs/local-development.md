# Running CSB locally

### Overview

This tutorial explains how to run CSB within the GoLand IDE, allowing for advanced debugging. 

### Requirements

Eden CLI
GoLand IDE
MySQL server

### Steps

1. Claim an environment with command `claim azure`
2. Get environment variables
```azure
ARM_CLIENT_ID=d5a12653-9991-47e9-b766-5cf9c7448de7
ARM_CLIENT_SECRET=ac844dfd18d721356a0d2dd77132e27da8488b42c6ee5722e9123230745ac753f53e294ee3007e9e063c16c069579e94
ARM_SUBSCRIPTION_ID=a5783b82-1c54-4ce7-9669-606db6128be8
ARM_TENANT_ID=b39138ca-3cee-4b4a-a4d6-cd83d9dd62f0
SECURITY_USER_NAME=brokeruser
SECURITY_USER_PASSWORD=brokeruserpassword
GSB_PROVISION_DEFAULTS={"resource_group":"apple-loon"}
DB_HOST=0.0.0.0
DB_USERNAME=root
DB_PASSWORD=password
DB_TLS=skip-verify
PORT=8080

```

3. Start local MySQL server
```azure
docker run -p 3306:3306 -e "MYSQL_ROOT_PASSWORD=password" -d mysql:latest

```
4. Create the servicebroker database:`
```mysql -u root -h 0.0.0.0 -p```
```CREATE DATABASE servicebroker;```

4. Run make build in the Brokerpak directory, having changed the binaries to use Darwin.
5. Amend IDE config:
   1. Files CSB
   2. Working directory CSB-brokerpak
   3. Environment (above values)
   4. Programme arguments - serve

6. Run in Debug mode in IDE
7. Export ENV vars:
```
export SB_BROKER_URL=http://localhost:8080
export SB_BROKER_USERNAME=brokeruser
export SB_BROKER_PASSWORD=brokeruserpassword
```
8. Run `eden catalog`

