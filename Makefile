BUILD_ID ?= 1
BUILD_SHA1 = $(shell git rev-parse --short=7 --verify HEAD)
BUILD_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
REPOSITORY ?= bernard/4-state-barcode
GITHUB_REPOSITORY ?= b-b3rn4rd/australia-post-barcode
MAJOR ?= 1
MINOR ?= 0
ifeq ($(BUILD_BRANCH),master)
	PATCH = $(BUILD_ID)
else
	PATCH = $(BUILD_ID)-$(BUILD_SHA1)
endif

IMAGE_TAG := $(MAJOR).$(MINOR).$(PATCH)

ci: push
.PHONY: ci

clean:
	REPOSITORY=$(REPOSITORY) \
	IMAGE_TAG=$(IMAGE_TAG) \
	docker-compose --project-name barcode down || true
.PHONY: clean

build: clean
	REPOSITORY=$(REPOSITORY) \
	IMAGE_TAG=$(IMAGE_TAG) \
	docker-compose --project-name barcode up
	docker cp 4-state-barcode-${IMAGE_TAG}:/tmp/release.zip .

	@curl -s \
		--data-binary @release.zip  \
		-H "Content-Type: application/zip" \
		"https://uploads.github.com/repos/$(GITHUB_REPOSITORY)/releases/$$(curl -s \
			--data "{\"tag_name\": \"$(IMAGE_TAG)\"}" \
			"https://api.github.com/repos/$(GITHUB_REPOSITORY)/releases?access_token=${GITHUB_TOKEN}" | jq '.id')/assets?name=release-$(IMAGE_TAG).zip&access_token=${GITHUB_TOKEN}"
	@echo successfully built docker image
.PHONY: build

push: build
	@docker login --username=${DOCKER_USERNAME} --password=${DOCKER_PASSWORD}

	REPOSITORY=$(REPOSITORY) \
	IMAGE_TAG=$(IMAGE_TAG) \
	docker-compose --project-name barcode push barcode
.PHONY: push

latest:
	@docker login --username=${DOCKER_USERNAME} --password=${DOCKER_PASSWORD}
	docker tag ${REPOSITORY}:$(IMAGE_TAG) ${REPOSITORY}:latest
	docker push ${REPOSITORY}:latest
.PHONY: latest