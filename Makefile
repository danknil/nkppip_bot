.PHONY: default run_debug run_prod

default: run_debug

run_debug:
	COMPOSE_BAKE=true docker compose up --build
run_prod:
	docker compose up --build -e "APP_ENV=production"
