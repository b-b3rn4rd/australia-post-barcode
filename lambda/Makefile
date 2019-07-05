build:
	GOOS=linux GOARCH=amd64 go build -o handler .
	@zip -9 -r ./handler.zip handler
.PHONY: build

package: build
	aws cloudformation package \
		--template-file sam.yaml \
		--output-template-file sam.out.yaml \
		--s3-bucket ${S3_BUCKET_NAME} \
		--s3-prefix cfn
.PHONY: package

deploy: package
	aws cloudformation deploy \
		--template-file sam.out.yaml \
		--capabilities CAPABILITY_IAM \
		--stack-name 4-state-barcode-api
.PHONY: deploy