
define setup_env
	$(eval ENV_FILE := .env.$(1))
	$(eval include $(ENV_FILE))
	$(eval export)
endef

.PHONY: local
local:
	$(call setup_env,local)
	go run cmd/app/main.go

.PHONY: rabbitserver
rabbitserver:
	docker-compose up -d
