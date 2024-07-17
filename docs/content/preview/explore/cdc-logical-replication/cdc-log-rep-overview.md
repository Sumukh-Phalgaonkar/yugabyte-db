---
title: Overview of CDC logical replication internals
linkTitle: Overview
description: Change Data Capture in YugabyteDB using logical replication.
headcontent: Change Data Capture in YugabyteDB using logical replication
menu:
  preview:
    parent: explore-cdc-logical-replication
    identifier: cdc-log-rep-overview
    weight: 10
type: docs
---

## Concepts<a id="concepts"></a>

YugabyteDB’s logical replication feature makes use of many concepts like replication slot, publication, replica identity etc. from Postgres. Understanding these key concepts is crucial for setting up and managing a logical replication environment effectively.

- Replication slot

  - A replication slot represents a stream of changes that can be replayed to a client in the order they were made on the origin server. Each slot streams a sequence of changes from a single database. ([PG documentation](https://www.postgresql.org/docs/current/logicaldecoding-explanation.html#LOGICALDECODING-REPLICATION-SLOTS))


- Publication

  - A publication is a set of changes generated from a table or a group of tables, and might also be described as a change set or replication set. Each publication exists in only one database. ([PG documentation](https://www.postgresql.org/docs/current/logical-replication-publication.html#LOGICAL-REPLICATION-PUBLICATION))


- Output Plugins

  - Output plugins transform the data from the write-ahead log's internal representation into the format the consumer of a replication slot desires. ([PG documentation](https://www.postgresql.org/docs/current/logicaldecoding-output-plugin.html))


- LSN

  - An LSN (Log Sequence Number) is an unsigned 64-bit integer used to determine a position in WAL.


- Replica Identity

  - Replica identity is a table level parameter that can be used to control the amount of information being written to the change records. ([PG documentation](https://www.postgresql.org/docs/current/sql-altertable.html#SQL-ALTERTABLE-REPLICA-IDENTITY))


- Replication Protocols

  - \<Siddharth to give content>

The following sections explain each of the above concept in detail:

### REPLICATION SLOT

A Replication Slot represents a stream of changes that can be replayed to a client in the order they were made on the origin server. Each slot streams a sequence of changes from a single database.

A logical slot emits each change just once in normal operation. The current position of each slot is persisted only at checkpoint, so in the case of a crash the slot may return to an earlier LSN, which will then cause recent changes to be sent **again** when the server restarts. Logical decoding clients are responsible for avoiding ill effects from handling the same message more than once. Clients may wish to record the last LSN they saw when decoding and skip over any repeated data or (when using the replication protocol) request that decoding start from that LSN rather than letting the server determine the start point.


### PUBLICATION

A publication is a set of changes generated from a table or a group of tables, and might also be described as a change set or replication set. Each publication exists in only one database.

Publications are different from schemas and do not affect how the table is accessed. Each table can be added to multiple publications if needed. Publications may currently only contain tables. Objects must be added explicitly, except when a publication is created for ALL TABLES.


### OUTPUT PLUGIN

Output plugins are the components used to decode WAL changes and transform them into a specific format that can be consumed by replication clients. These plugins are notified about the change events that need to be processed and sent via various callbacks.

Yugabyte supports the following four output plugins:

- yboutput

- pgoutput

- test\_decoding

- wal2json

All these plugins are pre-packaged with yugabyte-db and do not require any external installation.


### LSN

### REPLICA IDENTITY

Replica identity is a table level parameter that controls the amount of information being written to the change records. Yugabyte supports the following four replica identities:

- CHANGE

- DEFAULT

- FULL

- NOTHING

The replica identity INDEX is not supported in Yugabyte. Replica identity CHANGE is the best performant and the default replica identity. The replica identity of a table can be changed by performing an alter table. However, for a given slot, the alter tables performed to change the replica identity after the creation of the slot will have no effect. This means that the Effective Replica Identity for any table for a slot, is the replica identity of the table that existed at the time of slot creation. A dynamically created table (a table created after slot creation) will have the default replica identity. For a replica identity modified after slot creation to take effect, a new slot will have to be created after performing the Alter table.

The tserver flag ysql\_yb\_default\_replica\_identity determines the default replica identity for user tables at the time of table creation. This flag has a default value CHANGE. The purpose of this flag is to set the replica identities for dynamically created tables. In order to create a dynamic table with desired replica identity, the flag must be set accordingly and then the table must be created. An advisory to users here is not to perform any alter replica identity on the dynamically created table for an interval of 5 minutes after creation.


### REPLICATION PROTOCOLS

\<Siddharth to give content>

{{< tip title="Explore" >}}
<!--TODO (Sumukh): Fix the Link to the getting started section. -->
See [Getting Started with Logical Replication](../../../explore/logical-replication/getting-started) in Explore to setup Logical Replication in YugabyteDB.

{{< /tip >}}

## Syntax and Semantics (Saurav)<a id="syntax-and-semantics-saurav"></a>

### Create Publication<a id="create-publication"></a>

    CREATE PUBLICATION name 
    [ FOR ALL TABLES
          | FOR publication_object [, ... ] ]

    where publication_object is:
        TABLE table_name 

    Parameters:
    name: The name of the new publication.
    FOR TABLE: Specifies a list of tables to add to the publication. 
    FOR ALL TABLES: Marks the publication as one that replicates changes for all tables in the database, including tables created in the future.

- CREATE PUBLICATION adds a new publication. The name of the publication should be unique in the database.

- If FOR TABLE or FOR ALL TABLES are not specified, then the publication starts out with an empty set of tables. That is useful if tables are to be added later.

- To create a publication, the invoking user must have the CREATE privilege for the current database. (Of course, superusers bypass this check.)

- To add a table to a publication, the invoking user must have ownership rights on the table. The FOR ALL TABLES clause requires the invoking user to be a superuser.

- Currently to publish a subset of operations (create, update, delete, truncate) via a Publication is not supported.


#### Examples<a id="examples"></a>

Create a publication that publishes all changes in two tables:

    CREATE PUBLICATION mypublication FOR TABLE users, departments;

Create a publication that publishes all changes in all tables:

    CREATE PUBLICATION alltables FOR ALL TABLES;


### Alter Publication<a id="alter-publication"></a>

    ALTER PUBLICATION name ADD publication_object [, ...]
    ALTER PUBLICATION name SET publication_object [, ...]
    ALTER PUBLICATION name DROP publication_object [, ...]
    ALTER PUBLICATION name OWNER TO { new_owner | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
    ALTER PUBLICATION name RENAME TO new_name

    where publication_object is one of:
    TABLE table_name

    Parameters:
    name: The name of an existing publication whose definition is to be altered.
    table_name: Name of an existing table. 
    new_owner: The user name of the new owner of the publication.
    new_name: The new name for the publication.

- The command ALTER PUBLICATION can change the attributes of a publication.

- The first three variants change which tables are part of the publication. The SET clause will replace the list of tables in the publication with the specified list; the existing tables that were present in the publication will be removed. The ADD and DROP clauses will add and remove one or more tables from the publication.

- The remaining variants change the owner and the name of the publication.

- You must own the publication to use ALTER PUBLICATION . Adding a table to a publication additionally requires owning that table. To alter the owner, you must also be a direct or indirect member of the new owning role. The new owner must have CREATE privilege on the database. Also, the new owner of a FOR ALL TABLES publication must be a superuser. However, a superuser can change the ownership of a publication regardless of these restrictions.


#### Examples<a id="examples-1"></a>

Add some tables to the publication:

    ALTER PUBLICATION my_publication ADD TABLE users, departments;


### Drop Publication<a id="drop-publication"></a>

    DROP PUBLICATION [ IF EXISTS ] name [, ...]

    Parameters:
    IF EXISTS: Do not throw an error if the publication does not exist. A notice is issued in this case.
    name: The name of an existing publication.

- DROP PUBLICATION removes an existing publication from the database.

- A publication can only be dropped by its owner or a superuser.


### Create Replication Slot<a id="create-replication-slot"></a>

    # Streaming Protocol
    CREATE_REPLICATION_SLOT slot_name LOGICAL output_plugin 
    [ 
          NOEXPORT_SNAPSHOT | USE_SNAPSHOT 
    ]
    [ WITH RECORD_TYPE record_type]

    # Function
    pg_create_logical_replication_slot(
      slot_name, 
      output_plugin,
      record_type
    )

    Parameters:
    slot_name: The name of the slot. It must be unique across all databases
    output_plugin: The name of the output plugin to be used. The only plugin that will be supported is 'yboutput'.
    record_type: This parameter determines the record structure of the data streamed to the client. The valid values are:
      * FULL
      * NOTHING
      * DEFAULT
      * CHANGE (default but configurable via the Tserver GFlag ysql_yb_default_replica_identity)

- A Replication Slot can be created either via the streaming protocol command or the standalone function shown above. The name of the replication slot should be unique across all databases.

- The output\_plugin parameter should be the name of a valid output plugin. Refer to the [Plugins section](https://docs.google.com/document/d/1fIYNd7ZNptBoSZVAk5HEalWwcxZ_AzOWt4ZHr7HHllE/edit#heading=h.hlc9eqfh38o2) for more.


#### Examples<a id="examples-2"></a>

Create a Replication Slot with name test\_replication\_slot and use the yboutput plugin.

    # Streaming Protocol
    CREATE_REPLICATION_SLOT test_replication_slot LOGICAL yboutput

    # Function
    pg_create_logical_replication_slot(
      'test_replication_slot', 
      'yboutput'
    )

Create a Replication Slot with name test\_replication\_slot and use the pgoutput plugin.

    # Streaming Protocol
    CREATE_REPLICATION_SLOT test_replication_slot LOGICAL yboutput

    # Function
    pg_create_logical_replication_slot(
      'test_replication_slot', 
      'pgoutput'
    )


### Drop Replication Slot<a id="drop-replication-slot"></a>

    # Streaming Protocol
    DROP_REPLICATION_SLOT slot_name

    # Function
    pg_drop_replication_slot(
      slot_name
    )

    Parameters:
    slot_name: The name of the slot to drop

- Removes a Replication Slot. A publication can only be dropped by its superuser or a user with replication privileges.

- PG also supports the `WAIT` option in the streaming protocol syntax. This is currently **unsupported** in YSQL.

- This command will **fail** if the replication slot is considered **active**. A replication slot is considered active if it has been polled in the last `ysql_cdc_active_replication_slot_window_ms` milliseconds. The default value of the flag is 5 minutes.


#### Examples<a id="examples-3"></a>

Drop an inactive replication slot:

    DROP_REPLICATION_SLOT inactive_replication_slot;


### START\_REPLICATION<a id="start_replication"></a>

Instructs server to start streaming WAL for logical replication, starting at WAL location XXX/XXX.

    START_REPLICATION SLOT slot_name LOGICAL XXX/XXX [ ( option_name [ option_value ] [, ...] ) ]

    SLOT slot_name: The name of the slot to stream changes from. This parameter is required, and must correspond to an existing logical replication slot created with CREATE_REPLICATION_SLOT in LOGICAL mode.

    XXX/XXX: The WAL location to begin streaming at.

    option_name: The name of an option passed to the slot's logical decoding plugin.

    option_value: Optional value, in the form of a string constant, associated with the specified option.

- The server can reply with an error, for example if the requested section of WAL has already been recycled. On success, the server responds with a CopyBothResponse message, and then starts to stream WAL to the frontend.

- The output plugin associated with the selected slot is used to process the output for streaming.


### ALTER REPLICA IDENTITY<a id="alter-replica-identity"></a>

    ALTER TABLE table_name REPLICA IDENTITY [replica_identity];

    Where replica_identity can be one of: CHANGE, DEFAULT, FULL, NOTHING


#### Examples<a id="examples-4"></a>

    ALTER TABLE users REPLICA IDENTITY FULL;

    ALTER TABLE departments REPLICA IDENTITY DEFAULT;
