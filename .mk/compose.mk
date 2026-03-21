_compose:
	docker compose $(compose_opt) $(cmd) $(svc) $(opt) $(arg)
_compose_run-%:
	docker compose run --rm $(container_env_options) $(compose_opt) $(*) bash -i -c '$(cmd)'
_compose_exec-%:
	docker compose exec $(container_env_options) $(compose_opt) $(*) bash -i -c '$(cmd)'
_compose_exec_or_run-%:
	if docker ps | grep -q $(app)-$(*)-; then make _compose_exec-$(*); else make _compose_run-$(*); fi

up: upd logsf
upd:
	@make _compose cmd="up -d"
down: _down clean-docker
_down:
	@make _compose cmd=down
# restart: down upd logsf
restart:
	@make _compose cmd=restart svc=squid
logs:
	@make _compose cmd=logs
logsf:
	@make _compose cmd="logs -f"
build-images:
	@make _compose cmd=build
build-images-nocache:
	@make build-images opt=--no-cache
clean-docker:
	@docker_clean -e

console: console-backend-api
console-backend-api:
	@make _compose_exec_or_run-backend-api cmd=bash
