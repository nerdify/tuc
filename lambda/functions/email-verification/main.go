package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

var (
	cfg, _ = external.LoadDefaultAWSConfig()
	svc    = ses.New(cfg)
	tmp    = template.Must(template.ParseFiles("template.html"))
)

func sendEmail(email, token string) {
	url := fmt.Sprintf("https://saldotuc.com/api/authenticate?email=%s&token=%s", email, token)

	var buf bytes.Buffer

	if err := tmp.Execute(&buf, url); err != nil {
		fmt.Println(err.Error())
		return
	}

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []string{email},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(buf.String()),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String("Verificación de inicio de sesión - Saldo TUC"),
			},
		},
		Source: aws.String("signin@saldotuc.com"),
	}

	req := svc.SendEmailRequest(input)
	_, err := req.Send()

	if err != nil {
		fmt.Println(err.Error())
	}
}

func handler(ctx context.Context, e events.DynamoDBEvent) {
	for _, record := range e.Records {
		if record.EventName == "REMOVE" {
			continue
		}

		item := record.Change.NewImage

		if item["verified"].Boolean() {
			continue
		}

		email := item["u_id"].String()
		token := item["verification_token"].String()

		sendEmail(email, token)
	}
}

func main() {
	lambda.Start(handler)
}
