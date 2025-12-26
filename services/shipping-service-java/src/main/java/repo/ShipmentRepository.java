package main.java.repo;

import main.java.domain.Shipment;
import java.util.Optional;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface ShipmentRepository extends JpaRepository<Shipment, UUID> {
    Optional<Shipment> findFirstByOrderId(UUID orderId);
}
