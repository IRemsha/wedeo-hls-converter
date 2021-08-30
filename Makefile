all: build down up

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

run:
	go run main.go

load:
	mc cp myphoto.jpg myminio/images
