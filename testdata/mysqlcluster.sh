#!/usr/bin/env bash

godoNum=${GODO_NUM}
gossh=~/go/bin/gossh
pump=~/go/bin/pump

# 注意主机顺序，主2，从1，主1
export GOSSH_HOSTS="192.168.136.(23 30):8022 root/112113 id=(2 1)"
export PUMP_DS="root:8BE4-62A56623CA8c@tcp(127.0.0.1:9633)/?charset=utf8mb4&parseTime=true&loc=Local"
export MYSQL_EXEC="%host MYSQL_PWD='8BE4-62A56623CA8c' mysql -uroot -h127.0.0.1 -P9633 -e "

if ((godoNum == 100)); then
    echo 启动MySQL
    ${gossh} --cmds="%host systemctl start mysqld"
    echo 构建集群
    ${gossh} --cmds="%host /opt/friday/shells/mci.sh \
--Master1Addr=192.168.136.30 --Master2Addr=192.168.136.251 \
--Password=8BE4-62A56623CA8c --Port=9633 --pbepwd=cba321 --ShellTimeout=10m"
    echo 建库建表
    ${pump} --sqls " \
create database a; create table a.ta(name varchar(10),age int); \
create database b; create table b.tb(name varchar(10),age int); \
create database c; create table c.tc(name varchar(10),age int); \
create database d; create table d.td(name varchar(10),age int); \
create database e; create table e.te(name varchar(10),age int); \
"
elif ((godoNum == 1)); then
    echo 重启MySQL服务
    ${gossh} --cmds="%host-1 systemctl restart mysqld"
elif ((godoNum == 2)); then
    echo 重启MySQL服务
    ${gossh} --cmds="%host-2 systemctl restart mysqld"
elif ((godoNum == 3)); then
    echo 查看从状态
    ${gossh} --quoteReplace=%q --cmds="${MYSQL_EXEC} %qshow slave status%q"
elif ((godoNum == 4)); then
    echo 查看各个库中表的数据行数
    ${gossh} --separator=%s --quoteReplace=%q --cmds="${MYSQL_EXEC} %q \
select (select count(*) from a.ta) as ca,  \
       (select count(*) from b.tb) as cb,  \
       (select count(*) from c.tc) as cc,  \
       (select count(*) from d.td) as cd,  \
        (select count(*) from e.te) as ce; \
%q"
elif ((godoNum == 5)); then
    echo 插入1000条数据
    ${pump}  -b 100 --rows 1000 -t a.ta,b.tb,c.tc,d.td,e.te
fi
