package main.java.events;

import java.time.OffsetDateTime;
import java.util.UUID;

public class EventEnvelope<T> {
    public String eventName;
    public int eventVersion;
    public UUID eventId;

    public UUID correlationId;
    public UUID causationId;

    public String producer;
    public String partitionKey;
    public long sequence;

    public OffsetDateTime occurredAt;
    public String schema;

    public T payload;
}
