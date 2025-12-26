package main.java.domain;

import jakarta.persistence.*;
import java.time.OffsetDateTime;
import java.util.UUID;

@Entity
@Table(name = "processed_events")
public class ProcessedEvent {

    @Id
    public UUID eventId;

    @Column(nullable = false)
    public String eventName;

    @Column(nullable = false)
    public OffsetDateTime processedAt;
}
