.PHONY: default run_debug run_prod

default: run_debug

run_debug:
	docker compose up --build
run_prod:
	COMPOSE_BAKE=true docker compose up --build -e "APP_ENV=production"
