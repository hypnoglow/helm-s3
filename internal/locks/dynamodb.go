package locks

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBLock struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBLock(tableName string, cfg aws.Config, optFns ...func(*dynamodb.Options)) *DynamoDBLock {
	return &DynamoDBLock{
		client: dynamodb.NewFromConfig(
			cfg,
			optFns...,
		),
		tableName: tableName,
	}
}

func NewDynamoDBLockWithDefaultConfig(tableName string, optFns ...func(*dynamodb.Options)) (*DynamoDBLock, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	return NewDynamoDBLock(tableName, cfg, optFns...), nil
}

func (d *DynamoDBLock) Lock(ctx context.Context, lockID string) error {
	params := &dynamodb.PutItemInput{
		Item: map[string]dynamodbtypes.AttributeValue{
			"LockID": &dynamodbtypes.AttributeValueMemberS{
				Value: lockID,
			},
		},
		TableName:           aws.String(d.tableName),
		ConditionExpression: aws.String("attribute_not_exists(LockID)"),
	}

	_, err := d.client.PutItem(ctx, params)
	if err != nil && strings.Contains(err.Error(), "ConditionalCheckFailedException") {
		return ErrAlreadyLocked
	}

	return err
}

func (d *DynamoDBLock) Unlock(ctx context.Context, lockID string) error {
	params := &dynamodb.DeleteItemInput{
		Key: map[string]dynamodbtypes.AttributeValue{
			"LockID": &dynamodbtypes.AttributeValueMemberS{
				Value: lockID,
			},
		},
		TableName: aws.String(d.tableName),
	}
	_, err := d.client.DeleteItem(ctx, params)
	return err
}
