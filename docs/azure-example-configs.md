# Some Config Examples with Notes

## Azure *csb-azure-mssql-db*

Example with default location and resource group, two SQL server instances, and three plans. The plans use DTU skus and the *standard-S3-server1* plan creates a DB on *sql-server1* from the *server_credentials* list.

### Config File
```yaml
provision:
  defaults: '{
    "location": "westus2", 
    "resource_group": "eb-test-rg1"
  }'

service:
  csb-azure-mssql-db:
    provision:
      defaults: '{ 
        "server_credentials": { 
          "sql-server1": { 
            "server_name":"csb-azsql-svr-b2d43b57-9396-4a8c-8592-6696e7b1d84d", 
            "admin_username":"TIrtZNKlGQEhmOwR", 
            "admin_password":"lSFMJ..PoD3H_wZ2cNLNgn9uTBwWskYkMzBkN6mN5A1ZL.V6t0qrebkYeyDYYnW7", 
            "server_resource_group":"eb-test-rg1" 
          }, 
          "sql-server2": { 
            "server_name":"csb-azsql-svr-dc6f6028-2c01-4d70-b6e6-81ddaaf6b56a", 
            "admin_username":"UomUxvtkVQxtkGKy", 
            "admin_password":"At76iTk0o6HkNfR1ZrNCrOZ6wZIWz~QECrp7H-U63.uH8JA-cWpFZaG_C.2MXaEm", 
            "server_resource_group":"eb-test-rg1" 
          }
        }
      }'
    plans: '[
      {
        "id":"881de5d9-e078-44e7-bed5-26faadabda3c",
        "name":"standard-S0",
        "description":"DTU: S0 - 10DTUS, 250GB storage",      
        "sku_name":"S0"
      },
      {
        "id":"1a1de5d9-e078-44e7-bed5-266aadabdaa6",
        "name":"premium-P1",
        "description":"DTU: P1 - 125DTUS, 500GB storage",      
        "sku_name":"P1"
      },
      {
        "id":"1a1de5d9-e079-44e7-bed5-266aadabdaa6",
        "name":"standard-S3-server1",
        "description":"Server1 DB - DTU: S3 - 100, 250GB storage",      
        "sku_name":"S3",
        "server":"sql-server1"
      }
    ]'
```

With this config file example, `cf marketplace` should include

```
csb-azure-mssql-db  small, medium, large, extra-large, standard-S0, premium-P1, standard-S3-server1   Manage Azure SQL Databases on pre-provisioned database servers
```

### Provision Examples

Create a *small* db instance on *sql-server2* 
```
cf create-service csb-azure-mssql-db small small-db -c '{"server":"sql-server2"}'
```

Create a *premium-P1* on *sql-server1*
```
cf create-service csb-azure-mssql-db premium-P1 p1-db -c '{"server":"sql-server1"}'
```

Create *standard-S3-server1*
```
cf create-service csb-azure-mssql-db standard-S3-server1 s3-db-server1
```

## Azure *csb-azure-mssql-db-failover-group*

Example with default location and resource group and two fail over group server pairs.

### Config File
```yaml
provision:
  defaults: '{
    "location": "westus2", 
    "resource_group": "eb-test-rg1"
  }'

service:
  csb-azure-mssql-db-failover-group:
    provision:
      defaults: '{
        "read_write_endpoint_failover_policy": "Manual",    
        "server_credential_pairs": { 
          "pair1": { 
            "admin_username": "anadminuser",         
            "admin_password": "This_S0uld-3eC0mplex~", 
            "primary": { 
              "server_name": "mssql-server-p-18342",        
              "resource_group": "rg-test-service-18342" 
            }, 
            "secondary": { 
              "server_name": "mssql-server-s-18342",         
              "resource_group": "rg-test-service-18342" 
            } 
          }, 
          "pair2": { 
            "admin_username": "foo", 
            "admin_password": "bar",         
            "primary": { 
              "server_name": "s1",
              "resource_group": "rg" 
            }, 
            "secondary": { 
              "server_name": "s2",         
              "resource_group": "rg" 
            }
          } 
        }
    }'
```

### Provision Examples

Create a *small* db instance on *pair2* 

```
cf create-service csb-azure-mssql-db-failover-group small db-small -c '{"server_pair":"pair2"}'
```

Create a DTU P1 instance on *pair1*
```
cf create-service csb-azure-mssql-db-failover-group small db-P1 -c '{"server_pair":"pair2","sku_name":"P1"}'
```


