package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"

	"github.com/b-b3rn4rd/4-state-barcode/src/australiapost"

	"github.com/aws/aws-lambda-go/lambda"
)

// HandleRequest handles AWS Lambda request - used using PROXY_LAMBDA integration
func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	b := bytes.Buffer{}
	r := &events.APIGatewayProxyResponse{}

	r.Headers = map[string]string{
		"Content-Type":                "application/json",
		"Access-Control-Allow-Origin": "*",
	}

	barcode, ok := request.QueryStringParameters["barcode"]
	if !ok {
		msg, _ := json.Marshal(map[string]string{"error": "Please provide the 'barcode' parameter"})
		r.StatusCode = 400
		r.Body = string(msg)
		return *r, nil
	}

	text := request.QueryStringParameters["text"]

	generator := australiapost.NewFourStateBarcode(barcode, &b, text)

	err := generator.Generate()
	if err != nil {
		msg, _ := json.Marshal(map[string]string{"error": err.Error()})
		r.StatusCode = 400
		r.Body = string(msg)
		return *r, nil
	}

	r.Body = base64.StdEncoding.EncodeToString(b.Bytes())
	r.StatusCode = 200
	r.Headers["Content-Type"] = "text/plain"
	return *r, nil
}

func main() {
	lambda.Start(HandleRequest)
}
