package catalog.api;

import catalog.domain.Product;
import catalog.service.CatalogService;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import java.time.Instant;
import java.util.List;
import java.util.Objects;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(controllers = CatalogController.class)
class CatalogControllerTest {

        @Autowired
        MockMvc mvc;

        @Autowired
        ObjectMapper om;

        @MockBean
        CatalogService service;

        @Test
        void health_returnsOk() throws Exception {
                mvc.perform(get("/api/catalog/health"))
                                .andExpect(status().isOk())
                                .andExpect(jsonPath("$.status").value("ok"))
                                .andExpect(jsonPath("$.service").value("catalog-service"));
        }

        @Test
        void listProducts_usesDefaultsAndMapsResponse() throws Exception {
                var p = new Product(
                                UUID.randomUUID(),
                                "SKU-1",
                                "Name",
                                "Desc",
                                12.34,
                                "USD",
                                true,
                                Instant.parse("2025-01-01T00:00:00Z"),
                                Instant.parse("2025-01-01T00:00:00Z"));
                when(service.listProducts(50, 0)).thenReturn(List.of(p));

                mvc.perform(get("/api/catalog/products"))
                                .andExpect(status().isOk())
                                .andExpect(jsonPath("$[0].id").value(p.getId().toString()))
                                .andExpect(jsonPath("$[0].sku").value("SKU-1"))
                                .andExpect(jsonPath("$[0].name").value("Name"));

                verify(service).listProducts(50, 0);
        }

        @Test
        void getProduct_whenFound_returnsProduct() throws Exception {
                var id = UUID.randomUUID();
                var p = new Product(
                                id,
                                "SKU-2",
                                "Keyboard",
                                "",
                                99.99,
                                "USD",
                                true,
                                Instant.parse("2025-01-01T00:00:00Z"),
                                Instant.parse("2025-01-01T00:00:00Z"));
                when(service.getProduct(id)).thenReturn(Optional.of(p));

                mvc.perform(get("/api/catalog/products/{id}", id))
                                .andExpect(status().isOk())
                                .andExpect(jsonPath("$.id").value(id.toString()))
                                .andExpect(jsonPath("$.sku").value("SKU-2"));
        }

        @Test
        void getProduct_whenMissing_returns404() throws Exception {
                var id = UUID.randomUUID();
                when(service.getProduct(id)).thenReturn(Optional.empty());

                mvc.perform(get("/api/catalog/products/{id}", id))
                                .andExpect(status().isNotFound());
        }

        @Test
        void createProduct_validRequest_returns201_andPassesNormalizedProductToService() throws Exception {
                var created = new Product(
                                UUID.randomUUID(),
                                "SKU-3",
                                "Mouse",
                                "",
                                49.95,
                                "USD",
                                true,
                                Instant.parse("2025-01-01T00:00:00Z"),
                                Instant.parse("2025-01-01T00:00:00Z"));
                when(service.createProduct(any(Product.class))).thenReturn(created);

                var body = om.writeValueAsString(java.util.Map.of(
                                "sku", "SKU-3",
                                "name", "Mouse",
                                "price", 49.95));

                mvc.perform(post("/api/catalog/products")
                                .contentType(Objects.requireNonNull(MediaType.APPLICATION_JSON))
                                .content(Objects.requireNonNull(body)))
                                .andExpect(status().isCreated())
                                .andExpect(jsonPath("$.id").value(created.getId().toString()))
                                .andExpect(jsonPath("$.sku").value("SKU-3"))
                                .andExpect(jsonPath("$.name").value("Mouse"));

                var captor = ArgumentCaptor.forClass(Product.class);
                verify(service).createProduct(captor.capture());
                var passed = captor.getValue();
                assertThat(passed.getSku()).isEqualTo("SKU-3");
                assertThat(passed.getName()).isEqualTo("Mouse");
                assertThat(passed.getDescription()).isEqualTo("");
        }

        @Test
        void createProduct_missingSku_returns400() throws Exception {
                var body = om.writeValueAsString(java.util.Map.of(
                                "name", "Missing SKU",
                                "price", 10.0));

                mvc.perform(post("/api/catalog/products")
                                .contentType(Objects.requireNonNull(MediaType.APPLICATION_JSON))
                                .content(Objects.requireNonNull(body)))
                                .andExpect(status().isBadRequest());

                verifyNoInteractions(service);
        }

        @Test
        void createProduct_nonPositivePrice_returns400() throws Exception {
                var body = om.writeValueAsString(java.util.Map.of(
                                "sku", "SKU-X",
                                "name", "Bad price",
                                "price", 0));

                mvc.perform(post("/api/catalog/products")
                                .contentType(Objects.requireNonNull(MediaType.APPLICATION_JSON))
                                .content(Objects.requireNonNull(body)))
                                .andExpect(status().isBadRequest());

                verifyNoInteractions(service);
        }
}
