IMAGE_NAME=dockerevaluator
CONTAINER_NAME=dockerevaluator-instance

all: build

build:
		docker build -t $(IMAGE_NAME) .

run: stop
		docker run --name $(CONTAINER_NAME) $(IMAGE_NAME)

stop:
		-docker stop $(CONTAINER_NAME) || true
		-docker rm $(CONTAINER_NAME) || true

clean: stop
		docker rmi $(IMAGE_NAME) || true

push:
		docker push $(IMAGE_NAME)

.PHONY: all build run stop clean list-images list-containers push
