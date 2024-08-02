# MySQL replica setup

## Setup

Infra required for this exercise is already setup in the `docker-compose.yml` file.
We have 3 mysql containers running in the same network.
One of them is the master and the other two are replicas.

Some volumes are attached to the containers to persist the data.

## Configuration

For configuration, we need to assign unique server-id to each mysql instance, enable binary logging and create a
dedicated user for replication.

### To setup dedicated user for replication

1. connect to master node

```bash
mysql -u root -proot
```

2. Create a dedicated user for replication

```sql
CREATE USER 'repl'@'%' IDENTIFIED BY 'repl'; -- creates a user with username repl and password repl
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';
```

### To setup server id and binary logging on master and replica nodes

Create a file `my.cnf` in `./data-master/etc/mysql` directory with the following content:

```ini
[mysqld]
bind-address = mysql-master
server-id = 2
log_bin = /var/log/mysql/mysql-bin.log
```

Create a file `my.cnf` in `./data-replica1/etc/mysql` directory:

```ini
[mysqld]
server-id = 21
log_bin = /var/log/mysql/mysql-bin.log
relay-log = /var/log/mysql/mysql-relay-bin.log
```

Create a file `my.cnf` in `./data-replica2/etc/mysql` directory:

```ini
[mysqld]
server-id = 21
log_bin = /var/log/mysql/mysql-bin.log
relay-log = /var/log/mysql/mysql-relay-bin.log
```

After making above changes, restart the mysql containers.

Verify the configuration by connecting to the mysql instances and running the following commands:

```sql
SHOW VARIABLES LIKE 'server_id'; -- check server id
SHOW VARIABLES LIKE 'log_bin'; -- binary logging enabled or not
```

Now that the configuration is done, we can proceed with the replication setup.

## Setting up replication

### On master node

Get the binlog coordinates from master node and record them.

```
FLUSH TABLES WITH READ LOCK;
SHOW BINARY LOG STATUS\G
```

```
example output:
*************************** 1. row ***************************
             File: mysql-bin.000001
         Position: 691
     Binlog_Do_DB: 
 Binlog_Ignore_DB: 
Executed_Gtid_Set: 
1 row in set (0.00 sec)
```

Unlock the tables

```sql
UNLOCK TABLES;
```

### On replica nodes

Connect to the replica nodes and run the following command to configure replication source.

```
CHANGE REPLICATION SOURCE TO
SOURCE_HOST='mysql-master',
SOURCE_PORT=3306,
SOURCE_USER='repl',
SOURCE_PASSWORD='repl',
SOURCE_LOG_FILE='mysql-bin.000001',
SOURCE_LOG_POS=691,
GET_SOURCE_PUBLIC_KEY=1;
```

After configuring the replication source, start the replication.

```
START REPLICA;
```

Verify the replication status by running the following command:

```
SHOW REPLICA STATUS\G
```

If there are no errors in output of above command, then the replication is setup successfully.
To see it in action, create a table and add some data to it in master node. The same data should be replicated to the
replica nodes.

Example:

```sql
CREATE DATABASE test;
USE test;
CREATE TABLE test_table
(
    id   INT PRIMARY KEY,
    name VARCHAR(50)
);
INSERT INTO test_table
VALUES (1, 'test');
```

# References:

- [DigitalOcean](https://www.digitalocean.com/community/tutorials/how-to-set-up-replication-in-mysql)
- [MySQL](https://dev.mysql.com/doc/refman/8.4/en/replication-configuration.html)
