package controller;

import domain.Shipment;
import repo.ShipmentRepository;
import java.util.UUID;
import java.util.Map;
import org.springframework.lang.NonNull;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/shipping")
public class ShippingController {

    private final ShipmentRepository repo;

    public ShippingController(ShipmentRepository repo) {
        this.repo = repo;
    }

    @GetMapping("/health")
    public ResponseEntity<Map<String, String>> health() {
        return ResponseEntity.ok(Map.of("status", "ok", "service", "shipping-service"));
    }

    @GetMapping("/{shippingId}")
    public ResponseEntity<Shipment> getById(@PathVariable @NonNull UUID shippingId) {
        return repo.findById(shippingId)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/by-order/{orderId}")
    public ResponseEntity<Shipment> getByOrder(@PathVariable @NonNull UUID orderId) {
        return repo.findFirstByOrderId(orderId)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }
}
