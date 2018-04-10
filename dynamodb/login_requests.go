package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"

	"github.com/nerdify/tuc"
)

var (
	loginRequestsTable = "tuc_login_requests"
)

// LoginRequestService represents an dynamodb implementation of tuc.LoginRequestService.
type LoginRequestService struct{}

var _ tuc.LoginRequestService = &LoginRequestService{}

// Create a new login request.
func (s *LoginRequestService) Create(request *tuc.LoginRequest) error {
	item, _ := dynamodbattribute.MarshalMap(request)
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: &loginRequestsTable,
	}

	req := svc.PutItemRequest(input)
	_, err := req.Send()

	return err
}

// Delete a login request.
func (s *LoginRequestService) Delete(email, code string) error {
	input := &dynamodb.DeleteItemInput{
		ConditionExpression: aws.String("#t = :t and #v = :v"),
		ExpressionAttributeNames: map[string]string{
			"#t": "request_token",
			"#v": "verified",
		},
		ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
			":t": {
				S: &code,
			},
			":v": {
				BOOL: aws.Bool(true),
			},
		},
		Key: map[string]dynamodb.AttributeValue{
			"u_id": {
				S: &email,
			},
		},
		TableName: &loginRequestsTable,
	}

	req := svc.DeleteItemRequest(input)
	_, err := req.Send()

	return err
}

// Verify a login request.
func (s *LoginRequestService) Verify(email, token string) error {
	input := &dynamodb.UpdateItemInput{
		ConditionExpression: aws.String("#t = :t and #v = :vf"),
		ExpressionAttributeNames: map[string]string{
			"#t": "verification_token",
			"#v": "verified",
		},
		ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
			":t": {
				S: &token,
			},
			":vf": {
				BOOL: aws.Bool(false),
			},
			":vt": {
				BOOL: aws.Bool(true),
			},
		},
		Key: map[string]dynamodb.AttributeValue{
			"u_id": {
				S: &email,
			},
		},
		TableName:        &loginRequestsTable,
		UpdateExpression: aws.String("SET #v = :vt"),
	}

	req := svc.UpdateItemRequest(input)
	_, err := req.Send()

	return err
}
