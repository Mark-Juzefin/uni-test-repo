# TODO

## Messaging: outbox → worker → queue → notifications
- [x] Outbox table + write event row in the SAME tx as product create/delete
- [ ] Publisher interface + LogPublisher stub (worker has no real queue yet)
- [ ] Outbox worker: poll unpublished rows → Publisher.Publish → set published_at
- [ ] Queue (SQS via LocalStack, pin `localstack/localstack:3.8`) + real SQSPublisher
- [ ] Notifications service: consume from queue, log messages

## Required by task
- [ ] Unit tests: create / list / delete
- [ ] README with run instructions

## Bonus
- [ ] Prometheus metrics on create/delete
