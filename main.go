package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ses"
	"log"
	"strings"
	"time"
)

//Refer to https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/ses-example-send-email.html
const (
	// Replace sender@example.com with your "From" address.
	// This address must be verified with Amazon SES.
	Sender = "update@prod.6225csyekeyuliu.me" //"liu.keyu@gmail.com"

	// The character encoding for the email.
	CharSet = "UTF-8"
)

var sess *session.Session
var svc_ses *ses.SES
var svc_db *dynamodb.DynamoDB

func handleRequest(ctx context.Context, event events.SNSEvent) error {
	log.Print("Event started: ")

	// event
	records, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return err
	}
	log.Printf("EVENT: %s", records)

	//log timestamp
	currentTime := time.Now()
	log.Printf("Invocation started: %v", currentTime.Format("2020-11-26 15:04:05"))

	//log sns event is null
	log.Printf("Event is NULL: %v", event.Records == nil)

	//log number of records
	log.Printf("Number of Records: %v", len(event.Records))

	//log record message
	for _, m := range event.Records {
		log.Printf("Record Message: %v", m.SNS.Message)

		//send email to the user with sns notification message
		SendSESEmail(m.SNS.Message, m.SNS.UnsubscribeURL)
	}

	//log timestamp
	currentTime = time.Now()
	log.Printf("Invocation completed: %v", currentTime.Format("2020-11-26 15:04:05"))

	return nil
}

func initSession() *session.Session {
	log.Println("initialize aws session")
	if sess == nil {
		newSess, err := session.NewSession(&aws.Config{
			Region:aws.String("us-east-1")},
		)
		/*test locally
		newSess, err := session.NewSessionWithOptions(session.Options{
			// Specify profile to load for the session's config
			Profile: "prod",

			// Provide SDK Config options, such as Region.
			Config: aws.Config{
				Region: aws.String("us-east-1"),
			},

			// Force enable Shared Config support
			SharedConfigState: session.SharedConfigEnable,
		})*/

		if err != nil {
			log.Println("can't load the aws session")
			return nil
		} else {
			log.Println("loaded aws session")
			sess = newSess
		}
	}
	return sess
}

func initSESClient() *ses.SES {
	if svc_ses == nil {
		sess = initSession()
		// Create S3 service client
		svc_ses = ses.New(sess)
	}

	return svc_ses
}

func initDBClient() *dynamodb.DynamoDB {
	if svc_db == nil {
		sess = initSession()
		// Create S3 service client
		svc_db = dynamodb.New(sess)
	}

	return svc_db
}

//send email to the user with sns notification message
func SendSESEmail(message string, unsubscribe_url string) {
	// Create an SES session.
	svc_ses := initSESClient()

	//message content:
	//0 Create
	//1 bookId
	//2 bookTitle
	//3 Lastname+firstName
	//4 EmailAddress
	email_context := strings.Split(message, ",")
	//"Create Book"+","+b.getId()+","+b.getTitle()+","+user.getLast_name()+" "+user.getFirst_name()+","+user.getEmailAddress()
	if len(email_context) != 5 {
		log.Println("Message length is not as expected")
		return
	}

	//prepare information before assemble the email
	Recipient := "liu.keyuneu@gmail.com"

	Subject := "Notification from 6225csyekeyuliu"
	HtmlBody := "<h1>Notification from 6225csyekeyuliu</h1><p>This email was sent with <a href='https://aws.amazon.com/ses/'>Amazon SES</a> using the <a href='https://aws.amazon.com/sdk-for-go/'>AWS SDK for Go</a>.</p>"
	TextBody := "This email was sent from prod.6225csyekeyuliu.me with Amazon SES."

	if email_context[0] == "Create Book" {
		Subject = fmt.Sprintf("Your book '%v' on 6225csyekeyuliu.me has been created", email_context[2])
		HtmlBody = fmt.Sprintf("<h1>Notification from 6225csyekeyuliu.me</h1>"+
			"<p>Hi %v,</p>"+
			"<p>The Book %v, %v owned by %v on 6225csyekeyuliuliukeyu.me has been created.</p>"+
			"<p>See more details: </p>"+
			"<p><a href='http://prod.6225csyekeyuliu.me/v1/books/%v'>This Book</a></p>"+
			"<p><a href='http://prod.6225csyekeyuliu.me/v1/mybooks'>Books under your name</a></p>"+
			"<p>This email was sent from prod.6225csyekeyuliuliukeyu.me</a> with <a href='https://aws.amazon.com/ses/'>Amazon SES</a>.</p>"+
			"<p><a href='%v'>Unsubscribe</a></p>",
			email_context[3], email_context[1], email_context[2], email_context[4], email_context[1], unsubscribe_url)
		TextBody = fmt.Sprintf("Hi %v,\n"+
		"The Book %v , %v owned by %v on 6225csyekeyuliuliukeyu.me has been created.\n"+
		"See more details: \n"+
		"http://prod.6225csyekeyuliu.me/v1/books/%v This Book \n"+
		"http://prod.6225csyekeyuliu.me/v1/mybooks Books under your name \n"+
		"This email was sent from prod.6225csyekeyuliuliukeyu.me with Amazon SES. \n"+
		"Unsubscribe: %v.",
		email_context[3], email_context[1], email_context[2], email_context[4], email_context[1], unsubscribe_url)
	} else if email_context[0] == "Delete Book" {
		Subject = fmt.Sprintf("The Book '%v' of '%v' on 6225csyekeyuliu.me has been deleted", email_context[1], email_context[2])
		HtmlBody = fmt.Sprintf("<h1>Notification from 6225csyekeyuliu.me</h1>"+
			"<p>Hi %v,</p>"+
			"<p>The Book %v, %v owned by %v on 6225csyekeyuliu.me has been deleted.</p>"+
			"<p>This email was sent from 6225csyekeyuliu.me with <a href='https://aws.amazon.com/ses/'>Amazon SES</a>.</p>"+
			"<p><a href='%v'>Unsubscribe</a></p>",
			email_context[3], email_context[1], email_context[2], email_context[4], unsubscribe_url)
		TextBody = fmt.Sprintf("Hi %v,\n"+
			"The Book %v, %v owned by %v on 6225csyekeyuliu.me has been deleted.\n"+
			"This email was sent from 6225csyekeyuliu.me with Amazon SES.\n"+
			"Unsubscribe: %v.",
			email_context[3], email_context[1], email_context[2], email_context[4], unsubscribe_url)
	} else {
		log.Println("The message is not started as expected")
	}

	//search for email, if already sent, return, otherwise, put in DynamoDB table, and send email
	isExist := searchItemInDynamoDB(TextBody)
	if isExist {
		log.Println("The email has already been sent")
		return
	}

	if err := addItemToDynamoDB(TextBody); err != nil {
		log.Printf("Failed to put email item into DynamoDB table: %v", err)
		return
	}

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(HtmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(Subject),
			},
		},
		Source: aws.String(Sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := svc_ses.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				log.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				log.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				log.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}

		return
	}

	log.Println("Email Sent to address: " + Recipient)
	log.Println(result)
}

func searchItemInDynamoDB(TextBody string) bool {
	//initialize dynamodb client

	svc_db := initDBClient()

	tableName := "csye6225"

	result, err := svc_db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(TextBody),
			},
		},
	})
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if result.Item == nil {
		log.Println("Search email in dynamodb: false")
		return false
	}

	log.Printf("Got item output: %v", result)
	return true
}

//add the email to DynomoDB to avoid sending duplicate emails to users
//refer to https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/dynamo-example-create-table-item.html
func addItemToDynamoDB(TextBody string) error {
	//initialize dynamodb client
	svc_db := initDBClient()

	tableName := "csye6225"

	item := map[string]*dynamodb.AttributeValue{
		"id": {
			S: aws.String(TextBody),
		},
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	}

	_, err := svc_db.PutItem(input)
	if err != nil {
		log.Printf("Got error calling PutItem: %v\n", err)
		return err
	}

	log.Println("Successfully added email: '" + TextBody + "'")

	return nil
}

func main() {
	lambda.Start(handleRequest)
	//test
	//SendSESEmail("update answer, QuestionID: 1, QuestionText: meaning of cat, UserName: Jane Jenny, UserEmail: jingzhangng20@gmail.com, AnswerID: 1, AnswerText: lovely, Link: http://prod.6225csyekeyuliu.me:80/v1/question/b1db1852-5c5f-457c-b94d-56b917064eee/answer/931bb982-3573-4187-a8e6-d0870901b880","https://sns.us-east-1.amazonaws.com/?Action=Unsubscribe\u0026SubscriptionArn=arn:aws:sns:us-east-1:907204364947:topic:93f10269-ced3-41e2-bf2c-484d0edbf8d1")
	//SendSESEmail("delete answer, QuestionID: 1, QuestionText: meaning of cat, UserName: Jane Jenny, UserEmail: jingzhangng20@gmail.com, AnswerID: 1, AnswerText: lovely")
}
