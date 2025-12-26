package main.java.repo;

import main.java.domain.EventSequence;
import java.util.Optional;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Lock;
import org.springframework.stereotype.Repository;

import jakarta.persistence.LockModeType;

@Repository
public interface EventSequenceRepository extends JpaRepository<EventSequence, String> {

    @Lock(LockModeType.PESSIMISTIC_WRITE)
    Optional<EventSequence> findByPartitionKey(String partitionKey);
}
