package catalog.integration;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.web.client.TestRestTemplate;
import org.springframework.boot.test.web.server.LocalServerPort;
import org.springframework.http.*;
import org.springframework.test.context.DynamicPropertyRegistry;
import org.springframework.test.context.DynamicPropertySource;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;

@Testcontainers
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
class CatalogServiceIT {

    @Container
    @SuppressWarnings("resource") // Testcontainers manages lifecycle automatically
    static final PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:16-alpine")
            .withDatabaseName("catalog_db")
            .withUsername("catalog_user")
            .withPassword("catalog_pass");

    @DynamicPropertySource
    static void registerProps(DynamicPropertyRegistry r) {
        r.add("spring.datasource.url", postgres::getJdbcUrl);
        r.add("spring.datasource.username", postgres::getUsername);
        r.add("spring.datasource.password", postgres::getPassword);
    }

    @LocalServerPort
    int port;

    @Autowired
    TestRestTemplate http;

    @Autowired
    ObjectMapper om;

    private String url(String path) {
        return "http://localhost:" + port + path;
    }

    @Test
    void health_isOk() {
        var res = http.getForEntity(url("/api/catalog/health"), String.class);
        assertThat(res.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(res.getBody()).contains("\"status\":\"ok\"");
    }

    @Test
    void createThenGetThenList_roundtripWithRealPostgres() throws Exception {
        var sku = "SKU-" + UUID.randomUUID();

        var createJson = om.writeValueAsString(java.util.Map.of(
                "sku", sku,
                "name", "Gaming Keyboard",
                "price", 129.99));

        var createRes = http.exchange(
                url("/api/catalog/products"),
                HttpMethod.POST,
                new HttpEntity<>(createJson, jsonHeaders()),
                String.class);

        assertThat(createRes.getStatusCode()).isEqualTo(HttpStatus.CREATED);
        JsonNode created = om.readTree(createRes.getBody());
        assertThat(created.get("id").asText()).isNotBlank();
        assertThat(created.get("sku").asText()).isEqualTo(sku);
        assertThat(created.get("currency").asText()).isEqualTo("USD");
        assertThat(created.get("description").asText()).isEqualTo("");

        var id = created.get("id").asText();

        var getRes = http.getForEntity(url("/api/catalog/products/" + id), String.class);
        assertThat(getRes.getStatusCode()).isEqualTo(HttpStatus.OK);
        JsonNode got = om.readTree(getRes.getBody());
        assertThat(got.get("id").asText()).isEqualTo(id);
        assertThat(got.get("sku").asText()).isEqualTo(sku);

        var listRes = http.getForEntity(url("/api/catalog/products?limit=50&offset=0"), String.class);
        assertThat(listRes.getStatusCode()).isEqualTo(HttpStatus.OK);
        JsonNode arr = om.readTree(listRes.getBody());
        assertThat(arr.isArray()).isTrue();
        assertThat(arr.toString()).contains(sku);
    }

    @Test
    void create_duplicateSku_returns409() throws Exception {
        var sku = "SKU-DUP-" + UUID.randomUUID();
        var json = om.writeValueAsString(java.util.Map.of(
                "sku", sku,
                "name", "Same SKU",
                "price", 10.0));

        var first = http.exchange(
                url("/api/catalog/products"),
                HttpMethod.POST,
                new HttpEntity<>(json, jsonHeaders()),
                String.class);
        assertThat(first.getStatusCode()).isEqualTo(HttpStatus.CREATED);

        var second = http.exchange(
                url("/api/catalog/products"),
                HttpMethod.POST,
                new HttpEntity<>(json, jsonHeaders()),
                String.class);

        assertThat(second.getStatusCode()).isEqualTo(HttpStatus.CONFLICT);
    }

    @Test
    void get_missingProduct_returns404() {
        var res = http.getForEntity(url("/api/catalog/products/" + UUID.randomUUID()), String.class);
        assertThat(res.getStatusCode()).isEqualTo(HttpStatus.NOT_FOUND);
    }

    @Test
    void create_invalidPayload_returns400() throws Exception {
        var json = om.writeValueAsString(java.util.Map.of(
                "name", "No SKU",
                "price", 10.0));

        var res = http.exchange(
                url("/api/catalog/products"),
                HttpMethod.POST,
                new HttpEntity<>(json, jsonHeaders()),
                String.class);

        assertThat(res.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
    }

    private static HttpHeaders jsonHeaders() {
        var h = new HttpHeaders();
        h.setContentType(MediaType.APPLICATION_JSON);
        return h;
    }
}
