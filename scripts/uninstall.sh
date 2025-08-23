#!/usr/bin/env bash
set -e

INSTALL_PATH="/usr/local/bin/termlogger"
CONFIG_DIR="$HOME/.termlogger"
BASH_RC="$HOME/.bashrc"
ZSH_RC="$HOME/.zshrc"

echo "🗑️ Starting uninstallation..."

if [ -f "$INSTALL_PATH" ]; then
  echo "🔐 Sudo required at $INSTALL_PATH."
  if sudo rm -f "$INSTALL_PATH"; then
    echo "✅ Binary removed successfully."
  else
    echo "❌ Failed to remove binary. Please check permissions."
    exit 1
  fi
else
  echo "🤔 Binary not found at $INSTALL_PATH (already removed)."
fi

remove_hook() {
  local rc_file=$1
  if [ -f "$rc_file" ]; then
    echo "🧹 Checking $rc_file for the hook..."
    local tmp_file=$(mktemp)
    sed '/### >>> logger start >>>/,/### <<< logger end <<</d' "$rc_file" > "$tmp_file"
    
    if ! cmp -s "$rc_file" "$tmp_file"; then
        mv "$tmp_file" "$rc_file"
        echo "✅ Hook removed from $rc_file."
    else
        echo "🤔 Hook not found in $rc_file."
        rm "$tmp_file"
    fi
  fi
}

remove_hook "$BASH_RC"
remove_hook "$ZSH_RC"

if [ -d "$CONFIG_DIR" ]; then
  echo "⚠️ Found configuration and log directory at $CONFIG_DIR."
  read -p "Do you want to permanently delete this directory? [y/N] " response
  
  if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    if rm -rf "$CONFIG_DIR"; then
      echo "✅ Directory $CONFIG_DIR removed."
    else
      echo "❌ Failed to remove $CONFIG_DIR."
    fi
  else
    echo "👍 Okay, leaving $CONFIG_DIR untouched."
  fi
fi

RC_FILE="'~/.your_shell_rc_file'"
if [ -f "$ZSH_RC" ]; then
    RC_FILE="$ZSH_RC"
elif [ -f "$BASH_RC" ]; then
    RC_FILE="$BASH_RC"
fi
echo -e "🎉 Uninstallation complete. Please restart your shell or run source "$RC_FILE" for changes to take effect."