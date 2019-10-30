**Get JSON message with files from RABBITMQ queue and send it to defined bucket on AWS S3.**

Fail fast approach, in case of any error send NACK to RABBITMQ and exit. 

**Usage:**

1 _Set environment variables:_

      - AMQP_URL=amqp://guest:guest@rabbitmq:5672/
      - AMQP_QUEUE=aws.s3.uploader
      - AWS_REGION=us-east-1
      - AWS_ACCESS_KEY_ID=AKID
      - AWS_SECRET_ACCESS_KEY=SECRET
      - AWS_S3_BUCKET=bucket

2 _Publish RABBITMQ JSON message:_

```json
{
  "files": [
    {
      "key": "/2019/30/10/hi.txt",
      "file_content_base64": "aGVsbG8="
    }
  ]
}
```


3 _Run app_

```bash
go run .
```


