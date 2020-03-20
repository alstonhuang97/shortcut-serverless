package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// URL : 網址相關操作
type URL struct{}

// DynamoInfo : Dynamo 儲存資訊
type DynamoInfo struct {
	ShortCode string `json:"ShortCode"` // 短網址代號
	OriginURL string `json:"OriginURL"` // 原網址
	Password  string `json:"Password"`  // 票卡密碼
	EndTime   string `json:"EndTime"`   // 使用期限 TTL
}

// query : 向 DynamoDB 查詢
func query(tableName, expressionAttribute, keyConditionExpression string) (*dynamodb.QueryOutput, error) {
	session := session.New()
	awsConfig := &aws.Config{
		Region: aws.String("ap-northeast-1"),
	}
	d := dynamodb.New(session, awsConfig)
	input := &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":shortCode": {
				S: aws.String(expressionAttribute),
			},
		},
		KeyConditionExpression: aws.String(keyConditionExpression),
		// Limit:                  aws.Int64(1),
	}
	result, err := d.Query(input)
	if err != nil {
		log.Printf("Query Error: %s", err.Error())
		return nil, err
	}
	if *result.Count == int64(0) {
		return nil, errors.New("Not exist")
	}
	return result, nil
}

// CheckShortCodeToDynamoDB : 從DynamoDB確認短網址是否存在, 並且回傳原始網址
func (r URL) CheckShortCodeToDynamoDB(shortCode string) (string, bool) {
	envStr := os.Getenv("GO_ENV")
	dynamoTable := "shortcut_url"
	if envStr == "production" {
		dynamoTable = "shortcut_url_production"
	}
	log.Printf("[CheckShortCodeToDynamoDB] Short code: %s", shortCode)

	urlInfo := DynamoInfo{}
	result, err := query(dynamoTable, shortCode, "ShortCode = :shortCode")
	if err != nil { // 短網址沒有重複
		log.Printf("[CheckShortCode] Short URL not found: %s", err.Error())
		return "", true
	}

	// err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &urlInfo)
	dynamodbattribute.UnmarshalMap(result.Items[0], &urlInfo)
	if err != nil {
		log.Printf("[CheckShortCode] UnmarshalMap error: %s", err.Error())
	}

	log.Printf("[CheckShortCode] Get origin url: %s", urlInfo.OriginURL)
	return urlInfo.OriginURL, false
}

// HandleRequest : 短網址還原 (lambda)
func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	input := request.PathParameters["shortURL"]
	url := URL{}
	originURL, _ := url.CheckShortCodeToDynamoDB(input)
	log.Println(originURL)
	if originURL == "" {
		return events.APIGatewayProxyResponse{Body: "Cannot found short url", StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{Headers: map[string]string{"location": originURL}, Body: originURL, StatusCode: 301}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
