package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"

	"github.com/nerdify/tuc"
)

var usersTable = "tuc_users"

// UserService represents an dynamodb implementation of tuc.UserService.
type UserService struct{}

var _ tuc.UserService = &UserService{}

// Find returns the User with the specified id.
func (s *UserService) Find(id string) (*tuc.User, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]dynamodb.AttributeValue{
			"id": {
				S: &id,
			},
		},
		TableName: &usersTable,
	}

	req := svc.GetItemRequest(input)
	res, err := req.Send()

	if err != nil {
		return nil, errors.Wrap(err, "getting item")
	}

	if len(res.Item) == 0 {
		return nil, nil
	}

	var u tuc.User

	if err := dynamodbattribute.UnmarshalMap(res.Item, &u); err != nil {
		return nil, errors.Wrap(err, "unmarshaling item")
	}

	return &u, nil
}

// Create creates a new user.
func (s *UserService) Create(user *tuc.User) error {
	item, _ := dynamodbattribute.MarshalMap(user)
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: &usersTable,
	}

	req := svc.PutItemRequest(input)
	_, err := req.Send()

	return err
}
