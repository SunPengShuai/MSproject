## 操作主数据库
进入mysql-master

`docker exec -it mysql-master /bin/bash`

登录mysql

`mysql -u root -proot`

创建mysql-slave接入的用户名和密码

`CREATE USER 'slave'@'%' IDENTIFIED WITH 'mysql_native_password' BY '123456';`

为该用户授权

`GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'slave'@'%';`

查看主数据库的二进制日志
```sql
mysql> SHOW BINARY LOG STATUS;
+-----------------------+----------+--------------+------------------+-------------------+
| File                  | Position | Binlog_Do_DB | Binlog_Ignore_DB | Executed_Gtid_Set |
+-----------------------+----------+--------------+------------------+-------------------+
| mall-mysql-bin.000005 |      158 |              | mysql            |                   |
+-----------------------+----------+--------------+------------------+-------------------+
1 row in set (0.01 sec)
```

记住File和Position，这是主数据库的二进制日志的位置


## 操作从数据库

进入mysql-slave

`docker exec -it mysql-master /bin/bash`

登录mysql

`mysql -u root -proot`

修改下面这条指令的文件SOURCE_LOG_FILE属性和位置SOURCE_LOG_POS属性，要和上表的一致

`CHANGE REPLICATION SOURCE TO SOURCE_HOST='mysql-master', SOURCE_USER='slave', SOURCE_PASSWORD='123456', SOURCE_PORT=3306, SOURCE_LOG_FILE='mall-mysql-bin.000005', SOURCE_LOG_POS=158, SOURCE_CONNECT_RETRY=30;`

在从数据库中开启主从同步

`START REPLICA;`

在从数据库中查看主从同步状态

`SHOW REPLICA STATUS \G;`

