
# build & start
default:
	docker compose up -d --force-recreate --build
	
start:
	docker compose up -d

stop:
	docker compose down

build:
	docker compose build --force-rm --no-cache 

clean:
	docker system prune

	
