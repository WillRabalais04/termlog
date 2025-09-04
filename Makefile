# files
ENV_FILE=.env
BIN_ENV_FILE=$(CONFIG_DIR)/.env
SOURCE_FILE=./cmd/logger.go

BASH_RC=$(HOME)/.bashrc
ZSH_RC=$(HOME)/.zshrc

# directories
BIN_DIR=./cmd/bin/
CONFIG_DIR=$(HOME)/.termlogger
CACHE_PATH=$(CONFIG_DIR)/cache.db
INSTALL_PATH=/usr/local/bin/termlogger
PROJECT_ROOT=$(shell pwd)
TEST_CACHE=./cmd/test/logs

# hooks
TERMLOGGER_HOOK_SCRIPT=./hooks/termlogger_hook.sh
REMOVE_HOOK_SCRIPT=./hooks/remove_hook.sh

.PHONY: all clean clean-cache clean-remote clean-test clean-proto config-dir env-setup log-bin proto remove-bin set-config remove-config remove-hook set-bin set-hook setup setup-test setup-all test-logdir uninstall help wait-for-db wait-for-db-test start-db start-db-test check-docker
all: help

# --- BUILD/CONFIG ---
env-setup:
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "📋 No .env file found. Copying from .env.example..."; \
		cp .env.example $(ENV_FILE); \
		echo "✅ .env file created. Please configure it with your credentials."; \
	fi

config-dir:
	@echo "🔧 Creating config directory at '$(CONFIG_DIR)'..."
	@mkdir -p "$(CONFIG_DIR)"
	@echo "✅ Config directory successfully created."
	@echo "$(PROJECT_ROOT)" > "$(CONFIG_DIR)/project_root"

set-config: env-setup config-dir
	@echo "🔧 Installing user config to '$(BIN_ENV_FILE)'..."
	@cp $(ENV_FILE) $(BIN_ENV_FILE)
	@echo "✅ User config installed/updated."
	
log-bin:
	@echo "📦 Compiling logger..."
	@if ! go build -o $(BIN_DIR) $(SOURCE_FILE); then \
		echo "❌ Compilation failed."; \
		exit 1; \
	fi
	@echo "✅ Compilation successful."

set-bin: log-bin
	@echo "🚀 Installing binary in '$(INSTALL_PATH)'..."
	@sudo mkdir -p "$(shell dirname $(INSTALL_PATH))"
	@sudo cp "$(BIN_DIR)/logger" "$(INSTALL_PATH)"
	@echo "✅ Binary installed/updated."

set-hook: set-bin
	@if [ -f "$(ZSH_RC)" ]; then \
		RC_FILE="$(ZSH_RC)"; \
	elif [ -f "$(BASH_RC)" ]; then \
		RC_FILE="$(BASH_RC)"; \
	else \
		echo "❌ RC file not found. Hook not set."; \
		exit 1; \
	fi; \
	cp "$$RC_FILE" "$$RC_FILE.backup.$(shell date +%s)"; \
	TMP_FILE=$$(mktemp); \
	sed '/### >>> logger start >>>/,/### <<< logger end <<</d' "$$RC_FILE" > "$$TMP_FILE"; \
	mv "$$TMP_FILE" "$$RC_FILE"; \
	echo "🪝 Installing/updating hook in '$$RC_FILE'..."; \
	cat $(TERMLOGGER_HOOK_SCRIPT) >> "$$RC_FILE"; \
	echo "✅ Hook installed/updated."
	@rm -r $(BIN_DIR)

test-logdir:
	@echo "🔧 Creating testing log directory at '$(TEST_CACHE)'..."
	@mkdir -p "$(TEST_CACHE)"
	@echo "✅ Test log directory successfully created."

proto: 
	@echo "📦 Building proto files..."
	@buf generate

# --- SETUP ---
setup: migrate-up set-hook
	@echo "🎉 Development setup complete."

setup-test: migrate-up-test
	@echo "🎉 Test environment setup complete."
	
setup-all: setup setup-test

# --- UNINSTALL & CLEAN ---
remove-bin:
	@if [ -f "$(INSTALL_PATH)" ]; then \
		echo "🔐 Sudo required at $(INSTALL_PATH)."; \
		if sudo rm -f "$(INSTALL_PATH)"; then \
			echo "✅ Binary removed successfully."; \
		else \
			echo "❌ Failed to remove binary. Please check permissions."; \
			exit 1; \
		fi; \
	else \
		echo "🤔 Binary not found at $(INSTALL_PATH)."; \
		echo "✅ Nothing to remove!"; \
	fi

remove-hook:
	@RC_FILE=""; \
	if [ -f "$(ZSH_RC)" ]; then \
		RC_FILE="$(ZSH_RC)"; \
	elif [ -f "$(BASH_RC)" ]; then \
		RC_FILE="$(BASH_RC)"; \
	fi; \
	if [ -n "$$RC_FILE" ]; then \
		if [ -f "$(REMOVE_HOOK_SCRIPT)" ]; then \
			echo "🪝 Removing hook from $$RC_FILE..."; \
			. $(REMOVE_HOOK_SCRIPT); \
			_remove_hook "$$RC_FILE"; \
		else \
			echo "🪝 Removing hook manually from $$RC_FILE..."; \
			TMP_FILE=$$(mktemp); \
			sed '/### >>> logger start >>>/,/### <<< logger end <<</d' "$$RC_FILE" > "$$TMP_FILE"; \
			mv "$$TMP_FILE" "$$RC_FILE"; \
			echo "✅ Hook removed successfully."; \
		fi; \
	else \
		echo "❌ RC file not found. Cannot remove hook."; \
	fi

remove-config:
	@if [ -d "$(CONFIG_DIR)" ]; then \
		echo "⛔️ Found configuration and log directory at '$(CONFIG_DIR)'."; \
		read -p "❓ Do you want to permanently delete this directory? [y/n] "  -n 1 -r; \
		echo ""; \
		if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
			if rm -rf "$(CONFIG_DIR)"; then \
				echo "✅ Directory '$(CONFIG_DIR)' removed."; \
			else \
				echo "❌ Failed to remove '$(CONFIG_DIR)'."; \
			fi; \
		else \
			echo "👍 Okay, leaving '$(CONFIG_DIR)' untouched."; \
		fi; \
	fi

uninstall: remove-bin remove-hook remove-config clean-cache clean-test clean-proto stop-all-dbs
	@echo "🎉 Uninstallation complete."
	
clean: clean-cache clean-remote clean-test clean-proto
	@echo "🧼 Cleaning logs.."

clean-cache:
	@if [ -d "$(CACHE_PATH)" ]; then \
		read -p "❓ Do you want to delete all of the main logs? [y/n] " -n 1 -r; \
		echo ""; \
		if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
			echo "🗑️  Deleting contents of '$(CACHE_PATH)'..."; \
			@rm -r $(CACHE_PATH) \
			echo "✅ Deletion complete."; \
		else \
			echo "⏩ Skipping deletion."; \
		fi; \
	else \
		echo "🤔 Log cache '$(CACHE_PATH)' not found."; \
		echo "✅ Nothing to remove!"; \
	fi

clean-remote: migrate-down
	@echo "🗑️ Deleting contents of postgres DB";
	@echo "✅ Deletion complete.";

clean-test:
	@if [ -d "$(TEST_CACHE)" ]; then \
		read -p "❓ Do you want to delete all of the testing logs? [y/n] " -n 1 -r; \
		echo ""; \
		if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
			echo "🗑️  Deleting contents of '$(TEST_CACHE)'..."; \
			find "$(TEST_CACHE)" -mindepth 1 -delete; \
			echo "✅ Deletion complete."; \
		else \
			echo "⏩ Skipping deletion."; \
		fi; \
	else \
		echo "🤔 Test log directory '$(TEST_CACHE)' not found."; \
		echo "✅ Nothing to remove!"; \
	fi;

clean-proto:
	@rm -rf api/gen/*

# --- DOCKER ---
check-docker:
	@if ! docker info > /dev/null 2>&1; then \
		echo ""; \
		echo "❌ Docker is not running. Start Docker Desktop and run again."; \
		echo ""; \
		exit 1; \
	fi

start-db: check-docker
	@echo "🐘 Starting development database..."
	@docker-compose up -d db > /dev/null 2>&1
	@echo "✅ Postgres database is running."

stop-db: check-docker
	@echo "🐘 Stopping development database..."
	@docker-compose stop db > /dev/null 2>&1
	@echo "✅ Postgres database stopped."

start-db-test: check-docker
	@echo "🐘 Starting Postgres test database..."
	@docker-compose up -d db-test > /dev/null 2>&1
	@echo "✅ Postgres test database is running."

stop-db-test: check-docker
	@echo "🐘 Stopping test database..."
	@docker-compose stop db-test > /dev/null 2>&1
	@echo "✅ Postgres test database stopped."

stop-all-dbs: check-docker
	@echo "🐘 Stopping all services and removing containers, networks, and volumes..."
	@docker-compose down --volumes > /dev/null 2>&1

wait-for-db: start-db
	@echo "⏳ Waiting for the development database to be ready..."
	@until docker-compose exec db pg_isready -U postgres -q; do sleep 1; done
	@echo "✅ Development database is ready."

wait-for-db-test: start-db-test
	@echo "⏳ Waiting for the test database to be ready..."
	@until docker-compose exec db-test pg_isready -U test -q; do sleep 1; done
	@echo "✅ Test database is ready."

# --- DB MIGRATIONS ---
migrate-up: wait-for-db proto set-config
	@echo "🚀 Applying migrations to development database..."
	@docker-compose run --rm migrate up > /dev/null 2>&1

migrate-down: check-docker
	@echo "⏪ Reverting last migration on development database..."
	@docker-compose run --rm migrate down 1 > /dev/null 2>&1

migrate-up-test: wait-for-db-test proto test-logdir
	@echo "🚀 Applying migrations to test database..."
	@docker-compose run --rm migrate-test up > /dev/null 2>&1

migrate-down-test: check-docker
	@echo "⏪ Reverting last migration on test database..."
	@docker-compose run --rm migrate-test down 1 > /dev/null 2>&1

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Primary Targets:"
	@echo "  setup           Builds and installs the full development environment (binary, hooks, db)."
	@echo "  setup-test      Builds and prepares the test environment (db and migrations)."
	@echo "  setup-all       Runs both 'setup' and 'setup-test'."
	@echo "  uninstall       Removes the binary, hooks, configs, and stops all docker containers."
	@echo ""
	@echo "Development DB:"
	@echo "  start-db        Starts the development postgres container."
	@echo "  stop-db         Stops the development postgres container."
	@echo "  migrate-up      Applies all pending migrations to the development database."
	@echo "  migrate-down    Reverts the last migration on the development database."
	@echo ""
	@echo "Test DB:"
	@echo "  start-db-test   Starts the test postgres container."
	@echo "  stop-db-test    Stops the test postgres container."
	@echo "  migrate-up-test Applies all pending migrations to the test database."
	@echo "  migrate-down-test Reverts the last migration on the test database."
	@echo ""
	@echo "Other Targets:"
	@echo "  all             Shows this help message."
	@echo "  clean           Deletes main and testing log files and cleans remote DB."
	@echo "  help            Shows this help message."