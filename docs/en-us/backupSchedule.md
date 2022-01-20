backupSchedule: "0 0 0 * * *"  # daily

Crontab takes 6 arguments from the traditional 5. The additional argument is a seconds field. Some crontab examples and their predefined schedulers:

| Entry         | Equivalent To          | Description                                |
| ------------- | -----                  | -----------                                |
| 15 0 0 1 1 *  | @yearly (or @annually) | Run once a year, midnight, Jan. 1st, 15th second        |
| 0 0 0 1 * *   | @monthly               | Run once a month, midnight, first of month, 0 second |
| 0 0 0 * * 0   | @weekly                | Run once a week, midnight between Sat/Sun, 0 second  |
| 0 0 0 * * *   | @daily (or @midnight)  | Run once a day, midnight, 0 second, 0 second                   |
| 0 0 * * * *   | @hourly                | Run once an hour, beginning of hour, 0 second        |