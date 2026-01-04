package service;

import config.RabbitConfig;
import events.EventEnvelope;
import events.shipping.ShippingCreatedPayload;
import java.time.OffsetDateTime;
import java.util.UUID;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.stereotype.Component;

@Component
public class ShippingPublisher {

    private final RabbitTemplate rabbit;

    public ShippingPublisher(RabbitTemplate rabbit) {
        this.rabbit = rabbit;
    }

    public void publishShippingCreated(
            ShippingCreatedPayload payload,
            UUID correlationId,
            UUID causationId,
            long sequence) {
        EventEnvelope<ShippingCreatedPayload> env = new EventEnvelope<>();
        env.eventName = "ShippingCreated";
        env.eventVersion = 1;
        env.eventId = payload.shippingId; // common: reuse shippingId as eventId, ok for a starter
        env.correlationId = correlationId;
        env.causationId = causationId;
        env.producer = "shipping-service";
        env.partitionKey = payload.orderId.toString();
        env.sequence = sequence;
        env.occurredAt = OffsetDateTime.now().withOffsetSameInstant(java.time.ZoneOffset.UTC);
        env.schema = "contracts/events/shipping/ShippingCreated.v1.payload.schema.json";
        env.payload = payload;

        rabbit.convertAndSend(RabbitConfig.EVENTS_EXCHANGE, RabbitConfig.SHIPPING_CREATED_ROUTING_KEY, env);
    }
}
