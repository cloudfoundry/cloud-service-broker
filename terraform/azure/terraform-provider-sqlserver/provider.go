package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

type MSSQLConfiguration struct {
	Address  string
	Port     int
	Instance string
	Username string
	Password string
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ADDRESS", ""),
			},
			"port": {
				Type:        schema.TypeInt,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PORT", "1433"),
			},
			"instance": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INSTANCE", ""),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("USERNAME", "sa"),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PASSWORD", ""),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"sqlserver_login":    resourceLogin(),
			"sqlserver_database": resourceDatabase(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	MSSQLConfig := MSSQLConfiguration{
		Address:  d.Get("address").(string),
		Port:     d.Get("port").(int),
		Instance: d.Get("instance").(string),
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
	}

	return &MSSQLConfig, nil
}

func connectToMSSQL(conf *MSSQLConfiguration) (*sql.DB, error) {
	var err error

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;",
		conf.Address, conf.Username, conf.Password, conf.Port)

	if conf.Instance != "" {
		connString = fmt.Sprintf("%s;database=%s", connString, conf.Instance)
	}
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}
	log.Print(fmt.Printf("connected to server at: %s", conf.Address))

	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Print(fmt.Printf("successful ping to server at: %s", conf.Address))

	return db, err
}
