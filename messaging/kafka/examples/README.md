# Kafkaclient tools

These too have been adapted from [Sarama](http://github.com/Shopify/sarama) to work with this Kafka client.

This folder contains applications that are useful for exploration of your Kafka cluster, or instrumentation.
Some of these tools mirror the tools that ship with Kafka, but these tools won't require installing the JVM to function.

- [kafka-syncproducer](./syncproducer): a command line tool to send messages via an sync producer.
- [kafka-asyncproducer](./asyncproducer): a command line tool to send messages via an async producer.
- [kafka-consumer](./consumer): a command line tool to the consume messages of a topic on your Kafka cluster.
