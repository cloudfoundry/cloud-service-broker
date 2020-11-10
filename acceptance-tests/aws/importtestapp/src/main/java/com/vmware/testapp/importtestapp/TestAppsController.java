package com.vmware.testapp.importtestapp;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;
import com.google.gson.Gson;
import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import org.springframework.http.HttpStatus;
import java.io.IOException;
import java.sql.*;
@RestController
public class TestAppsController {
    
    String vcap_services = System.getenv("VCAP_SERVICES");




    @RequestMapping(value = "/testpostgres", method = RequestMethod.GET)
    public String testgcpbigquery() throws IOException {

        if (vcap_services != null && vcap_services.length() > 0) {
            try
            {
            
                JsonObject convertedObject = new Gson().fromJson(vcap_services, JsonObject.class);
                System.out.println(convertedObject);
                System.out.println("--------------------------------");

                JsonArray jsonArray = convertedObject.getAsJsonArray("csb-aws-postgresql");
                  
                  //System.out.println(jsonArray);
        
                  for(int i = 0 ; i < jsonArray.size() ; i++){

                      JsonObject creds = jsonArray.get(i).getAsJsonObject().get("credentials").getAsJsonObject();
                      System.out.println("--------------------------------");
                      String hostname = creds.get("hostname").getAsString();  
                      String jdbcUrl = creds.get("jdbcUrl").getAsString();
                      String name = creds.get("name").getAsString();  
                      String password = creds.get("password").getAsString();
                      String port = creds.get("port").getAsString();  
                      String username = creds.get("username").getAsString();

                      
                      System.out.println(hostname);

                      System.out.println(jdbcUrl);

                      Class.forName("org.postgresql.Driver");
                      Connection connection = DriverManager.getConnection(jdbcUrl, username, password);
                      System.out.println("Database Connected ..");
                      System.out.println("--------------------------------");

                  }
            }catch(Exception e){
                System.out.println(e.toString());
                throw new ResponseStatusException(
                  HttpStatus.NOT_FOUND, "Page Not Found", e);
            }
        }else{
            Exception e = new Exception("No Vcap Service");
            throw new ResponseStatusException(
                HttpStatus.NOT_FOUND, "Page Not Found", e );
        }
        return "";
    }
}
