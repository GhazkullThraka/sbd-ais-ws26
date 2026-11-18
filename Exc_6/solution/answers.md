## Which option do you choose for this scenario, and why? What does it cost, and how would you notice in production before it hurts? Which option would you choose if your team owned the payment path?

For this scenario, I'd choose log-based Change Data Capture because it is the only option that satisfies both R1 and R2 without modifying the certified payment path or adding query load to the system of record. Polling would miss payments that commit out of ID order and add query load for every reader, while dual writes and the outbox pattern both require prohibited changes to card_processosor with dual writes also risking lost or invented events if a crash occurs. By reading Postgres' write-ahead log directly, Debezium guarantees that every committed payment reaches Kafka in exact commit order. The main operational cost is the PostgreSQL replication slot, which retains all unread WAL files; if the connector goes down for an extended period, the disk can fill up and bring down both the database and the payment path. To notice this in production before it hurts, the team should set up automated alerts on replication slot lag, retained WAL disk usage, and Kafka Connect connector health so operators can intervene long before storage runs out. 

If my team owned the payment path, I would choose the outbox pattern instead. Writing the payment and an outbox row in the same database transaction delivers the same transactional completeness for the small cost of one extra row per transaction, while letting the application explicitly define the events it publishes.

### Run 1 - Debezium's Defaults

* **Prediction:** Only one velocity instance will process payments while the other two stay idle, causing lag to grow and R4 to fail.
* **Velocity scoreboard:**
  ```text
  R3  frauds caught          1/8  (12%)   FAIL  (>= 90 %)
  R3  false alerts           0   PASS  (at most 10 % of the flagged transactions)
  R4  alert latency p50/p95  98.3 s / 104.3 s   FAIL  (p95 <= 5 s after the commit)

Debezium creates topics with 1 partition by default, and within a consumer group each partition can only be read by a single instance. That left vel-2 and vel-3 idle while vel-1 fell behind at ~1,000 payments/s, pushing p95 latency to 104.3 s and leaving most card-testing frauds stuck in the backlog.

### Run 2 - More Partitions

* **Prediction:** With 6 partitions, all three velocity instances will share the workload evenly, eliminating the backlog so that both R4 and R3 pass.
* **Velocity scoreboard:**
  ```text
    R3  frauds caught          1/7  (14%)   FAIL  (>= 90 %)
    R3  false alerts           0   PASS  (at most 10 % of the flagged transactions)
    R4  alert latency p50/p95  4.5 s / 4.8 s   PASS  (p95 <= 5 s after the commit)

With 6 partitions, each of the three velocity instances received 2 partitions and kept up with the traffic, so R4 passed. However, Debezium still keyed events by the table's primary key, which hashed the 12 test payments of a single stolen card across all six partitions. Because each instance only saw about a third of those payments, none reached the threshold of 8 payments in 60 seconds.

### Run 3 - Cardholder as the Key
* **Prediction:** The event key will change to {"user_id": ...}, sending all of a cardholder's payments to the same partition and instance. Because of this, both R3 and R4 should pass, maybe with a few false alerts from normal shoppers.

* **Velocity Scoreboard:**
```text
    R3  frauds caught          7/7  (100%)   PASS  (>= 90 %)
    R3  false alerts           0   PASS  (at most 10 % of the flagged transactions)
    R4  alert latency p50/p95  1.1 s / 2.3 s   PASS  (p95 <= 5 s after the commit)
```
Setting the key to user_id sent all of a cardholder's payments to the same partition and the same velocity instance in order. All three instances split the 6 partitions evenly, catching 100% of the frauds with a p95 latency of 2.3s.

## In Closing

If nobody had checked Debezium's two defaults, nothing would have crashed or thrown an error. With 1 partition, consumer lag would have grown continuously under load while extra instances sat idle; with the primary key as the Kafka key, the velocity model would have quietly missed almost all card-testing attacks, only surfacing weeks later as a spike in chargebacks. 

The team could have caught the partition issue early by monitoring consumer group lag and per-instance partition assignment/throughput, and caught the keying issue by running shadow evaluations against labeled fraud/chargeback data or asserting in staging that all events for a test cardholder land on a single partition.