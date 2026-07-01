# ExplorViz Token Listener Extension

This extension listens for landscape token events from the [user-service](https://github.com/ExplorViz/user-service). The token events are consumed from a Kafka topic. The extension keeps track of all currently valid landscape tokens and their secret values in-memory. Other Collector components can access the token store to validate incoming telemetry.

Interaction with Kafka is implemented using the [franz-go](https://github.com/twmb/franz-go) library. For serialization, we use [Protobuf](https://protobuf.dev/getting-started/gotutorial/).

## Configuration

The extension optionally provides the following configuration options:

- `broker` (default = localhost:9091) Network endpoint of the Kafka broker to use
- `topic` (default = tokens.events) Kafka topic to consume landscape token events from

## Example

```yaml
extensions:
    explorviz_token_validator:
        broker: kafka-hostname:9092
        topic: your_topic_name
```
