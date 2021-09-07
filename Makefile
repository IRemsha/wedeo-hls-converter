all: build down up

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker logs -f app

bash:
	docker exec -it app bash

run:
	go run /src/app/main.go

load:
	mc cp myphoto.jpg myminio/images
