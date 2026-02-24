.PHONY: all build-api build-user build-article clean gen gen-kitex gen-hz docker-up docker-down

all: build-api build-user build-article

build-api:
	cd api && go build -o ../output/api-server .

build-user:
	cd service/user && go build -o ../../output/user-server .

build-article:
	cd service/article && go build -o ../../output/article-server .

gen: gen-kitex gen-hz

gen-kitex:
	cd kitex_gen && kitex -module byte.dance/kitex_gen -gen-path . ../idl/user.thrift
	cd kitex_gen && kitex -module byte.dance/kitex_gen -gen-path . ../idl/article.thrift
	cd kitex_gen && go mod tidy

gen-hz:
	cd api && hz update -idl ../idl/api.thrift

clean:
	rm -rf output/

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

run-user:
	cd service/user && go run .

run-article:
	cd service/article && go run .

run-api:
	cd api && go run .
