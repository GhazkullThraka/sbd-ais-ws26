# Exercise 7

## Part 1:

With Architechture A, writing Parquet files to LAKE and updating a separate state file are two independent operations. If a job stops while writing or after writing, before save_state() completes, the data is already visible in LAKE while the state file still holds the old offsets, causing duplicates on the next run. Updating the state file first causes missing records instead.

Architecture C keeps readers accurate using DEDUP_SQL, but pushes deduplication cost onto every query, letting duplicate files accumulate on disk after crashes, and breaks any reader that queries the raw Parquet files directly.

Therefore, I chose Architecture B, storing no separate state file, writing each batch to STAGING, and moving it atomically into LAKE. Architecture B relies on atomic directory renames, Cloud storage like Amazon S3 doesn't have real folders, so a crash halfway through would leave half a batch visible. In the cloud, formats like Delta Lake or Apache Iceberg solve this by keeping a small log file that lists which Parquet files officially belong to the lake, and updating that log in one final step after all files are written.

### Implementation:
```python
def run_once():
    # 1. where did the last run stop?
    start = lake_position()
    # 2. read what is new
    new = read_new(start)
    # 3. write it
    run = write_files(new, STAGING)
    # 4. remember where you stopped (if your architecture needs it)
    publish(run)
```
### Results:
```
situation                                 in lake  duplicates  missing  not loaded    run  check
-----------------------------------------------------------------------------------------------
1 first run                             2,700,000           0        0           0  23.4s  PASS
2 new records arrive, run again         2,800,000           0        0           0   4.8s  PASS
3 nothing new, run again                2,800,000           0        0           0   0.2s  PASS
4 run stops WHILE writing: right after  2,800,000           0        0     100,000         PASS
  ... then the next run                 2,900,000           0        0           0   4.2s  PASS
5 run stops AFTER writing: right after  2,900,000           0        0     100,000         PASS
  ... then the next run                 3,000,000           0        0           0   4.4s  PASS
-----------------------------------------------------------------------------------------------
Exactly once after every run: YES
```

A job that copies everything fails on the second run because read_new(None) re-reads the entire Kafka topic from offset 0 and writes directly into LAKE, duplicating every record. 

In (check_all) All five situations pass with 0 duplicates and 0 missing records

lake_position() starts at None which loads all 2,700,000 records across 4 partitions into 4 Parquet files, then reads only the 100,000 new records (+4 files, 4.8s). When nothing new has arrived, the job finishes in 0.2s without adding new files. And in both failure tests, incomplete or unpublished files stay in STAGING. Immediately after the crash, LAKE remains consistent (0 duplicates, 0 missing, 100,000 not yet loaded, and no partial files added to LAKE). The next run reads lake_position() from the untouched lake and loads the remaining 100,000 records.




