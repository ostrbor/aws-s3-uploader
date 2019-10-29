package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/streadway/amqp"
	"log"
	"os"
)

type queueMessage struct {
	FileName                 string `json:"file_name"`
	Base64EncodedFileContent string `json:"base64_encoded_file_content"`
}

func main() {
	amqpUrl := getEnv("AMQP_URL")
	amqpQueueName := getEnv("AMQP_QUEUE_NAME")

	var bucket, key string
	// TODO get bucket, key and file from rabbitmq

	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	result, err := svc.ListBuckets(nil)
	if err != nil {
		log.Fatal("unable to list buckets")
	}
	var found bool
	for _, b := range result.Buckets {
		bucketName := aws.StringValue(b.Name)
		if bucketName == bucket {
			found = true
		}
	}
	if !found {
		_, err = svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			log.Fatal("unable to create bucket")
		}
	}

	uploader := s3manager.NewUploader(sess)

	for message := range setUpRabbitMQ(amqpUrl, amqpQueueName) {
		m := &queueMessage{}
		err := json.Unmarshal(message.Body, m)
		if err != nil {
			message.Nack(false, true)
			log.Fatal("unable to unmarshal", message.Body)
		}

		fc, err := base64.StdEncoding.DecodeString(m.Base64EncodedFileContent)
		if err != nil {
			message.Nack(false, true)
			log.Fatal("unable to decode")
		}

		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   bytes.NewReader(fc),
		})
		if err != nil {
			message.Nack(false, true)
			log.Fatal("unable to upload")
		}
	}

}

func setUpRabbitMQ(amqpUrl, queueName string) <-chan amqp.Delivery {
	amqpConn, err := amqp.Dial(amqpUrl)
	if err != nil {
		log.Fatal(err)
	}
	amqpChannel, err := amqpConn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	amqpChannel.Qos(1, 0, false)
	amqpQueue, err := amqpChannel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	messageChannel, err := amqpChannel.Consume(amqpQueue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	return messageChannel
}

func getEnv(k string) (v string) {
	v = os.Getenv(k)
	if v == "" {
		log.Fatalf("%v must be set\n", k)
	}
	return v
}
