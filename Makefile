PROJECT = gomic

build:
	cd src/main && docker build -t $(PROJECT):latest .

start:
	cd src/integration && docker-compose up

stop:
	cd src/integration && docker-compose down
