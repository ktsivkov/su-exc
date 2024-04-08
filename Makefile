up:
	docker compose -f deployment/docker-compose.yml up -d

down:
	docker compose -f deployment/docker-compose.yml down --volumes
