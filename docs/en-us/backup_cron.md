Contents
=============
- [Overview](#overview)
- [Configuration of scheduled backups](#configuration-of-scheduled-backups)
  - [Cron expression format](#cron-expression-format)
    - [Special characters](#special-characters)
    - [Predefined schedules](#predefined-schedules)

# Overview
The scheduled backup is currently supported for both S3 and NFS backups. You can use the cron expression to specify the backup schedule. Set the `backupSchedule` parameter under the `spec` field in the YAML file of the cluster, for example:

```yaml
... 
spec:
  replicas: 3
  mysqlVersion: "5.7"
  backupSchedule: "0 0 0 * * *"  # daily
  ...
```
# Configuration of scheduled backups

## Cron expression format

A cron expression represents a set of times, using 6 space-separated fields in the format of `[second] [minute] [hour] [day] [month] [day of week]`.

| Field name   | Mandatory | Allowed values  | Allowed special characters |
| ------------ | --------- | --------------- | -------------------------- |
| Seconds      | Yes       | 0-59            | * / , -                    |
| Minutes      | Yes       | 0-59            | * / , -                    |
| Hours        | Yes       | 0-23            | * / , -                    |
| Day of month | Yes       | 1-31            | * / , - ?                  |
| Month        | Yes       | 1-12 or JAN-DEC | * / , -                    |
| Day of week  | Yes       | 0-6 or SUN-SAT  | * / , - ?                  |

> Note: `Month` and `Day-of-week` field values are case-insensitive. `SUN`, `Sun`, and `sun` are equally accepted.

### Special characters
Asterisk ( * )

The asterisk indicates that the cron expression will match for all values of the field. For example, using an asterisk in the 5th field (month) would indicate every month.

Slash ( / )

Slashes are used to describe increments of ranges. For example `3-59/15` in the 1st field (minutes) would indicate the 3rd minute of the hour and every 15 minutes thereafter. The form `*\/...` is equivalent to the form `first-last/...`, that is, an increment over the largest possible range of the field. The form `N/...` is accepted as meaning `N-MAX/...`, that is, starting at `N`, use the increment until the end of that specific range. It does not wrap around.

Comma ( , )

Commas are used to separate items of a list. For example, using `MON,WED,FRI` in the 5th field (day of week) would mean Mondays, Wednesdays and Fridays.

Hyphen ( - )

Hyphens are used to define ranges. For example, `9-17` would indicate every hour between 9am and 5pm inclusive.

Question mark ( ? )

Question mark may be used instead of `*` for leaving either day-of-month or day-of-week blank.

### Predefined schedules

You may use one of several pre-defined schedules in place of a cron expression.

| Entry                  | Description                                | Equivalent To |
| ---------------------- | ------------------------------------------ | ------------- |
| @yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 0 1 1 *   |
| @monthly               | Run once a month, midnight, first of month | 0 0 0 1 * *   |
| @weekly                | Run once a week, midnight on Sunday        | 0 0 0 * * 0   |
| @daily (or @midnight)  | Run once a day, midnight                   | 0 0 0 * * *   |
| @hourly                | Run once an hour, beginning of hour        | 0 0 * * * *   |