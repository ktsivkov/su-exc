up:
	docker compose -f deployment/docker-compose.yml up -d --remove-orphans

down:
	docker compose -f deployment/docker-compose.yml down --volumes
