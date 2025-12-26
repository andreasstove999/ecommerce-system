package domain;

import jakarta.persistence.*;

@Entity
@Table(name = "event_sequences")
public class EventSequence {

    @Id
    @Column(name = "partition_key", nullable = false, length = 200)
    public String partitionKey;

    @Column(name = "next_sequence", nullable = false)
    public long nextSequence;
}
