# Go task scheduler

A distributed task scheduler based off of the architecture used at slack

## How to run

A docker-compose file has been provided with all the different components. Make sure you have docker installed and run the following command

```bash
docker-compose up
```

Some services will continously fail and restart until kafka and consul are running so you have to wait.

## Testing

The `kafkagate` service is listening for requests on port 8000. Send a `POST` request to `http://localhost:8000/job` with the following payload.
```json
{
    "jobType": "test",
    "payload": "hello world"
}

```
Check the logs from docker compose and the job should be added to kafka, then sent to redis, then processed by the worker as shown below.

![image](./images/image.png
)
