BUILD_DIR := .aws-sam/build
SAM_TEMPLATE := template.yaml

# Commands
SAM := sam

# Targets
.PHONY: build rim deploy

# Build the SAM application
build:
	$(SAM) build --template-file $(SAM_TEMPLATE) --build-dir $(BUILD_DIR)

# Run locally with SAM on port 4000
run: build
	$(SAM) local start-api --port 4000

# deploy to AWS
deploy: build
	$(SAM) deploy