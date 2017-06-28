# Kafka

The package provides functions to consume topics and to produce sync/async messages
to kafka. Minimal supported version of kafka is determined by [sarama](github.com/Shopify/sarama)
 library - Kafka 0.10 and 0.9, although older releases are still likely to work.

If you don't have kafka installed locally you can use docker image for testing:
 ```
sudo docker run -p 2181:2181 -p 9092:9092 --name kafka --rm \
 --env ADVERTISED_HOST=172.17.0.1 --env ADVERTISED_PORT=9092 spotify/kafka
```