package main.java.domain;

import jakarta.persistence.*;
import java.time.OffsetDateTime;
import java.util.UUID;

@Entity
@Table(name = "shipments")
public class Shipment {

    @Id
    public UUID shippingId;

    @Column(nullable = false)
    public UUID orderId;

    @Column(nullable = false)
    public UUID userId;

    @Embedded
    public Address address;

    @Column(nullable = false)
    public String shippingMethod;

    @Column(nullable = false)
    public String carrier;

    @Column(nullable = false)
    public OffsetDateTime createdAt;
}
