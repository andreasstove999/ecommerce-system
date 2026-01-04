package domain;

import jakarta.persistence.*;
import java.time.OffsetDateTime;
import java.util.UUID;

@Entity
@Table(name = "shipments")
public class Shipment {

    @Id
    @Column(name = "shipping_id", nullable = false)
    public UUID shippingId;

    @Column(name = "order_id", nullable = false)
    public UUID orderId;

    @Column(name = "user_id", nullable = false)
    public UUID userId;

    @Embedded
    @AttributeOverrides({
            @AttributeOverride(name = "line1", column = @Column(name = "line1")),
            @AttributeOverride(name = "line2", column = @Column(name = "line2")),
            @AttributeOverride(name = "city", column = @Column(name = "city")),
            @AttributeOverride(name = "state", column = @Column(name = "state")),
            @AttributeOverride(name = "postalCode", column = @Column(name = "postal_code")),
            @AttributeOverride(name = "country", column = @Column(name = "country"))
    })
    public Address address;

    @Column(name = "shipping_method", nullable = false)
    public String shippingMethod;

    @Column(name = "carrier", nullable = false)
    public String carrier;

    @Column(name = "created_at", nullable = false)
    public OffsetDateTime createdAt;
}
