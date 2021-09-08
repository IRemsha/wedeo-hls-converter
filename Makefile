all: build down up

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker logs -f wedeo__hls_converter

bash:
	docker exec -it wedeo__hls_converter bash

run:
	go run /src/app/main.go

load:
	mc cp myphoto.jpg myminio/images
