.PHONY: templ
templ:
	templ generate --watch --proxy http://localhost:8080

mongo:
	docker run --name mongodb -p 27017:27017 -d mongo

serve:
	air

tailwind:
	npx tailwindcss -c ./tailwind.config.js -i ./public/styles.css -o ./public/output.css --watch
