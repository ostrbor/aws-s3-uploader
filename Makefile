.PHONY: build-and-push

tag = aws-s3-uploader
docker_registry = docker-registry.com:5000

build-and-push:
	docker build -t $(docker_registry)/$(tag) .
	docker push $(docker_registry)/$(tag)
