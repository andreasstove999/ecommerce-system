package main.java.controller;

import main.java.domain.Shipment;
import main.java.repo.ShipmentRepository;
import java.util.UUID;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/shipping")
public class ShippingController {

    private final ShipmentRepository repo;

    public ShippingController(ShipmentRepository repo) {
        this.repo = repo;
    }

    @GetMapping("/{shippingId}")
    public ResponseEntity<Shipment> getById(@PathVariable UUID shippingId) {
        return repo.findById(shippingId)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/by-order/{orderId}")
    public ResponseEntity<Shipment> getByOrder(@PathVariable UUID orderId) {
        return repo.findFirstByOrderId(orderId)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }
}
