default:
	docker-compose build --force-rm --no-cache 
	docker-compose up -d
	docker ps
	sleep 5
	docker logs ocrmypdf-watchdog

clean:
	docker system prune

	
