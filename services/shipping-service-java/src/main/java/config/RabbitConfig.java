package config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import java.util.Map;
import org.springframework.amqp.core.*;
import org.springframework.amqp.rabbit.connection.ConnectionFactory;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.amqp.rabbit.config.SimpleRabbitListenerContainerFactory;
import org.springframework.amqp.support.converter.Jackson2JsonMessageConverter;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.lang.NonNull;

@Configuration
public class RabbitConfig {

    public static final String ORDER_COMPLETED_QUEUE = "order.completed";
    public static final String SHIPPING_CREATED_QUEUE = "shipping.created";

    // DLQ
    public static final String DLX_NAME = "shipping-service.dlx";
    public static final String DLQ_NAME = "shipping-service.dlq";
    public static final String DLQ_ROUTING_KEY = "order.completed.dlq";

    @Bean
    public DirectExchange shippingDlx() {
        return new DirectExchange(DLX_NAME, true, false);
    }

    @Bean
    public Queue shippingDlq() {
        return QueueBuilder.durable(DLQ_NAME).build();
    }

    @Bean
    public Binding shippingDlqBinding(Queue shippingDlq, DirectExchange shippingDlx) {
        return BindingBuilder.bind(shippingDlq).to(shippingDlx).with(DLQ_ROUTING_KEY);
    }

    @Bean
    public Queue orderCompletedQueue() {
        // When the listener rejects (or fails and is configured to not requeue),
        // message goes to DLQ
        return QueueBuilder.durable(ORDER_COMPLETED_QUEUE)
                .withArguments(Map.of(
                        "x-dead-letter-exchange", DLX_NAME,
                        "x-dead-letter-routing-key", DLQ_ROUTING_KEY))
                .build();
    }

    @Bean
    public Queue shippingCreatedQueue() {
        return new Queue(SHIPPING_CREATED_QUEUE, true);
    }

    @Bean
    public ObjectMapper objectMapper() {
        ObjectMapper om = new ObjectMapper();
        om.registerModule(new JavaTimeModule());
        return om;
    }

    @Bean
    public Jackson2JsonMessageConverter messageConverter(@NonNull ObjectMapper om) {
        return new Jackson2JsonMessageConverter(om);
    }

    @Bean
    public RabbitTemplate rabbitTemplate(@NonNull ConnectionFactory cf,
            @NonNull Jackson2JsonMessageConverter converter) {
        RabbitTemplate t = new RabbitTemplate(cf);
        t.setMessageConverter(converter);
        return t;
    }

    /**
     * Critical for DLQ behavior:
     * - If a listener throws, and requeue is false, Rabbit will dead-letter it
     * (DLX->DLQ).
     */
    @Bean
    public SimpleRabbitListenerContainerFactory rabbitListenerContainerFactory(
            @NonNull ConnectionFactory cf,
            @NonNull Jackson2JsonMessageConverter converter) {
        SimpleRabbitListenerContainerFactory factory = new SimpleRabbitListenerContainerFactory();
        factory.setConnectionFactory(cf);
        factory.setMessageConverter(converter);

        // Don’t requeue poison messages endlessly; dead-letter them.
        factory.setDefaultRequeueRejected(false);

        return factory;
    }
}
