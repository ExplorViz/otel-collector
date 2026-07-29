# ExplorViz Parsing Processor

The ExplorViz parsing processor attempts to extract visualizable entities which are described by the incoming telemetry. Note that the term "entity" here describes a visualization entity for ExplorViz and does not necessarily match the [OTel definition](https://opentelemetry.io/docs/specs/otel/entities/) of entities. Examples for visualization entities include functions in code, HTTP endpoints, and databases.

If an entity can be successfully extracted, information about it is written into the telemetry attributes for downstream processing. Telemetry for which no entity can be derived is dropped. Additionally, the identifier of the ExplorViz landscape to which visualization information should be matched must be specified by the telemetry agent by writing the `explorviz.token.id` attribute (and the `explorviz.token.secret` value if landscape token validation is enabled).

For use with ExplorViz, this processor should be used in conjunction with the [ExplorViz exporter](../../exporter/explorvizexporter/README.md). Additionally, if you want to enable the validation of landscape tokens, the [ExplorViz token validator](../../extension/explorviztokenvalidator/README.md) extension must be present.

## Configuration

The processor optionally provides the following configuration options:

- `validate_tokens` (default = false) Whether to verify the existence and secret value of landscape tokens provided in the incoming telemetry. If set to true, telemetry with invalid token information is dropped. Requires the [token validator extension](../../extension/explorviztokenvalidator/README.md).

## Example

```yaml
processors:
    explorviz_parsing:
        validate_tokens: true
```
