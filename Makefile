.PHONY: templ
templ:
	templ generate --watch --proxy http://localhost:8080

mongo:
	docker-compose up -d

serve:
	air

tailwind:
	npx tailwindcss -c ./tailwind.config.js -i ./public/styles.css -o ./public/output.css --watch

format:
	gofmt -w .
	templ fmt .
