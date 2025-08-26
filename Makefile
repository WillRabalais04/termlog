SOURCE_FILE=./logger.go
COMPILED_OUTPUT=./bin/logger
INSTALL_PATH=/usr/local/bin/termlogger
CONFIG_DIR=$(HOME)/.termlogger
LOG_DIR=$(CONFIG_DIR)/logs
TEST_LOG_DIR=./testing/logs
BASH_RC=$(HOME)/.bashrc
ZSH_RC=$(HOME)/.zshrc
PROJECT_ROOT=$(shell pwd)
TERMLOGGER_HOOK_SCRIPT=./hooks/termlogger_hook.sh
REMOVE_HOOK_SCRIPT=./hooks/remove_hook.sh

.PHONY: all clean cleanml cleantl configdir help logbin logdir proto protoclean removebin removeconfigdir removehook setbin sethook setup testlogdir uninstall
all: help

logbin:
	@echo "📦 Compiling logger..."
	@if ! go build -o $(COMPILED_OUTPUT) $(SOURCE_FILE); then \
		echo "❌ Compilation failed."; \
		exit 1; \
	fi
	@echo "✅ Compilation successful."

setbin: logbin
	@echo "🚀 Installing binary in '$(INSTALL_PATH)'..."
	@sudo mkdir -p "$(shell dirname $(INSTALL_PATH))"
	@sudo cp "$(COMPILED_OUTPUT)" "$(INSTALL_PATH)"
	@echo "✅ Binary installed successfully."

configdir:
	@echo "🔧 Creating config directory at '$(CONFIG_DIR)'..."
	@mkdir -p "$(CONFIG_DIR)"
	@echo "✅ Config directory successfully created."
	@echo "$(PROJECT_ROOT)" > "$(CONFIG_DIR)/project_root"
	@$(MAKE) logdir

logdir:
	@echo "🔧 Creating log directory at '$(LOG_DIR)'..."
	@mkdir -p "$(LOG_DIR)"
	@echo "✅ Log directory successfully created."

testlogdir:
	@echo "🔧 Creating testing log directory at '$(TEST_LOG_DIR)'..."
	@mkdir -p "$(TEST_LOG_DIR)"
	@echo "✅ Test log directory successfully created."

sethook:
	@if [ -f "$(ZSH_RC)" ]; then \
		RC_FILE="$(ZSH_RC)"; \
	elif [ -f "$(BASH_RC)" ]; then \
		RC_FILE="$(BASH_RC)"; \
	else \
		echo "❌ RC file not found. Hook not set."; \
		exit 1; \
	fi; \
	echo "Found shell configuration file at '$$RC_FILE'"; \
	cp "$$RC_FILE" "$$RC_FILE.backup.$(shell date +%s)"; \
	TMP_FILE=$$(mktemp); \
	sed '/### >>> termlogger start >>>/,/### <<< termlogger end <<</d' "$$RC_FILE" > "$$TMP_FILE"; \
	mv "$$TMP_FILE" "$$RC_FILE"; \
	echo "🪝 Installing/updating hook in '$$RC_FILE' ..."; \
	cat $(TERMLOGGER_HOOK_SCRIPT) >> "$$RC_FILE"; \
	echo "✅ Hook installed/updated."
	@rm -f $(COMPILED_OUTPUT)

setup:
	@echo "🚀  Starting setup..."
	@$(MAKE) proto
	@$(MAKE) configdir
	@$(MAKE) setbin
	@$(MAKE) testlogdir
	@$(MAKE) sethook

removebin:
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

removehook:
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

removeconfigdir:
	@if [ -d "$(CONFIG_DIR)" ]; then \
		echo "⛔️ Found configuration and log directory at '$(CONFIG_DIR)'."; \
		read -p "❓ Do you want to permanently delete this directory? [y/n] " response; \
		if [ "$$response" = "y" ] || [ "$$response" = "Y" ] || [ "$$response" = "yes" ] || [ "$$response" = "YES" ]; then \
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
	@$(MAKE) removebin
	@$(MAKE) removehook
	@$(MAKE) removeconfigdir
	@$(MAKE) cleantl
	@echo "🎉 Uninstallation complete."
	
clean:
	@echo "🧼 Cleaning logs.."
	@$(MAKE) cleanml
	@$(MAKE) cleantl
	@$(MAKE) protoclean

cleanml:
	@if [ -d "$(LOG_DIR)" ]; then \
		read -p "❓ Do you want to delete all of the main logs? (y/n) " -n 1 -r; \
		echo ""; \
		if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
			echo "🗑️  Deleting contents of '$(LOG_DIR)'..."; \
			find "$(LOG_DIR)" -mindepth 1 -delete; \
			echo "✅ Deletion complete."; \
		else \
			echo "⏩ Skipping deletion."; \
		fi; \
	else \
		echo "🤔 Log directory '$(LOG_DIR)' not found."; \
		echo "✅ Nothing to remove!"; \
	fi

cleantl:
	@if [ -d "$(TEST_LOG_DIR)" ]; then \
		read -p "❓ Do you want to delete all of the testing logs? (y/n) " -n 1 -r; \
		echo ""; \
		if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
			echo "🗑️  Deleting contents of '$(TEST_LOG_DIR)'..."; \
			find "$(TEST_LOG_DIR)" -mindepth 1 -delete; \
			echo "✅ Deletion complete."; \
		else \
			echo "⏩ Skipping deletion."; \
		fi; \
	else \
		echo "🤔 Test log directory '$(TEST_LOG_DIR)' not found."; \
		echo "✅ Nothing to remove!"; \
	fi

proto: 
	@buf generate

protoclean:
	@rm -rf api/gen/*

migrate-up:
	docker-compose run --rm migrate

migrate-down:
	docker-compose run --rm migrate down 1

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all             Shows this help message."
	@echo "  logbin          Compiles the logger Go source file."
	@echo "  setbin          Places the compiled binary into the user's binary folder."
	@echo "  configdir       Creates the configuration directory."
	@echo "  logdir          Creates the main log directory."
	@echo "  testlogdir      Creates the testing log directory."
	@echo "  sethook         Installs the shell hook for termlogger."
	@echo "  setup           Builds and installs the termlogger binary and shell hook."
	@echo "  uninstall       Removes the termlogger binary, shell hook, and optionally the config directory."
	@echo "  removebin       Removes the installed binary."
	@echo "  removehook      Removes the shell hook."
	@echo "  removeconfigdir Removes the configuration directory."
	@echo "  clean           Deletes main and testing log files."
	@echo "  cleanml         Deletes main log files."
	@echo "  cleantl         Deletes testing log files."
	@echo "  proto           Builds proto files."
	@echo "  protoclean      Removes proto files."
	@echo "  help            Shows this help message."