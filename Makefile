PROJECT = gomic

buildDocker:
	cd src/main && docker build -t $(PROJECT):latest .

integrationTestUp:
	cd src/integrationTest && docker-compose up

integrationTestDown:
	cd src/integrationTest && docker-compose down
