package com.example.demo;

import java.io.IOException;
import java.nio.charset.Charset;
import org.springframework.cloud.gcp.storage.GoogleStorageResource;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.io.Resource;
import org.springframework.util.StreamUtils;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;
import com.google.cloud.storage.Storage;
import java.io.OutputStream;
import org.springframework.core.io.WritableResource;
import org.springframework.web.bind.annotation.RequestBody;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import com.google.cloud.ReadChannel;
import java.nio.ByteBuffer;
@RestController
public class GcsController {


  @Autowired
  private Storage storage;

  @RequestMapping(value = "/", method = RequestMethod.GET)
  public String readGcsFile() throws IOException {
   
    try (ReadChannel channel = storage.reader("svennela-bucket-account-test", "testapp.txt")) {
        ByteBuffer bytes = ByteBuffer.allocate(64 * 1024);
        while (channel.read(bytes) > 0) {
            bytes.flip();
            String data = new String(bytes.array(), 0, bytes.limit());
            bytes.clear();
            return data;
        }
    }
    return "";
  }

  @RequestMapping(value = "/", method = RequestMethod.POST)
  String writeGcs(@RequestBody String data) throws IOException {
    
    String gcsFileLocation = "gs://svennela-bucket-account-test/testapp.txt";

    Resource resource = new GoogleStorageResource(this.storage, gcsFileLocation);
        try (OutputStream os = ((WritableResource) resource).getOutputStream()) {
            os.write(data.getBytes());
        }

    return "file was updated\n";
  }
}