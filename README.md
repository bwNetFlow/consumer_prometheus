# consumer_prometheus

This is a basic consumer that grabs flow data from a Kafka topic and exposes
some stats using the Prometheus metric format.  

## Usage

The simplest call could look like this, which would start the consumer with TLS encryption and SASL auth enabled.  
Exposed Stats are available via port `8080` to prometheus at `/metrics` and `/flowdata` endpoints.

```bash
./consumer_prometheus \
  --kafka.brokers=kafka.local:9092 \
  --kafka.topic=flows-enriched \
  --kafka.consumer_group=test-consumer-group \
  --kafka.user=username \
  --kafka.pass=ultraSecretPassword
```

Also check out our [demo](https://github.com/bwNetFlow/demo) for more examples.

### Additional options

Note that these values are set to `false` by default.  
TLS and SASL auth can also be deactivated with:

```bash
  --kafka.disable_auth=true \
  --kafka.disable_tls=true \
```
There is also an option to connect as ```anon``` user to your kafka instance.

```bash
  --kafka.auth_anon=true
```