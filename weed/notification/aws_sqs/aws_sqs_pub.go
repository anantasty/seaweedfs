package aws_sqs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/joeslay/seaweedfs/weed/glog"
	"github.com/joeslay/seaweedfs/weed/notification"
	"github.com/joeslay/seaweedfs/weed/util"
	"github.com/golang/protobuf/proto"
)

func init() {
	notification.MessageQueues = append(notification.MessageQueues, &AwsSqsPub{})
}

type AwsSqsPub struct {
	svc      *sqs.SQS
	queueUrl string
}

func (k *AwsSqsPub) GetName() string {
	return "aws_sqs"
}

func (k *AwsSqsPub) Initialize(configuration util.Configuration) (err error) {
	glog.V(0).Infof("filer.notification.aws_sqs.region: %v", configuration.GetString("region"))
	glog.V(0).Infof("filer.notification.aws_sqs.sqs_queue_name: %v", configuration.GetString("sqs_queue_name"))
	return k.initialize(
		configuration.GetString("aws_access_key_id"),
		configuration.GetString("aws_secret_access_key"),
		configuration.GetString("region"),
		configuration.GetString("sqs_queue_name"),
	)
}

func (k *AwsSqsPub) initialize(awsAccessKeyId, aswSecretAccessKey, region, queueName string) (err error) {

	config := &aws.Config{
		Region: aws.String(region),
	}
	if awsAccessKeyId != "" && aswSecretAccessKey != "" {
		config.Credentials = credentials.NewStaticCredentials(awsAccessKeyId, aswSecretAccessKey, "")
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return fmt.Errorf("create aws session: %v", err)
	}
	k.svc = sqs.New(sess)

	result, err := k.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
			return fmt.Errorf("unable to find queue %s", queueName)
		}
		return fmt.Errorf("get queue %s url: %v", queueName, err)
	}

	k.queueUrl = *result.QueueUrl

	return nil
}

func (k *AwsSqsPub) SendMessage(key string, message proto.Message) (err error) {

	text := proto.MarshalTextString(message)

	_, err = k.svc.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"key": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(key),
			},
		},
		MessageBody: aws.String(text),
		QueueUrl:    &k.queueUrl,
	})

	if err != nil {
		return fmt.Errorf("send message to sqs %s: %v", k.queueUrl, err)
	}

	return nil
}
