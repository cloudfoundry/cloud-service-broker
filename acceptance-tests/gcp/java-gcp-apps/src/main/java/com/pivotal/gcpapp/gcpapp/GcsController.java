package com.pivotal.vcapcreds.vcapcreds;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.nio.charset.Charset;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;

import org.springframework.http.HttpStatus;
import com.google.cloud.spanner.TransactionRunner;
import com.google.cloud.spanner.Spanner;
import com.google.cloud.spanner.SpannerOptions;
import com.google.cloud.storage.BlobId;
import com.google.cloud.storage.BlobInfo;
import com.google.gson.Gson;
import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import com.google.auth.oauth2.GoogleCredentials;
import com.google.cloud.storage.Storage;
import com.google.cloud.storage.StorageOptions;
import com.google.api.gax.paging.Page;
import com.google.cloud.storage.Bucket;
import com.google.cloud.spanner.DatabaseClient;
import com.google.cloud.spanner.DatabaseId;
import com.google.cloud.spanner.Mutation;
import com.google.cloud.spanner.TransactionContext;

import com.google.cloud.bigquery.BigQuery;
import com.google.cloud.bigquery.BigQuery.QueryOption;
import com.google.cloud.bigquery.BigQuery.QueryResultsOption;
import com.google.cloud.bigquery.BigQueryError;
import com.google.cloud.bigquery.BigQueryOptions;
import com.google.cloud.bigquery.TableId;
import com.google.cloud.bigquery.TableInfo;

import com.google.cloud.bigquery.Field;
import com.google.cloud.bigquery.FieldValueList;
import com.google.cloud.bigquery.Schema;
import com.google.cloud.bigquery.StandardTableDefinition;
import com.google.cloud.bigquery.LegacySQLTypeName;

@RestController
public class GcsController {

  
  private JsonObject vcapServicesObject;

  private GoogleCredentials googleCredentials;
  private String projectId;
  private String serviceInstanceId;
  
  String vcap_services = System.getenv("VCAP_SERVICES");




  @RequestMapping(value = "/testgcpbigquery", method = RequestMethod.GET)
  public String testgcpbigquery() throws IOException {
    

          if (vcap_services != null && vcap_services.length() > 0) {
            try
            {
            
                JsonObject convertedObject = new Gson().fromJson(vcap_services, JsonObject.class);
                //System.out.println(convertedObject);
                System.out.println("--------------------------------");
                JsonArray jsonArray = convertedObject.getAsJsonArray("csb-google-bigquery");
                  
                  //System.out.println(jsonArray);
        
                  for(int i = 0 ; i < jsonArray.size() ; i++){

                      JsonObject creds = jsonArray.get(i).getAsJsonObject().get("credentials").getAsJsonObject();
                      String ProjectId = creds.get("ProjectId").getAsString();  
                      String dataset_id = creds.get("dataset_id").getAsString();
                      
                
                    
                      System.out.println("ProjectId = " + ProjectId);
                      System.out.println("bucket_name = " + dataset_id);
                     

                      String googleCredentialStr = creds.get("Credentials").getAsString();

                      // System.out.println(googleCredentialStr);
                      try {
                        googleCredentials = GoogleCredentials.fromStream(new ByteArrayInputStream(googleCredentialStr.getBytes()));
                        System.out.println("googleCredentials = " + googleCredentials);

                        BigQuery bigquery = BigQueryOptions.newBuilder().setCredentials(googleCredentials).build().getService();


                        TableId tableId = TableId.of(dataset_id, "Table_csb_created");

                        Field stringField = Field.of("StringField", LegacySQLTypeName.STRING);
                        // Table schema definition
                        Schema schema = Schema.of(stringField);
                        // Create a table
                        StandardTableDefinition tableDefinition = StandardTableDefinition.of(schema);
                        bigquery.create(TableInfo.of(tableId, tableDefinition));

                      } catch (IOException e) {
                        System.out.println(e.toString());
                        throw new ResponseStatusException(
                          HttpStatus.NOT_FOUND, "Page Not Found", e);
                      }


                  
                  
                  }


                
                System.out.println("--------------------------------");
              



          

            }  catch (Exception e)
            {
                System.out.println(e.toString());
                throw new ResponseStatusException(
                  HttpStatus.NOT_FOUND, "Page Not Found", e);
            }
      }
return "";
}
  
  @RequestMapping(value = "/testgcpspanner", method = RequestMethod.GET)
  public String testgcpspanner() throws IOException {
    

          if (vcap_services != null && vcap_services.length() > 0) {
            try
            {
            
                JsonObject convertedObject = new Gson().fromJson(vcap_services, JsonObject.class);
                //System.out.println(convertedObject);
                System.out.println("--------------------------------");
                JsonArray jsonArray = convertedObject.getAsJsonArray("csb-google-spanner");
                  
                  //System.out.println(jsonArray);
        
                  for(int i = 0 ; i < jsonArray.size() ; i++){

                      JsonObject creds = jsonArray.get(i).getAsJsonObject().get("credentials").getAsJsonObject();
                      String ProjectId = creds.get("ProjectId").getAsString();  
                      String db_name = creds.get("db_name").getAsString();
                      
                      String instance = creds.get("instance").getAsString();
                    
                      System.out.println("ProjectId = " + ProjectId);
                      System.out.println("bucket_name = " + db_name);
                      System.out.println("serviceInstanceId = " + instance);
                    

                      String googleCredentialStr = creds.get("Credentials").getAsString();

                      // System.out.println(googleCredentialStr);
                      try {
                        googleCredentials = GoogleCredentials.fromStream(new ByteArrayInputStream(googleCredentialStr.getBytes()));
                        System.out.println("googleCredentials = " + googleCredentials);

                        SpannerOptions options = SpannerOptions.newBuilder().setCredentials(googleCredentials).setProjectId(ProjectId).build();
                        Spanner spanner = options.getService();
                        DatabaseClient dbClient = spanner.getDatabaseClient(DatabaseId.of(options.getProjectId(), instance, db_name));

                        // dbClient.readWriteTransaction()
                        // .run(new TransactionRunner.TransactionCallable<Void>() {
                        //     public Void run(TransactionContext transaction) throws Exception {
                        //         transaction.buffer(Mutation.newInsertBuilder("table")
                        //                 .set("key")
                        //                 .to("value")
                        //                 .set("value1")
                        //                 .to(213123)
                        //                 .set("value2")
                        //                 .to("value")
                        //                 .build());
                        //         return null;
                        //     }
                        // });

                      } catch (IOException e) {
                        System.out.println(e.toString());
                        throw new ResponseStatusException(
                          HttpStatus.NOT_FOUND, "Page Not Found", e);
                      }


                  
                  
                  }


                
                System.out.println("--------------------------------");
              



          

            }  catch (Exception e)
            {
                System.out.println(e.toString());
                throw new ResponseStatusException(
                  HttpStatus.NOT_FOUND, "Page Not Found", e);
            }
      }
return "";
}

  
  @RequestMapping(value = "/testgcpbucket", method = RequestMethod.GET)
  public String testgcpstorage() throws IOException {

    //System.out.println("********************************");
    //System.out.println("vcap_services = " + vcap_services);
    //System.out.println("********************************");

    if (vcap_services != null && vcap_services.length() > 0) {
          try
          {
          
              JsonObject convertedObject = new Gson().fromJson(vcap_services, JsonObject.class);
              System.out.println("--------------------------------");
              //System.out.println("jsonArray = " + convertedObject);
              JsonArray jsonArray = convertedObject.getAsJsonArray("csb-google-storage-bucket");
                
                
       
                for(int i = 0 ; i < jsonArray.size() ; i++){

                    JsonObject creds = jsonArray.get(i).getAsJsonObject().get("credentials").getAsJsonObject();
                    String ProjectId = creds.get("ProjectId").getAsString();
                    String bucket_name = creds.get("bucket_name").getAsString();
                    String serviceInstanceId = creds.get("Name").getAsString();
                  
                    System.out.println("ProjectId = " + ProjectId);
                    System.out.println("bucket_name = " + bucket_name);
                    System.out.println("serviceInstanceId = " + serviceInstanceId);
                  

                    String googleCredentialStr = creds.get("Credentials").getAsString();

                    // System.out.println(googleCredentialStr);
                    try {
                      googleCredentials = GoogleCredentials.fromStream(new ByteArrayInputStream(googleCredentialStr.getBytes()));
                      System.out.println("googleCredentials = " + googleCredentials);

                      //Storage storage  = StorageOptions.newBuilder().setCredentials(googleCredentials).build().getService();
                      Storage storage = StorageOptions.newBuilder().setProjectId(projectId).setCredentials(googleCredentials).build().getService();

                      BlobInfo blobInfo = BlobInfo.newBuilder(BlobId.of(bucket_name, "testapp.txt")).build();
                      String data="Test this data";
                      storage.create(blobInfo, data.getBytes());

    
                    } catch (IOException e) {
                      System.out.println(e.toString());
                      throw new ResponseStatusException(
                        HttpStatus.NOT_FOUND, "Page Not Found", e);
                    }


                
                
                }


              
              System.out.println("--------------------------------");
             



        

          }  catch (Exception e)
          {
              System.out.println(e.toString());
              throw new ResponseStatusException(
                HttpStatus.NOT_FOUND, "Page Not Found", e);
          }
    }
    return "";
  }


}