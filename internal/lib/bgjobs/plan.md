
## Background Jobs
- Using redis server as a message queue
- Maybe switch over to asynq once I figure out how the redis server works
- Figure out how to view all data in redis server and how to delete data
- Does restarting the redis server take away all data, or is it restart of machine
- Do people run redis server on its own machine?
- Run redis server locally, and create client (should be like asynq) that pushes and pulls tasks from queue
- From webhook, it should add task to queue
- Create a workers process that constantly runs and checks for tasks in queue
