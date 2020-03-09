package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceLogin() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Login name",
				ForceNew:    true,
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Login password",
				ForceNew:    false,
				Sensitive:   true,
				StateFunc:   hashSum,
			},
			"default_database": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "DEFAULT_DATABASE parameter",
				ForceNew:    false,
			},
			"check_policy": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "CHECK_POLICY parameter",
				ForceNew:    false,
			},
		},
		Read:   resourceLoginRead,
		Create: resourceLoginCreate,
		Update: resourceLoginUpdate,
		Delete: resourceLoginDelete,
	}
}

func resourceLoginCreate(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMSSQL(meta.(*MSSQLConfiguration))
	if err != nil {
		return err
	}

	var stmt strings.Builder
	var stmtOpts strings.Builder
	checkPolicy := d.Get("check_policy")
	name := d.Get("name")
	pass := d.Get("password")

	fmt.Fprintf(&stmt, "CREATE LOGIN %s WITH PASSWORD=N'%s';", name, pass)

	if checkPolicy != "" {
		fmt.Fprintf(&stmtOpts, ", CHECK_POLICY=%s", checkPolicy)
	}

	fmt.Fprintf(&stmt, "%s;", stmtOpts.String())

	log.Print(stmt.String())

	s, err := db.Prepare(stmt.String())
	if err != nil {
		log.Fatal(err)
		return err
	}

	_, err = s.Exec()
	if err != nil {
		log.Fatal(err)
		return err
	}

	d.SetId(name.(string))

	defer db.Close()
	return nil
}

func resourceLoginRead(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMSSQL(meta.(*MSSQLConfiguration))
	resultFound := false
	if err != nil {
		return err
	}

	var stmt strings.Builder
	name := d.Get("name")

	fmt.Fprintf(&stmt, "USE [master]; ")
	fmt.Fprintf(&stmt, "SELECT name FROM syslogins WHERE name = N'%s';", name)
	log.Print(fmt.Printf("resourceLoginRead query: %s", stmt.String()))

	rows, err := db.Query(stmt.String())
	if err != nil {
		log.Fatal("Cannot query: ", err.Error())
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var nameResult string
		err = rows.Scan(&nameResult)
		if err != nil {
			log.Fatal(err)
			continue
		}
		log.Print(fmt.Printf("resourceLoginRead query scan result: %s", nameResult))
		resultFound = true
	}

	if !resultFound {
		d.SetId("")
	}

	defer db.Close()
	return nil
}

func resourceLoginUpdate(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMSSQL(meta.(*MSSQLConfiguration))
	if err != nil {
		return err
	}

	var stmt strings.Builder
	var stmtOpts strings.Builder
	checkPolicy := d.Get("check_policy")
	name := d.Get("name")

	var newpw interface{}
	if d.HasChange("password") {
		_, newpw = d.GetChange("password")
	}

	fmt.Fprintf(&stmt, "ALTER LOGIN %s WITH PASSWORD=N'%s'", name, newpw)

	if checkPolicy != "" {
		fmt.Fprintf(&stmtOpts, ", CHECK_POLICY=%s", checkPolicy)
	}

	fmt.Fprintf(&stmt, "%s;", stmtOpts.String())

	log.Print(stmt.String())

	s, err := db.Prepare(stmt.String())
	if err != nil {
		log.Fatal(err)
		return err
	}

	_, err = s.Exec()
	if err != nil {
		log.Fatal(err)
		return err
	}

	d.SetId(name.(string))

	defer db.Close()
	return nil
}

func resourceLoginDelete(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMSSQL(meta.(*MSSQLConfiguration))
	if err != nil {
		return err
	}

	var stmt strings.Builder

	name := d.Get("name")

	fmt.Fprintf(&stmt, "DROP LOGIN [%s];", name)

	log.Print(stmt.String())

	s, err := db.Prepare(stmt.String())
	if err != nil {
		log.Fatal(err)
		return err
	}

	_, err = s.Exec()
	if err != nil {
		log.Fatal(err)
		return err
	}

	d.SetId("")

	defer db.Close()
	return nil
}
