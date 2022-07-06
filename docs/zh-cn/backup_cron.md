目录
=============

# 简介
目前无论 S3 还是 NFS 备份, 均支持定时备份, 使用 crontab 表达式来指定备份的时间策略. 直接在 cluster 的 yaml 中 spec 下设置 
字段 backupSchedule 即可. 例如:

```yaml
... 
spec:
  replicas: 3
  mysqlVersion: "5.7"
  backupSchedule: "0 0 0 * * *"  # daily
  ...
```
# 定时备份配置方式

## CRON 表达式格式

cron 表达式格式为: `[second] [minute] [hour] [day of month] [month] [day of week]` 由6个空格分隔的字段组成的时间结合


字段名   | 必须? | 允许值  | 允许的特殊符号
----------   | ---------- | --------------  | --------------------------
Seconds      | Yes        | 0-59            | * / , -
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?

> Note: Month 和 Day-of-week 字段值大小写不敏感. "SUN", "Sun", 和 "sun" 同等接受.

### 特殊字母
星号 ( * )

星号指示 cron 表达式将匹配所有值的字段; 例如, 使用 5th 字段 (月) 中的星号将指示每月. 

左斜线 ( / )

左斜线用来指示范围的增量. 例如, 在第一个字段 (分钟) 中使用 3-59/15 中第一部分(分钟)将指示小时中第 3 分钟开始, 此后表示每 15 分钟. "*\/..." 等同于 "first-last/...", 即增量超过最大可能范围的字段. "N/..." 等同于 "N-MAX/...", 即从 N 开始使用增量, 直到结束特定范围. 它不会超过这个范围

逗号 ( , )

逗号用来隔离列表中的项目. 例如, 使用 "MON,WED,FRI" 在第 5 个字段 (星期) 中将指示周一, 周三和周五.

连字号 ( - )

连字号用来指定范围. 例如, 9-17 将指示 9-17 从9am 到5pm 中的每一个小时.

问号 ( ? )

问号可以用来代替星号, 以便留空 day-of-month 或 day-of-week.

预定义预约 ( @ )
你可以用如下的一个预定义的预约来代替 cron 表达式.

Entry                  | Description                                | Equivalent To
-----                  | -----------                                | -------------
@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 0 1 1 *
@monthly               | Run once a month, midnight, first of month | 0 0 0 1 * *
@weekly                | Run once a week, midnight between Sat/Sun  | 0 0 0 * * 0
@daily (or @midnight)  | Run once a day, midnight                   | 0 0 0 * * *
@hourly                | Run once an hour, beginning of hour        | 0 0 * * * *
