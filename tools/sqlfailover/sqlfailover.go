package main

import (
	"context"
	"fmt"
	sqlsdk "github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/v3.0/sql"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	"log"
	"os"
)

var server string
var subid string
var resgroup string
var failovergroup string

// https://github.com/Azure/azure-sdk-for-go/blob/master/services/preview/sql/mgmt/v3.0/sql/failovergroups.go
// The only available documentation on calling this API.
func main() {

	var err error
	if len(os.Args) < 3 {
		log.Fatal("Usage: sqlfailover <resource-group> <server> <failover-group>")
		os.Exit(1)
	}
	temp := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(temp) == 0 {
		log.Fatal("Environment variable AZURE_SUBSCRIPTION_ID not found")
		os.Exit(2)
	}

	subid = temp
	temp = os.Getenv("AZURE_TENANT_ID")
	if len(temp) == 0 {
		log.Fatal("Environment variable AZURE_TENANT_ID not found")
		os.Exit(2)
	}

	temp = os.Getenv("AZURE_CLIENT_ID")
	if len(temp) == 0 {
		log.Fatal("Environment variable AZURE_CLIENT_ID not found")
		os.Exit(2)
	}

	temp = os.Getenv("AZURE_CLIENT_SECRET")
	if len(temp) == 0 {
		log.Fatal("Environment variable AZURE_CLIENT_SECRET not found")
		os.Exit(2)
	}

	resgroup = os.Args[1]
	server = os.Args[2]
	failovergroup = os.Args[3]

	// Create auth token from env variables (see here for details https://github.com/Azure/azure-sdk-for-go)
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err == nil {
		// Create AzureSQL SQL Failover Groups client
		dbclient := sqlsdk.NewFailoverGroupsClient(subid)
		dbclient.Authorizer = authorizer

		ctx := context.Background()
		dbclient.Failover(ctx, resgroup, server, failovergroup)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}
