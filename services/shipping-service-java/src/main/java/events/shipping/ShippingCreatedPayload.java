package main.java.events.shipping;

import java.time.OffsetDateTime;
import java.util.UUID;
import domain.Address;

public class ShippingCreatedPayload {
    public UUID shippingId;
    public UUID orderId;
    public UUID userId;

    public Address address;

    public String shippingMethod;
    public String carrier;

    public OffsetDateTime createdAt;
}
