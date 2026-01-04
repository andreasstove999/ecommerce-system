package domain;

import jakarta.persistence.*;
import java.time.OffsetDateTime;
import java.util.UUID;

@Entity
@Table(name = "processed_events")
public class ProcessedEvent {

    @Id
    @Column(name = "event_id", nullable = false)
    public UUID eventId;

    @Column(name = "event_name", nullable = false)
    public String eventName;

    @Column(name = "processed_at", nullable = false)
    public OffsetDateTime processedAt;
}
