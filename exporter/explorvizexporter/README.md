# ExplorViz Exporter

The ExplorViz exporter reads entities previously extracted by the [ExplorViz parsing processor](../../processor/explorvizparsingprocessor/README.md) from the telemetry data and exports them to a Kafka topic to be consumed by the [landscape-service](https://github.com/ExplorViz/landscape-service). Any telemetry not containing the relevant attributes is dropped.

Interaction with Kafka is implemented using the [franz-go](https://github.com/twmb/franz-go) library. For serialization, we use [Protobuf](https://protobuf.dev/getting-started/gotutorial/).

## Configuration

The exporter optionally provides the following configuration options:

- `broker` (default = localhost:9091) Network endpoint of the Kafka broker to use
- `topic` (default = telemetry.entities) Kafka topic to produce parsing results into
- `metrics`
    - `produce_generic_entities` (default = true): Whether to allow the metrics pipeline to produce generic entities, i.e. entities that cannot be further classified based on their attributes. Depending on the number of instrumented scopes, these can produce a lot of clutter in the visualization while potentially never being involved in any communication.
- `logs`
    - `produce_generic_entities` (default = true): Whether to allow the logs pipeline to produce generic entities, i.e. entities that cannot be further classified based on their attributes. Depending on the number of instrumented scopes, these can produce a lot of clutter in the visualization while potentially never being involved in any communication.

## Example

```yaml
exporters:
    explorviz:
        broker: kafka-hostname:9092
        topic: your_topic_name
        metrics:
            produce_generic_entities: true
        logs:
            produce_generic_entities: true
```
