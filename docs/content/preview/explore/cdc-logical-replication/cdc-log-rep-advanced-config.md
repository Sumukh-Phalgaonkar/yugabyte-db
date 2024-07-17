---
title: Advanced Configurations for Logical Replication
headerTitle: Config
linkTitle: Config
description: Advanced Configurations for Logical Replication.
headcontent: Advanced Configurations for Logical Replication.
menu:
  preview:
    parent: explore-cdc-logical-replication
    identifier: cdc-log-rep-advanced-config
    weight: 60
type: docs
---

## GFLAGS

The following tserver flags can be used to tune logical replication deployment configuration.

##### --ysql_yb_default_replica_identity

The default replica identity to be assigned to user defined tables at the time of creation. The flag is case sensitive and can take four possible values, `FULL`, `DEFAULT`,`'NOTHING` and `CHANGE`. If any value other than these is assigned to the flag, the replica identity `CHANGE` will be used as default at the time of table creation.

Default: `CHANGE`

##### --cdcsdk_enable_dynamic_table_support

This is a preview flag that can be used to switch the dynamic addition of tables ON or OFF.

Default: `false`

##### --cdcsdk_publication_list_refresh_interval_secs

Interval in seconds at which the table list in the publication will be refreshed.

Default: `3600`

##### --cdcsdk_vwal_getchanges_resp_max_size_bytes

Max size (in bytes) of GetChanges response for all GetChanges requests sent from Virtual WAL.

Default: `1 MB`

##### --cdcsdk_max_consistent_records

Controls the maximum number of records sent in GetConsistentChanges response.

Default: `500`

## Settings For Longer Retention

Anand to provide content.