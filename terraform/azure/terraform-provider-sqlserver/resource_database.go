package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Database name",
				ForceNew:    true,
			},
			"containment": {
				Type:         schema.TypeString,
				Required:     false,
				Optional:     true,
				Description:  "CONTAINMENT parameter",
				ForceNew:     false,
				ValidateFunc: validateContainment,
			},
			"nested_triggers": {
				Type:         schema.TypeString,
				Required:     false,
				Optional:     true,
				Description:  "NESTED_TRIGGERS parameter",
				ForceNew:     false,
				ValidateFunc: validateOnOff,
			},
			"trustworthy": {
				Type:         schema.TypeString,
				Required:     false,
				Optional:     true,
				Description:  "TRUSTWORTHY parameter",
				ForceNew:     false,
				ValidateFunc: validateOnOff,
			},
			"default_language": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "DEFAULT_LANGUAGE parameter",
				ForceNew:    false,
			},
			"default_fulltext_language": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "DEFAULT_FULLTEXT_LANGUAGE parameter",
				ForceNew:    false,
			},
			"compatibility_level": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Description:  "COMPATIBILITY_LEVEL parameter (runs in a separate ALTER DATABASE statement)",
				ForceNew:     false,
				ValidateFunc: validateCompatibilityLevel,
			},
		},
		Read:   resourceDatabaseRead,
		Create: resourceDatabaseCreate,
		Update: resourceDatabaseUpdate,
		Delete: resourceDatabaseDelete,
	}
}

func resourceDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMSSQL(meta.(*MSSQLConfiguration))
	if err != nil {
		return err
	}

	var stmt strings.Builder
	var opts []string
	var hasOptions bool
	hasOptions = false

	containment := d.Get("containment")
	nestedTriggers := d.Get("nested_triggers")
	trustworthy := d.Get("trustworthy")
	defaultLanguage := d.Get("default_language")
	defaultFulltextLanguage := d.Get("default_fulltext_language")
	compatibilityLevel := d.Get("compatibility_level")
	name := d.Get("name")

	fmt.Fprintf(&stmt, "USE [master]; ")
	fmt.Fprintf(&stmt, "CREATE DATABASE [%s]", name)

	if containment != "" {
		fmt.Fprintf(&stmt, " CONTAINMENT=%s", containment)
	}

	if defaultFulltextLanguage != "" {
		hasOptions = true
		opts = append(opts, fmt.Sprintf("DEFAULT_FULLTEXT_LANGUAGE=%s", defaultFulltextLanguage))
	}

	if defaultLanguage != "" {
		hasOptions = true
		opts = append(opts, fmt.Sprintf("DEFAULT_LANGUAGE=%s", defaultLanguage))
	}

	if nestedTriggers != "" {
		hasOptions = true
		opts = append(opts, fmt.Sprintf("NESTED_TRIGGERS=%s", nestedTriggers))
	}

	if trustworthy != "" {
		hasOptions = true
		opts = append(opts, fmt.Sprintf("TRUSTWORTHY %s", trustworthy))
	}

	if hasOptions {
		fmt.Fprintf(&stmt, " WITH ")
	}
	fmt.Fprintf(&stmt, "%s", strings.Join(opts, ", "))
	fmt.Fprintf(&stmt, "; ")

	if compatibilityLevel != "" {
		fmt.Fprintf(&stmt, "ALTER DATABASE [%s] SET COMPATIBILITY_LEVEL=%d;", name, compatibilityLevel)
	}

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

func resourceDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMSSQL(meta.(*MSSQLConfiguration))
	resultFound := false
	if err != nil {
		return err
	}

	var stmt strings.Builder
	name := d.Get("name")

	fmt.Fprintf(&stmt, "USE [master]; ")
	fmt.Fprintf(&stmt, "SELECT name FROM sys.databases WHERE name = N'%s';", name)
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

func resourceDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMSSQL(meta.(*MSSQLConfiguration))
	var opts []string
	if err != nil {
		return err
	}

	var stmt strings.Builder
	var hasOptions bool
	hasOptions = false

	containment := d.Get("containment")
	nestedTriggers := d.Get("nested_triggers")
	trustworthy := d.Get("trustworthy")
	defaultLanguage := d.Get("default_language")
	defaultFulltextLanguage := d.Get("default_fulltext_language")
	compatibilityLevel := d.Get("compatibility_level")
	name := d.Get("name")

	fmt.Fprintf(&stmt, "USE [master]; ")
	fmt.Fprintf(&stmt, "ALTER DATABASE [%s]", name)

	if containment != "" {
		hasOptions = true
		opts = append(opts, fmt.Sprintf("CONTAINMENT=%s", containment))
	}

	if defaultFulltextLanguage != "" {
		hasOptions = true
		opts = append(opts, fmt.Sprintf("DEFAULT_FULLTEXT_LANGUAGE=%s", defaultFulltextLanguage))
	}

	if defaultLanguage != "" {
		hasOptions = true
		opts = append(opts, fmt.Sprintf("DEFAULT_LANGUAGE=%s", defaultLanguage))
	}

	if nestedTriggers != "" {
		hasOptions = true
		opts = append(opts, fmt.Sprintf("NESTED_TRIGGERS=%s", nestedTriggers))
	}

	if trustworthy != "" {
		hasOptions = true
		opts = append(opts, fmt.Sprintf("TRUSTWORTHY %s", trustworthy))
	}

	if hasOptions {
		fmt.Fprintf(&stmt, " SET ")
	}
	fmt.Fprintf(&stmt, "%s", strings.Join(opts, ", "))
	fmt.Fprintf(&stmt, "; ")

	if compatibilityLevel != "" {
		fmt.Fprintf(&stmt, "ALTER DATABASE [%s] SET COMPATIBILITY_LEVEL=%d;", name, compatibilityLevel)
	}

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

func resourceDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMSSQL(meta.(*MSSQLConfiguration))
	if err != nil {
		return err
	}

	var stmt strings.Builder

	name := d.Get("name")

	fmt.Fprintf(&stmt, "USE [master]; ")
	fmt.Fprintf(&stmt, "ALTER DATABASE [%s] SET SINGLE_USER WITH ROLLBACK IMMEDIATE;", name)
	fmt.Fprintf(&stmt, "DROP DATABASE [%s];", name)

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
