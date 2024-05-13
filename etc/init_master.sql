CREATE USER slave IDENTIFIED BY 'slave';
GRANT SELECT, REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'slave'@'%';
ALTER USER 'slave'@'%' IDENTIFIED WITH mysql_native_password BY 'slave';
CREATE USER canal IDENTIFIED BY 'canal';  
GRANT SELECT, REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'canal'@'%';
FLUSH PRIVILEGES;