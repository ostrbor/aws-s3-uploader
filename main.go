package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/streadway/amqp"
	"log"
	"os"
)

type queueMessage struct {
	Files []struct {
		Key               string `json:"key"`
		FileContentBase64 string `json:"file_content_base64"`
	} `json:"files"`
}

func main() {
	amqpUrl := getEnv("AMQP_URL")
	amqpQueueName := getEnv("AMQP_QUEUE")
	bucketName := getEnv("AWS_S3_BUCKET")

	sess := session.Must(session.NewSession())
	createBucketIfNotExists(sess, bucketName)
	uploader := s3manager.NewUploader(sess)

	for qmr := range queueMessagesRaw(amqpUrl, amqpQueueName) {
		m := &queueMessage{}
		err := json.Unmarshal(qmr.Body, m)
		if err != nil {
			qmr.Nack(false, true)
			log.Fatalf("unable to unmarshal %s", qmr.Body)
		}

		for _, f := range m.Files {
			fc, err := base64.StdEncoding.DecodeString(f.FileContentBase64)
			if err != nil {
				qmr.Nack(false, true)
				log.Fatalf("unable to decode file %s", f.Key)
			}
			_, err = uploader.Upload(&s3manager.UploadInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(f.Key),
				Body:   bytes.NewReader(fc),
			})
			if err != nil {
				qmr.Nack(false, true)
				log.Fatalf("unable to upload: %v", err)
			}
			qmr.Ack(false)
			fmt.Println("successfully uploaded ", f.Key)
		}
	}
}

func createBucketIfNotExists(sess *session.Session, bucketName string) {
	svc := s3.New(sess)
	output, err := svc.ListBuckets(nil)
	if err != nil {
		log.Fatalf("unable to list buckets: %v; sess: %v", err, sess)
	}
	var found bool
	for _, b := range output.Buckets {
		if bucketName == aws.StringValue(b.Name) {
			found = true
		}
	}
	if !found {
		_, err = svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			log.Fatalf("unable to create bucket %s", bucketName)
		}
		fmt.Println("successfully created bucket ", bucketName)
	}
}

func queueMessagesRaw(amqpUrl, queueName string) <-chan amqp.Delivery {
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
