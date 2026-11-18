## Which option do you choose for this scenario, and why? What does it cost, and how would you notice in production before it hurts? Which option would you choose if your team owned the payment path?

For this scenario, I'd choose log-based Change Data Capture because it is the only option that satisfies both R1 and R2 without modifying the certified payment path or adding query load to the system of record. Polling would miss payments that commit out of ID order and add query load for every reader, while dual writes and the outbox pattern both require prohibited changes to card_processosor with dual writes also risking lost or invented events if a crash occurs. By reading Postgres' write-ahead log directly, Debezium guarantees that every committed payment reaches Kafka in exact commit order. The main operational cost is the PostgreSQL replication slot, which retains all unread WAL files; if the connector goes down for an extended period, the disk can fill up and bring down both the database and the payment path. To notice this in production before it hurts, the team should set up automated alerts on replication slot lag, retained WAL disk usage, and Kafka Connect connector health so operators can intervene long before storage runs out. 

If my team owned the payment path, I would choose the outbox pattern instead. Writing the payment and an outbox row in the same database transaction delivers the same transactional completeness for the small cost of one extra row per transaction, while letting the application explicitly define the events it publishes.

# Exercise 6 — Card Sentinel

## Part 1 — Architectural Decision

For this scenario, I choose **log-based Change Data Capture (CDC)** because it is the only option that satisfies both R1 and R2 without modifying the certified payment path or adding query load to the system of record. Polling would miss payments that commit out of ID order and add query load for every reader, while dual writes and the outbox pattern both require prohibited changes to `card_processor.py`—with dual writes also risking lost or invented events if a crash occurs. By reading PostgreSQL's write-ahead log (WAL) directly, Debezium guarantees that every committed payment reaches Kafka in exact commit order. The main operational cost is the PostgreSQL replication slot, which retains all unread WAL files; if the connector goes down for an extended period, the disk can fill up and bring down both the database and the payment path. To notice this in production before it hurts, the team should set up automated alerts on replication slot lag, retained WAL disk usage, and Kafka Connect connector health so operators can intervene long before storage runs out. If my team owned the payment path, I would choose the **outbox pattern** instead. Writing the payment and an outbox row in the same database transaction delivers the same transactional completeness for the small cost of one extra row per transaction, while letting the application explicitly define the 

### Run 1 — Debezium's Defaults

* **Prediction:** Only one velocity instance will process payments while the other two stay idle, causing lag to grow and R4 to fail.
* **Velocity scoreboard:**
  ```text
  R3  frauds caught          1/8  (12%)   FAIL  (>= 90 %)
  R3  false alerts           0   PASS  (at most 10 % of the flagged transactions)
  R4  alert latency p50/p95  98.3 s / 104.3 s   FAIL  (p95 <= 5 s after the commit)

Debezium creates topics with 1 partition by default, and within a consumer group each partition can only be read by a single instance. That left vel-2 and vel-3 idle while vel-1 fell behind at ~1,000 payments/s, pushing p95 latency to 104.3 s and leaving most card-testing frauds stuck in the backlog.

