package main

import (
	"context"
	"fmt"
	"os"
	"log"
	sqlsdk "github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/v3.0/sql"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	_ "github.com/jinzhu/gorm/dialects/mssql"
)


var server string
var database string
var subid string
var resgroup string

// https://godoc.org/github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/v3.0/sql#DatabasesClient.Failover
// The only available documentation on calling this API.
func main() {

	var err error
	if (len(os.Args) < 4) {
		log.Fatal("Usage: sqlfailover <resource-group> <server> <database> <primary/secondary>")
		os.Exit(1)
	}
	temp := os.Getenv("ARM_SUBSCRIPTION_ID")
	if len(temp) == 0 {
		log.Fatal("Environment variable ARM_SUBSCRIPTION_ID not found")
		os.Exit(2)
	}	

	subid = temp
	temp = os.Getenv("ARM_TENANT_ID")
	if len(temp) == 0 {
		log.Fatal("Environment variable ARM_TENANT_ID not found")
		os.Exit(2)
	}	
	os.Setenv("AZURE_TENANT_ID", temp)	
	temp = os.Getenv("ARM_CLIENT_ID")
	if len(temp) == 0 {
		log.Fatal("Environment variable ARM_CLIENT_ID not found")
		os.Exit(2)
	}	
	os.Setenv("AZURE_CLIENT_ID", temp)
	temp = os.Getenv("ARM_CLIENT_SECRET")
	if len(temp) == 0 {
		log.Fatal("Environment variable ARM_CLIENT_SECRET not found")
		os.Exit(2)
	}	
	os.Setenv("AZURE_CLIENT_SECRET", temp)

	resgroup = os.Args[1]
	server = os.Args[2]
	database = os.Args[3]
	// Create auth token from env variables (see here for details https://github.com/Azure/azure-sdk-for-go)
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err == nil {
		// Create AzureSQL SDK client
		dbclient := sqlsdk.NewDatabasesClient(subid)
		dbclient.Authorizer = authorizer
		var mode sqlsdk.ReplicaType
		if os.Args[4] == "primary" {
			mode = "Primary"
		}
		if os.Args[4] == "secondary" {
			mode = "ReadableSecondary"
		}
		ctx := context.Background()
		future, err := dbclient.Failover(ctx, resgroup, server, database, mode)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			err = future.WaitForCompletionRef(ctx, dbclient.Client)
			if err != nil {
				fmt.Println(err.Error())
			}
		}			
	}
}
