package com.example.springbootdynamodb;

import java.io.IOException;

import com.google.gson.JsonObject;
import com.google.gson.Gson;
import com.google.gson.JsonArray;

import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;

import com.amazonaws.auth.AWSStaticCredentialsProvider;
import com.amazonaws.auth.BasicAWSCredentials;
import com.amazonaws.regions.Regions;
import com.amazonaws.services.dynamodbv2.AmazonDynamoDB;
import com.amazonaws.services.dynamodbv2.AmazonDynamoDBClientBuilder;
import com.amazonaws.services.dynamodbv2.datamodeling.DynamoDBMapper;
import com.amazonaws.services.dynamodbv2.datamodeling.DynamoDBMapperConfig;
import com.amazonaws.services.dynamodbv2.datamodeling.DynamoDBQueryExpression;
import com.amazonaws.services.dynamodbv2.datamodeling.DynamoDBSaveExpression;
import com.amazonaws.services.dynamodbv2.datamodeling.DynamoDBQueryExpression;
import com.amazonaws.services.dynamodbv2.datamodeling.DynamoDBSaveExpression;
import com.amazonaws.services.dynamodbv2.model.AttributeValue;
import com.amazonaws.services.dynamodbv2.model.ConditionalCheckFailedException;
import com.amazonaws.services.dynamodbv2.model.ExpectedAttributeValue;
import com.amazonaws.services.dynamodbv2.model.ReturnConsumedCapacity;
import com.amazonaws.services.dynamodbv2.model.ReturnValue;
import com.amazonaws.services.dynamodbv2.model.UpdateItemRequest;

@RestController
public class DynamodbController {
    private JsonObject vcapServicesObject;

    String vcap_services = System.getenv("VCAP_SERVICES");

    Customer customer;
    @RequestMapping(value = "/testdynamodb", method = RequestMethod.GET)
    public String testdynamodb() throws IOException {

        if (vcap_services != null && vcap_services.length() > 0) {
            try
            {

                 JsonObject convertedObject = new Gson().fromJson(vcap_services, JsonObject.class);
                //System.out.println(convertedObject);
                System.out.println("--------------------------------");
                JsonArray jsonArray = convertedObject.getAsJsonArray("csb-aws-dynamodb");
                  
                  //System.out.println(jsonArray);
        
                  for(int i = 0 ; i < jsonArray.size() ; i++){

                      JsonObject creds = jsonArray.get(i).getAsJsonObject().get("credentials").getAsJsonObject();
                      String access_key_id = creds.get("access_key_id").getAsString();  
                      String secret_access_key = creds.get("secret_access_key").getAsString(); 
                      String region = creds.get("region").getAsString(); 
                      String dynamodb_table_id = creds.get("dynamodb_table_id").getAsString(); 
                      String dynamodb_table_name = creds.get("dynamodb_table_name").getAsString(); 

                      AmazonDynamoDB client = AmazonDynamoDBClientBuilder.standard()
                      .withCredentials(new AWSStaticCredentialsProvider(new BasicAWSCredentials(access_key_id, secret_access_key)))
                      .withRegion(region)
                      .build();
                      
                      DynamoDBMapper mapper = new DynamoDBMapper(client, DynamoDBMapperConfig.DEFAULT);
                      try {
                        customer = new Customer("TestName",30);

                        System.out.println("Started - Inserting customer with customerId={}");
                        var saveIfNotExistsExpression = new DynamoDBSaveExpression()
                            .withExpectedEntry("PK", new ExpectedAttributeValue(false))
                            .withExpectedEntry("SK", new ExpectedAttributeValue(false));
                        mapper.save(customer, saveIfNotExistsExpression);
                        System.out.println("Completed - Inserting customer with customerId={}");
                    } catch (ConditionalCheckFailedException ce) {
                        throw new UniquePrimaryKeyConstraintViolationException("Failed to insert customer because an item with " +
                            "this PK and SK already exists; " + customer);
                    } catch (Exception e) {
                        System.out.println(e);
                        throw new ResponseStatusException(
                                HttpStatus.NOT_FOUND, "Page Not Found", e);
                    }

                  }
            }
           
        catch (Exception e)
        {
            System.out.println(e.toString());
            throw new ResponseStatusException(
              HttpStatus.NOT_FOUND, "Page Not Found", e);
        }
      
    }
    return "";
}

}
