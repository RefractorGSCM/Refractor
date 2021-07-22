up:
	docker-compose -f docker-compose.dev.yml up --build --force-recreate
down:
	docker-compose -f docker-compose.dev.yml down