package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SakyaSumedh/irsa/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const (
	serverPort = ":8080"
)

var (
	sqsClient       *sqs.Client
	snsClient       *sns.Client
	s3Client        *s3.Client
	presignS3Client *s3.PresignClient
	redisClient     *redis.Client
	dynamodbClient  *dynamodb.Client
	lambdaClient    *lambda.Client
	ctx             = context.Background()
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		os.Exit(1)
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}

func UploadToS3() (map[string]string, error) {
	log.Println("Uploading file to S3 bucket")
	reader := strings.NewReader("IRSA test file")

	key := strconv.FormatInt(time.Now().Unix(), 10) + ".txt"

	presignUrl, err := presignS3Client.PresignPutObject(
		ctx, &s3.PutObjectInput{
			Bucket: aws.String(GetEnv("AWS_S3_BUCKET_NAME")),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(time.Minute*15),
	)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, presignUrl.URL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "multipart/form-data")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return map[string]string{"key": key}, err
}

func ListS3Objects() (*s3.ListObjectsV2Output, error) {
	log.Println("Listing all objects in s3 bucket")
	data, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(GetEnv("AWS_S3_BUCKET_NAME")),
	})
	return data, err
}

func HandleS3Request(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		data, err := ListS3Objects()
		if err != nil {
			log.Println("Error fetching objects from S3 bucket:", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"msg": "Error listing objects."})
			return
		}
		json.NewEncoder(w).Encode(data.Contents)

	} else if r.Method == "POST" {
		data, err := UploadToS3()
		if err != nil {
			log.Println("Error uploading data to S3 bucket:", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"msg": "Error uploading data."})
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(data)

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"msg": "Method Not Allowed"})
		return
	}
}

func SendMessageToSQS() error {
	log.Println("Sending message to SQS queue")
	_, err := sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:               aws.String(GetEnv("AWS_SQS_FIFO_URL")),
		MessageGroupId:         aws.String(GetEnv("AWS_SQS_MESSAGE_GROUPID")),
		MessageDeduplicationId: aws.String(strconv.FormatInt(time.Now().Unix(), 10)),
		MessageBody:            aws.String("IRSA Test Message"),
	})
	return err
}

func HandleSQSRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" {
		err := SendMessageToSQS()
		if err != nil {
			log.Println("Error sending message to SQS:", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"msg": "Error sending message to SQS."})
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"msg": "Message queued..."})

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"msg": "Method Not Allowed."})
		return
	}
}

func PublishNotificationToSNS() error {
	log.Println("Publishing message to SNS Topic")
	_, err := snsClient.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(GetEnv("AWS_SNS_TOPIC_ARN")),
		Message:  aws.String("IRSA Test Message"),
	})
	return err
}

func HandleSNSRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" {
		err := PublishNotificationToSNS()
		if err != nil {
			log.Println("Error publishing message to SNS:", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"msg": "Error publishing message to SNS."})
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"msg": "Message published..."})

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"msg": "Method Not Allowed."})
		return
	}
}

func HandleRedisRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		log.Println("Fetching from redis...")
		user_data := redisClient.Get(ctx, "user")

		err := user_data.Err()
		if err != nil {
			log.Println("Error fetching from key 'user' -> ", err.Error())
			resp, _ := json.Marshal(map[string]string{"msg": "Empty data"})
			w.Write(resp)
			return
		}
		data, _ := user_data.Result()
		fmt.Println(data)
		arr_data := strings.Split(data, ", ")

		var resp_data []User
		for i := 1; i < len(arr_data); i++ {
			var user User
			json.Unmarshal([]byte(arr_data[i]), &user)
			resp_data = append(resp_data, user)
		}
		json.NewEncoder(w).Encode(resp_data)

	} else if r.Method == "POST" {
		log.Println("Appending to redis...")
		var user User

		reqBody := json.NewDecoder(r.Body)
		reqBody.DisallowUnknownFields()
		if err := reqBody.Decode(&user); err != nil {
			log.Println("Error Parsing request data -> ", err.Error())
			msg := strings.Split(err.Error(), ": ")
			msg = strings.Split(msg[1], "\"")
			log.Println(msg)

			resp, _ := json.Marshal(map[string]string{msg[1]: msg[0][:len(msg[0])-1]})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(resp)
			return
		}

		data, _ := json.Marshal(user)
		fmt.Println(data)
		if err := redisClient.Append(ctx, "user", ", "+string(data)).Err(); err != nil {
			http.Error(w, `{"msg": "Error appending data to redis"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"msg": "Added user data to redis"}`))

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"msg": "Method Not Allowed."})
		return
	}
}

func HandleDynamoDBRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	type User struct {
		Name  string `dynamodbav:"name"`
		Email string `dynamodbav:"email"`
	}
	var user User

	if r.Method == "GET" {
		data, err := dynamodbClient.Scan(ctx, &dynamodb.ScanInput{
			TableName: aws.String(GetEnv("AWS_DYNAMODB_TABLE_NAME")),
		})
		if err != nil {
			log.Println("Error fetching data:", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"msg": "Error fetching data."})
			return
		}

		var users []User
		err = attributevalue.UnmarshalListOfMaps(data.Items, &users)
		if err != nil {
			log.Println("Error parsing fetched data:", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"msg": "Error parsing fetched data."})
			return
		}
		json.NewEncoder(w).Encode(users)

	} else if r.Method == "POST" {
		reqBody := json.NewDecoder(r.Body)
		reqBody.DisallowUnknownFields()
		if err := reqBody.Decode(&user); err != nil {
			log.Println("Error Parsing request data -> ", err.Error())
			msg := strings.Split(err.Error(), ": ")
			msg = strings.Split(msg[1], "\"")
			log.Println(msg)

			resp, _ := json.Marshal(map[string]string{msg[1]: msg[0][:len(msg[0])-1]})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(resp)
			return
		}

		item, err := attributevalue.MarshalMap(user)
		if err != nil {
			log.Println("Error parsing data:", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"msg": "Error parsing data."})
			return
		}
		_, err = dynamodbClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(GetEnv("AWS_DYNAMODB_TABLE_NAME")),
			Item:      item,
		})
		if err != nil {
			log.Println("Error writing data to table:", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"msg": "Error writing data to dynamodb table."})
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"msg": "Data written to dynamodb table."})

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"msg": "Method Not Allowed."})
		return
	}
}

func HandleLambdaRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		functionName := GetEnv("AWS_FUNCTION_NAME")
		if err := services.InvokeLambda(ctx, lambdaClient, functionName); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"msg": "Error Invoking Lambda."})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"msg": "Successfully Invoked Lambda."})
		return
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"msg": "Method Not Allowed."})
		return
	}
}

func init() {
	LoadEnv()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(GetEnv("AWS_REGION")),
	)
	if err != nil {
		panic(err)
	}

	sqsClient = sqs.NewFromConfig(cfg)
	snsClient = sns.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)
	presignS3Client = s3.NewPresignClient(s3Client)
	dynamodbClient = dynamodb.NewFromConfig(cfg)
	lambdaClient = lambda.NewFromConfig(cfg)

	redisClient = redis.NewClient(&redis.Options{
		Addr: GetEnv("AWS_REDIS_HOST"), // master node
		DB:   1,
	})
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"msg": "Server is running!!!"})
	})
	http.HandleFunc("/s3", HandleS3Request)
	http.HandleFunc("/sqs", HandleSQSRequest)
	http.HandleFunc("/sns", HandleSNSRequest)
	http.HandleFunc("/dynamodb", HandleDynamoDBRequest)
	http.HandleFunc("/redis", HandleRedisRequest)
	http.HandleFunc("/lambda", HandleLambdaRequest)

	server := &http.Server{
		Addr:         serverPort,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Initialized server ...")
	go func() {
		defer wg.Done()
		log.Printf("Server listening on port %s ... \n", serverPort)
		if err := server.ListenAndServe(); err != nil {
			log.Fatal("Server error:", err)
			os.Exit(1)
		}
	}()
	wg.Wait()
}
