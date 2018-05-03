package dynamodb

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"

	"github.com/nerdify/tuc"
)

var cardsTable = "tuc_cards"

// CardService represents an dynamodb implementation of tuc.CardService.
type CardService struct{}

var _ tuc.CardService = &CardService{}

// List all Cards.
func (s *CardService) List(userID string) ([]tuc.Card, error) {
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
			":id": {
				S: &userID,
			},
		},
		KeyConditionExpression: aws.String("u_id = :id"),
		TableName:              &cardsTable,
	}

	req := svc.QueryRequest(input)
	res, err := req.Send()

	if err != nil {
		return nil, errors.Wrap(err, "getting items")
	}

	cards := []tuc.Card{}

	if err := dynamodbattribute.UnmarshalListOfMaps(res.Items, &cards); err != nil {
		return nil, errors.Wrap(err, "unmarshaling items")
	}

	return cards, nil
}

// Get individual card.
func (s *CardService) Get(userID, cardID string) (*tuc.Card, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]dynamodb.AttributeValue{
			"id": {
				S: &cardID,
			},
			"u_id": {
				S: &userID,
			},
		},
		TableName: &cardsTable,
	}

	req := svc.GetItemRequest(input)
	res, err := req.Send()

	if err != nil {
		return nil, errors.Wrap(err, "getting item")
	}

	if len(res.Item) == 0 {
		return nil, nil
	}

	var c tuc.Card

	if err := dynamodbattribute.UnmarshalMap(res.Item, &c); err != nil {
		return nil, errors.Wrap(err, "unmarshaling item")
	}

	return &c, nil
}

// Create a new card.
func (s *CardService) Create(card *tuc.Card) error {
	item, _ := dynamodbattribute.MarshalMap(card)
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: &cardsTable,
	}

	req := svc.PutItemRequest(input)
	_, err := req.Send()

	return err
}

// Update a card.
func (s *CardService) Update(userID, cardID string, balance float64) (*tuc.Card, error) {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
			":b": {
				N: aws.String(strconv.FormatFloat(balance, 'f', -1, 64)),
			},
		},
		Key: map[string]dynamodb.AttributeValue{
			"id": {
				S: &cardID,
			},
			"u_id": {
				S: &userID,
			},
		},
		ReturnValues:     dynamodb.ReturnValueAllNew,
		TableName:        &cardsTable,
		UpdateExpression: aws.String("SET balance = :b"),
	}

	req := svc.UpdateItemRequest(input)
	res, err := req.Send()

	if err != nil {
		return nil, errors.Wrap(err, "updating item")
	}

	var c tuc.Card

	if err := dynamodbattribute.UnmarshalMap(res.Attributes, &c); err != nil {
		return nil, errors.Wrap(err, "unmarshaling item")
	}

	return &c, nil
}

// Delete card.
func (s *CardService) Delete(userID, cardID string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]dynamodb.AttributeValue{
			"u_id": {
				S: &userID,
			},
			"id": {
				S: &cardID,
			},
		},
		TableName: &cardsTable,
	}

	req := svc.DeleteItemRequest(input)
	_, err := req.Send()

	return err
}
