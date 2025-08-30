ENV_FILE=./.env
SOURCE_FILE=./cmd/logger.go
BIN_DIR=./cmd/bin/
INSTALL_PATH=/usr/local/bin/termlogger
CONFIG_DIR=$(HOME)/.termlogger
BIN_ENV_FILE=$(CONFIG_DIR)/.env
CACHE_PATH=$(CONFIG_DIR)/cache.db
TEST_CACHE=./testing/logs
BASH_RC=$(HOME)/.bashrc
ZSH_RC=$(HOME)/.zshrc
PROJECT_ROOT=$(shell pwd)
TERMLOGGER_HOOK_SCRIPT=./hooks/termlogger_hook.sh
REMOVE_HOOK_SCRIPT=./hooks/remove_hook.sh

.PHONY: all clean clean-cache clean-remote clean-test clean-proto config-dir env-setup help log-bin proto clean-proto remove-bin set-config remove-config remove-hook set-bin set-hook setup test-logdir uninstall
all: help

env-setup:
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "📋 No .env file found. Copying from .env.example..."; \
		cp .env.example $(ENV_FILE); \
		echo "✅ .env file created. Please configure it with your credentials."; \
	fi

set-config: env-setup
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

config-dir:
	@echo "🔧 Creating config directory at '$(CONFIG_DIR)'..."
	@mkdir -p "$(CONFIG_DIR)"
	@echo "✅ Config directory successfully created."
	@echo "$(PROJECT_ROOT)" > "$(CONFIG_DIR)/project_root"
	
test-logdir:
	@echo "🔧 Creating testing log directory at '$(TEST_CACHE)'..."
	@mkdir -p "$(TEST_CACHE)"
	@echo "✅ Test log directory successfully created."

set-hook:
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
	sed '/### >>> termlogger start >>>/,/### <<< termlogger end <<</d' "$$RC_FILE" > "$$TMP_FILE"; \
	mv "$$TMP_FILE" "$$RC_FILE"; \
	echo "🪝 Installing/updating hook in '$$RC_FILE'..."; \
	cat $(TERMLOGGER_HOOK_SCRIPT) >> "$$RC_FILE"; \
	echo "✅ Hook installed/updated."
	@rm -r $(BIN_DIR)

setup:
	@echo "🚀  Starting setup..."
	@$(MAKE) proto
	@$(MAKE) config-dir
	@$(MAKE) set-config
	@$(MAKE) migrate-up > /dev/null 2>&1
	@$(MAKE) set-bin
	@$(MAKE) test-logdir
	@$(MAKE) set-hook

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
			sed '/### >>> termlogger start >>>/,/### <<< termlogger end <<</d' "$$RC_FILE" > "$$TMP_FILE"; \
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

uninstall:
	@echo "🗑️  Starting uninstallation..."
	@$(MAKE) remove-bin
	@$(MAKE) remove-hook
	@$(MAKE) remove-config
	@$(MAKE) clean
	@echo "🎉 Uninstallation complete."
	
clean:
	@echo "🧼 Cleaning logs.."
	@$(MAKE) clean-cache
	@$(MAKE) clean-remote
	@$(MAKE) clean-test
	@$(MAKE) clean-proto

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

clean-remote:
	@echo "🗑️ Deleting contents of postgres DB";
	@$(MAKE) migrate-down > /dev/null 2>&1;
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
proto: 
	@buf generate

migrate-up:
	docker-compose run --rm migrate

migrate-down:
	docker-compose run --rm migrate down 1

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all             Shows this help message."
	@echo "  log-bin          Compiles the logger Go source file."
	@echo "  set-bin          Places the compiled binary into the user's binary folder."
	@echo "  config-dir       Creates the configuration directory."
	@echo "  test-log-dir      Creates the testing log directory."
	@echo "  set-hook         Installs the shell hook for termlogger."
	@echo "  setup           Builds and installs the termlogger binary and shell hook."
	@echo "  uninstall       Removes the termlogger binary, shell hook, and optionally the config directory."
	@echo "  remove-bin       Removes the installed binary."
	@echo "  remove-hook      Removes the shell hook."
	@echo "  remove-config Removes the configuration directory."
	@echo "  clean           Deletes main and testing log files."
	@echo "  clean-cache         Deletes main log files."
	@echo "  clean-test         Deletes testing log files."
	@echo "  proto           Builds proto files."
	@echo "  clean-proto      Removes proto files."
	@echo "  help            Shows this help message."