package catalog;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class CatalogServiceApplication {
  public static void main(String[] args) {

    // TODO: This is yet to be tested for frontend and gateway.

    // TODO: seed some products for the fontend to show.
    SpringApplication.run(CatalogServiceApplication.class, args);
  }
}
