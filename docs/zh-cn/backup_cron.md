目录
=============

# 简介
目前，无论 S3 还是 NFS 备份，均支持定时备份，并支持使用 crontab 表达式来指定备份的时间策略。您只需直接在集群的 YAML 文件的 spec 下设置 backupSchedule 字段。例如：

```yaml
... 
spec:
  replicas: 3
  mysqlVersion: "5.7"
  backupSchedule: "0 0 0 * * *"  # daily
  ...
```
# 定时备份配置方式

## cron 表达式格式

cron 表达式的格式为: `[second] [minute] [hour] [day of month] [month] [day of week]`，即由6个使用空格分隔的字段组成的时间组合。


字段名   | 必配 | 允许值  | 允许的特殊符号
----------   | ---------- | --------------  | --------------------------
Seconds      | Yes        | 0-59            | * / , -
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?

> 注意：`Month` 和 `Day of week` 字段值大小写不敏感，即 `SUN`, `Sun`, 和 `sun` 均接受。

### 特殊字符
星号（*）

星号可用在所有字段中，表示对应时间域的每一个时刻。例如，第 5 个字段（月）值为星号，表示每个月。

反斜线（/）

表示范围增量。例如，第 2 个字段（分钟）中的 3-59/15 表示从该小时的第 3 分钟开始，此后以 15 分钟为时间间隔执行备份。`*/y` 等同于 `min-max/y`。`n/y` 等同于 `n-max/y`，即从 n 开始使用增量, 直到特定范围结束。

逗号（,）

逗号用来隔离列表中的项目。例如，在第 5 个字段 (星期) 中使用 `MON,WED,FRI` 将表示周一、周三和周五。

连字号（-）

连字号用来指定范围。例如，在第 3 个字段 (小时) 中使用 `9-17` 表示从 9 点到 17 点间的每一个小时。

问号（?）

不指定值，仅日期和星期域支持该字符。当日期或星期域其中之一被指定了值以后，为了避免冲突，需要将另一个域的值设为`?`。

预定义时间表（@）
你可以用如下的预定义时间来代替 cron 表达式。

Entry                  | Description                                | Equivalent To
-----                  | -----------                                | -------------
@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 0 1 1 *
@monthly               | Run once a month, midnight, first of month | 0 0 0 1 * *
@weekly                | Run once a week, midnight between Sat/Sun  | 0 0 0 * * 0
@daily (or @midnight)  | Run once a day, midnight                   | 0 0 0 * * *
@hourly                | Run once an hour, beginning of hour        | 0 0 * * * *
