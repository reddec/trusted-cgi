# Scheduler

Each action could be automatically called in cron-tab like style.

Accuracy is +/- 30 seconds.

Each schedule task invoke sequentially (to reduce resource usage), so
ensure that you set maximum execution time properly.

If any error occurred during execution - error will be printed in a log. 

UI:
 
1. click to any created application
2. click to scheduler tab