package service;

import config.RabbitConfig;
import domain.Address;
import domain.ProcessedEvent;
import domain.Shipment;
import events.EventEnvelope;
import events.order.OrderCompletedPayload;
import events.shipping.ShippingCreatedPayload;
import repo.ProcessedEventRepository;
import repo.ShipmentRepository;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.Optional;
import java.util.UUID;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

@Component
public class ShippingEventHandler {

    private final ObjectMapper om;
    private final ShipmentRepository shipments;
    private final ProcessedEventRepository processed;
    private final ShippingPublisher publisher;
    private final SequenceService sequenceService;

    public ShippingEventHandler(
            ObjectMapper om,
            ShipmentRepository shipments,
            ProcessedEventRepository processed,
            ShippingPublisher publisher,
            SequenceService sequenceService) {
        this.om = om;
        this.shipments = shipments;
        this.processed = processed;
        this.publisher = publisher;
        this.sequenceService = sequenceService;
    }

    @RabbitListener(queues = RabbitConfig.ORDER_COMPLETED_QUEUE)
    @Transactional
    public void onOrderCompleted(byte[] body) throws Exception {
        EventEnvelope<OrderCompletedPayload> env = om.readValue(
                body,
                new TypeReference<EventEnvelope<OrderCompletedPayload>>() {
                });

        if (!"OrderCompleted".equals(env.eventName) || env.eventVersion != 1) {
            return;
        }

        // Idempotency by eventId
        if (env.eventId != null && processed.existsById(env.eventId)) {
            return;
        }

        UUID orderId = env.payload.orderId;
        UUID userId = env.payload.userId;

        // Extra safety: only one shipment per order
        Optional<Shipment> existing = shipments.findFirstByOrderId(orderId);
        if (existing.isPresent()) {
            markProcessed(env);
            return;
        }

        Address addr = new Address();
        addr.line1 = "123 Market St";
        addr.city = "Aarhus";
        addr.state = "DK";
        addr.postalCode = "8000";
        addr.country = "DK";

        UUID shippingId = UUID.randomUUID();
        OffsetDateTime now = OffsetDateTime.now(ZoneOffset.UTC);

        Shipment s = new Shipment();
        s.shippingId = shippingId;
        s.orderId = orderId;
        s.userId = userId;
        s.address = addr;
        s.shippingMethod = "standard";
        s.carrier = "PostNord";
        s.createdAt = now;

        shipments.save(s);

        ShippingCreatedPayload payload = new ShippingCreatedPayload();
        payload.shippingId = shippingId;
        payload.orderId = orderId;
        payload.userId = userId;
        payload.address = addr;
        payload.shippingMethod = s.shippingMethod;
        payload.carrier = s.carrier;
        payload.createdAt = now;

        String partitionKey = (env.partitionKey != null && !env.partitionKey.isBlank())
                ? env.partitionKey
                : orderId.toString();

        long seq = sequenceService.next(partitionKey);

        publisher.publishShippingCreated(payload, env.correlationId, env.eventId, seq);

        markProcessed(env);
    }

    private void markProcessed(EventEnvelope<?> env) {
        if (env.eventId == null)
            return;
        ProcessedEvent pe = new ProcessedEvent();
        pe.eventId = env.eventId;
        pe.eventName = env.eventName;
        pe.processedAt = OffsetDateTime.now(ZoneOffset.UTC);
        processed.save(pe);
    }
}
