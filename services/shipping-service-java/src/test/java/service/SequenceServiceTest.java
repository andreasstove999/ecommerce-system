package service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import domain.EventSequence;
import java.util.Optional;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Captor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import repo.EventSequenceRepository;

@ExtendWith(MockitoExtension.class)
class SequenceServiceTest {

    @Mock
    private EventSequenceRepository repo;

    @InjectMocks
    private SequenceService sequenceService;

    @Captor
    private ArgumentCaptor<EventSequence> sequenceCaptor;

    @BeforeEach
    void setup() {
        sequenceCaptor = ArgumentCaptor.forClass(EventSequence.class);
    }

    @Test
    void nextCreatesNewSequenceWhenMissing() {
        when(repo.findByPartitionKey("orders-1")).thenReturn(Optional.empty());
        when(repo.save(any(EventSequence.class))).thenAnswer(invocation -> invocation.getArgument(0));

        long next = sequenceService.next("orders-1");

        assertThat(next).isEqualTo(1L);
        verify(repo).save(sequenceCaptor.capture());
        EventSequence saved = sequenceCaptor.getValue();
        assertThat(saved.partitionKey).isEqualTo("orders-1");
        assertThat(saved.nextSequence).isEqualTo(2L);
    }

    @Test
    void nextIncrementsExistingSequence() {
        EventSequence existing = new EventSequence();
        existing.partitionKey = "orders-2";
        existing.nextSequence = 5L;

        when(repo.findByPartitionKey(existing.partitionKey)).thenReturn(Optional.of(existing));
        when(repo.save(any(EventSequence.class))).thenAnswer(invocation -> invocation.getArgument(0));

        long next = sequenceService.next(existing.partitionKey);

        assertThat(next).isEqualTo(5L);
        verify(repo).save(sequenceCaptor.capture());
        EventSequence updated = sequenceCaptor.getValue();
        assertThat(updated.partitionKey).isEqualTo(existing.partitionKey);
        assertThat(updated.nextSequence).isEqualTo(6L);
    }
}
