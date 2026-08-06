# Include custom targets and environment variables here

# Load local .env for make deploy / pluginctl (not committed to git).
ifneq (,$(wildcard ./.env))
include .env
export
endif
