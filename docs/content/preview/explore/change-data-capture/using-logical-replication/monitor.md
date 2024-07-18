---
title: CDC monitoring in YugabyteDB
headerTitle: Monitor
linkTitle: Monitor
description: Monitor Change Data Capture in YugabyteDB.
menu:
  preview:
    parent: explore-change-data-capture-logical-replication
    identifier: monitor
    weight: 30
type: docs
---

## Status of the deployed connector

You can use the rest APIs to monitor your deployed connectors. The following operations are available:

* List all connectors

   ```sh
   curl -X GET localhost:8083/connectors/
   ```

* Get a connector's configuration

   ```sh
   curl -X GET localhost:8083/connectors/<connector-name>
   ```

* Get the status of all tasks with their configuration

   ```sh
   curl -X GET localhost:8083/connectors/<connector-name>/tasks
   ```

* Get the status of the specified task

   ```sh
   curl -X GET localhost:8083/connectors/<connector-name>/tasks/<task-id>
   ```

* Get the connector's status, and the status of its tasks

   ```sh
   curl -X GET localhost:8083/connectors/<connector-name>/status
   ```

## Metrics

### CDC Service metrics

Provide information about CDC service in YugabyteDB.

| Metric name | Type | Description |
| :---- | :---- | :---- |
| cdcsdk_change_event_count | `long` | The Change Event Count metric shows the number of records sent by the CDC Service. |
| cdcsdk_traffic_sent | `long` | The number of milliseconds since the connector has read and processed the most recent event. |
| cdcsdk_event_lag_micros | `long` | The LAG metric is calculated by subtracting the timestamp of the latest record in the WAL of a tablet from the last record sent to the CDC connector. |
| cdcsdk_expiry_time_ms | `long` | The time left to read records from WAL is tracked by the Stream Expiry Time (ms). |

In addition to the built-in support for JMX metrics that Zookeeper, Kafka, and Kafka Connect provide, the Debezium YugabyteDB connector provides the following types of metrics.

### Snapshot metrics

The **MBean** is `debezium.postgres:type=connector-metrics,context=snapshot,server=<topic.prefix>`.

Snapshot metrics are not exposed unless a snapshot operation is active, or if a snapshot has occurred since the last connector start.

The following table lists the shapshot metrics that are available.

| Attributes | Type | Description |
| :--------- | :--- | :---------- |
| `LastEvent` | string | The last snapshot event that the connector has read. |
| `MilliSecondsSinceLastEvent` | long | The number of milliseconds since the connector has read and processed the most recent event. |
| `TotalNumberOfEventsSeen` | long | The total number of events that this connector has seen since last started or reset. |
| `NumberOfEventsFiltered` | long | The number of events that have been filtered by include/exclude list filtering rules configured on the connector. |
| `CapturedTables` | string[] | The list of tables that are captured by the connector. |
| `QueueTotalCapacity` | int | The length the queue used to pass events between the snapshotter and the main Kafka Connect loop. |
| `QueueRemainingCapacity` | int | The free capacity of the queue used to pass events between the snapshotter and the main Kafka Connect loop. |
| `TotalTableCount` | int | The total number of tables that are being included in the snapshot. |
| `RemainingTableCount` | int | The number of tables that the snapshot has yet to copy. |
| `SnapshotRunning` | boolean | Whether the snapshot was started. |
| `SnapshotPaused` | boolean | Whether the snapshot was paused. |
| `SnapshotAborted` | boolean | Whether the snapshot was aborted. |
| `SnapshotCompleted` | boolean | Whether the snapshot completed. |
| `SnapshotDurationInSeconds` | long | The total number of seconds that the snapshot has taken so far, even if not complete. Includes also time when snapshot was paused. |
| `SnapshotPausedDurationInSeconds` | long | The total number of seconds that the snapshot was paused. If the snapshot was paused several times, the paused time adds up. |
| `RowsScanned` | Map<String, Long> | Map containing the number of rows scanned for each table in the snapshot. Tables are incrementally added to the Map during processing. Updates every 10,000 rows scanned and upon completing a table. |
| `MaxQueueSizeInBytes` | long | The maximum buffer of the queue in bytes. This metric is available if `max.queue.size.in.bytes` is set to a positive long value. |
| `CurrentQueueSizeInBytes` | long | The current volume, in bytes, of records in the queue. |

The connector also provides the following additional snapshot metrics when an incremental snapshot is executed:

| Attributes | Type | Description |
| :--------- | :--- | :---------- |
| `ChunkId` | string | The identifier of the current snapshot chunk. |
| `ChunkFrom` | string | The lower bound of the primary key set defining the current chunk. |
| `ChunkTo` | string | The upper bound of the primary key set defining the current chunk. |
| `TableFrom` | string | The lower bound of the primary key set of the currently snapshotted table. |
| `TableTo` | string | The upper bound of the primary key set of the currently snapshotted table. |

### Streaming metrics

The **MBean** is `debezium.postgres:type=connector-metrics,context=streaming,server=<topic.prefix>`.

The following table lists the streaming metrics that are available.

| Attributes | Type | Description |
| :--------- | :--- | :---------- |
| `LastEvent` | string | The last streaming event that the connector has read. |
| `MilliSecondsSinceLastEvent` | long | The number of milliseconds since the connector has read and processed the most recent event. |
| `TotalNumberOfEventsSeen` | long | The total number of events that this connector has seen since the last start or metrics reset. |
| `TotalNumberOfCreateEventsSeen` | long | The total number of create events that this connector has seen since the last start or metrics reset. |
| `TotalNumberOfUpdateEventsSeen` | long | The total number of update events that this connector has seen since the last start or metrics reset. |
| `TotalNumberOfDeleteEventsSeen` | long | The total number of delete events that this connector has seen since the last start or metrics reset. |
| `NumberOfEventsFiltered` | long | The number of events that have been filtered by include/exclude list filtering rules configured on the connector. |
| `CapturedTables` | string[] | The list of tables that are captured by the connector. |
| `QueueTotalCapacity` | int | The length the queue used to pass events between the streamer and the main Kafka Connect loop. |
| `QueueRemainingCapacity` | int | The free capacity of the queue used to pass events between the streamer and the main Kafka Connect loop. |
| `Connected` | boolean | Flag that denotes whether the connector is currently connected to the database server. |
| `MilliSecondsBehindSource` | long | The number of milliseconds between the last change event’s timestamp and the connector processing it. The values will incoporate any differences between the clocks on the machines where the database server and the connector are running. |
| `NumberOfCommittedTransactions` | long | The number of processed transactions that were committed. |
| `SourceEventPosition` | Map<String, String> | The coordinates of the last received event. |
| `LastTransactionId` | string | Transaction identifier of the last processed transaction. |
| `MaxQueueSizeInBytes` | long | The maximum buffer of the queue in bytes. This metric is available if `max.queue.size.in.bytes` is set to a positive long value. |
| `CurrentQueueSizeInBytes` | long | The current volume, in bytes, of records in the queue. |