# execution-system

## API
### POST /api/task/:task_id/execute
Asynchronous call to execute a task (maybe a learning or applying type).
To track the health of running task - use GET /api/task/:task_id/state.

On success, schedules a task to run in kubernetes cluster.

#### Request
```json
{
  "model_id": string,
  "user_input_id": string,
}
```

#### Responses
```json
200
{
  "result": "SUCCESS" | "ALREADY_RUNNING" | "FAILED"
}
```
```json
400
{
  "error": string
}
```

### GET /api/task/:task_id/state
Check current state of task.

In future, we can get rid of `STILL_RUNNING` in favor of more concrete stage.

#### Responses
```json
{
  "state": "FINISHED" | "STILL_RUNNING" | "NOT_SCHEDULED"
}
```

## Configuration
TODO
