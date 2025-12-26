package service;

import domain.EventSequence;
import repo.EventSequenceRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class SequenceService {

    private final EventSequenceRepository repo;

    public SequenceService(EventSequenceRepository repo) {
        this.repo = repo;
    }

    /**
     * Returns the next sequence number for this partition key, starting at 1.
     * Monotonic per partition_key.
     */
    @Transactional
    public long next(String partitionKey) {
        EventSequence es = repo.findByPartitionKey(partitionKey)
                .orElseGet(() -> {
                    EventSequence n = new EventSequence();
                    n.partitionKey = partitionKey;
                    n.nextSequence = 1L;
                    return n;
                });

        long current = es.nextSequence;
        es.nextSequence = current + 1;
        repo.save(es);

        return current;
    }
}
