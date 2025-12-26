package main.java.events.order;

import java.time.OffsetDateTime;
import java.util.UUID;

public class OrderCompletedPayload {
    public UUID orderId;
    public UUID userId;
    public OffsetDateTime timestamp;
}
