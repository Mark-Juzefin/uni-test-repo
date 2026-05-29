# TODO

## Messaging: outbox → worker → queue → notifications
- [x] Outbox table + write event row in the SAME tx as product create/delete
- [x] Publisher interface + LogPublisher stub (worker has no real queue yet)
- [x] Outbox worker: poll unpublished rows → Publisher.Publish → set published_at
- [x] Broker: Kafka (Redpanda in docker, `localhost:19092`) + KafkaPublisher
- [x] Notifications service: consume from `product-events` topic, log messages

## Required by task
- [x] Unit tests: create / list / delete
- [x] README with run instructions

## Bonus
- [ ] Prometheus metrics on create/delete
