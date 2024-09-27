---
layout: default
title: Scheduler
parent: Usage
nav_order: 4
---
# Scheduler

Each action can be automatically called in a cron-tab like style.

Accuracy is +/- 30 seconds.

Each schedule task is invoked sequentially (to reduce resource usage), so
ensure that you set the maximum execution time properly.

If any error occurred during execution - it will be printed in a log. 

UI:
 
1. click to any created application
2. click to scheduler tab


Format:

`[second] [minute] [hour] [day] [month] [week]`

You can use [https://crontab.guru/](https://crontab.guru/) to check, however, add seconds after test
