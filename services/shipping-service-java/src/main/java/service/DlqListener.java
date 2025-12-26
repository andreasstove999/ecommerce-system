package main.java.service;

import main.java.config.RabbitConfig;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.stereotype.Component;

@Component
public class DlqListener {
    private static final Logger log = LoggerFactory.getLogger(DlqListener.class);

    @RabbitListener(queues = RabbitConfig.DLQ_NAME)
    public void onDlqMessage(byte[] body) {
        log.error("DLQ message received ({} bytes). Payload: {}", body.length, new String(body));
    }
}
