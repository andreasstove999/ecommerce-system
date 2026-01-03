package controller;

import static org.hamcrest.Matchers.is;
import static org.hamcrest.Matchers.notNullValue;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import domain.Address;
import domain.Shipment;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.UUID;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.web.servlet.MockMvc;
import repo.ShipmentRepository;

@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@AutoConfigureMockMvc
@ActiveProfiles("test")
class ShippingControllerIntegrationTest {

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private ShipmentRepository shipmentRepository;

    @BeforeEach
    void clean() {
        shipmentRepository.deleteAll();
    }

    @Test
    void getByIdReturnsPersistedShipment() throws Exception {
        Shipment shipment = buildShipment(UUID.randomUUID(), UUID.randomUUID());
        shipmentRepository.save(shipment);

        mockMvc.perform(get("/api/shipping/" + shipment.shippingId).accept(MediaType.APPLICATION_JSON))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.shippingId", is(shipment.shippingId.toString())))
                .andExpect(jsonPath("$.orderId", is(shipment.orderId.toString())))
                .andExpect(jsonPath("$.userId", is(shipment.userId.toString())))
                .andExpect(jsonPath("$.createdAt", notNullValue()));
    }

    @Test
    void getByOrderReturnsFirstShipmentForOrder() throws Exception {
        UUID orderId = UUID.randomUUID();
        Shipment first = buildShipment(UUID.randomUUID(), orderId);
        Shipment second = buildShipment(UUID.randomUUID(), orderId);
        shipmentRepository.save(first);
        shipmentRepository.save(second);

        mockMvc.perform(get("/api/shipping/by-order/" + orderId).accept(MediaType.APPLICATION_JSON))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.shippingId", is(first.shippingId.toString())))
                .andExpect(jsonPath("$.orderId", is(orderId.toString())));
    }

    @Test
    void getByIdReturnsNotFoundWhenMissing() throws Exception {
        mockMvc.perform(get("/api/shipping/" + UUID.randomUUID()).accept(MediaType.APPLICATION_JSON))
                .andExpect(status().isNotFound());
    }

    private Shipment buildShipment(UUID shippingId, UUID orderId) {
        Shipment shipment = new Shipment();
        shipment.shippingId = shippingId;
        shipment.orderId = orderId;
        shipment.userId = UUID.randomUUID();
        shipment.shippingMethod = "express";
        shipment.carrier = "DHL";
        shipment.createdAt = OffsetDateTime.of(2024, 1, 1, 12, 0, 0, 0, ZoneOffset.UTC);
        shipment.address = new Address();
        shipment.address.line1 = "123 Test St";
        shipment.address.city = "Test City";
        shipment.address.state = "TS";
        shipment.address.postalCode = "12345";
        shipment.address.country = "TS";
        return shipment;
    }
}
