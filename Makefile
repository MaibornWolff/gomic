PROJECT = gomic

build:
	cd src/main && docker build -t $(PROJECT):latest .

start: build
	cd src/integration && docker-compose up

stop:
	cd src/integration && docker-compose down
