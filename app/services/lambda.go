package services

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

func InvokeLambda(ctx context.Context, lambdaClient *lambda.Client, functioName string) error {
	payload, _ := json.Marshal(map[string]string{"first_name": "Sumedh", "last_name": "Shakya"})
	op, err := lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(functioName),
		Payload:      payload,
		LogType:      types.LogTypeTail,
	})
	if err != nil {
		log.Println("Error Invoking Lambda --> ", err)
		return err
	}
	log.Println("Lambda Invoke O/P -> ", string(op.Payload))
	return nil
}
