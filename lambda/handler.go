package main

import (
	"bytes"
	"context"

	"github.com/b-b3rn4rd/4-state-barcode/src/australiapost"

	"github.com/aws/aws-lambda-go/lambda"
)

type Barcode struct {
	Barcode string `json:"barcode"`
	Text    string `json:"text"`
}

func HandleRequest(ctx context.Context, barcode Barcode) (string, error) {
	b := bytes.Buffer{}

	generator := australiapost.NewFourStateBarcode(barcode.Barcode, &b, barcode.Text)

	err := generator.Generate()
	if err != nil {
		return "", nil
	}

	s := b.String()

	return s, nil
}

func main() {
	lambda.Start(HandleRequest)
}
