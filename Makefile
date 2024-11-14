.PHONY: up down logs ps clean test

# Start all services
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# List running services
ps:
	docker-compose ps

# Clean volumes
clean:
	docker-compose down -v

# Run tests
test:
	go test ./... -v

# Start specific service
start-%:
	docker-compose up -d $*

# Stop specific service
stop-%:
	docker-compose stop $*

# Restart specific service
restart-%:
	docker-compose restart $*

# View logs for specific service
logs-%:
	docker-compose logs -f $*

# Initialize database
init-db:
	docker-compose exec mysql mysql -uroot -ppassword -e "CREATE DATABASE IF NOT EXISTS goProject;"

# Connect to MySQL CLI
mysql-cli:
	docker-compose exec mysql mysql -uroot -ppassword goProject

# Connect to Redis CLI
redis-cli:
	docker-compose exec redis redis-cli -a password 