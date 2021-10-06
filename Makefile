# Dev commands
up:
	docker-compose -f docker-compose.dev.yml up --build --force-recreate
down:
	docker-compose -f docker-compose.dev.yml down

# Prod commands
prod-remake-nginx:
	docker-compose -f docker-compose.yml -f compose-frontend-svelte.yml up -d --force-recreate --build nginx
prod-remake-refractor:
	docker-compose -f docker-compose.yml -f compose-frontend-svelte.yml up -d --force-recreate --build refractor
prod-remake-svelte:
	docker-compose -f docker-compose.yml -f compose-frontend-svelte.yml up -d --force-recreate --build refractor-frontend
deploy-cleanup:
	rm -f ./docker-compose.yml ./deploy/kratos/kratos.yml  ./deploy/postgres/init.sql ./deploy/nginx/app.conf \
	./Refractor-Svelte/rollup.config.js ./.neversetup 2> /dev/null