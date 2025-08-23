#!/usr/bin/env bash
set -e

SOURCE_FILE="./logger.go"
COMPILED_OUTPUT="logger"
INSTALL_PATH="/usr/local/bin/termlogger"
LOG_DIR="$HOME/.termlogger/logs"
BASH_RC="$HOME/.bashrc"
ZSH_RC="$HOME/.zshrc"

PROJECT_ROOT="$PWD"
CONFIG_DIR="$HOME/.termlogger"
mkdir -p "$CONFIG_DIR"
echo "$PROJECT_ROOT" > "$CONFIG_DIR/project_root"


# checks if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: 'go' command not found. Please install Go to continue."
    exit 1
fi

# checks if logger.go is present
if [ ! -f "$SOURCE_FILE" ]; then
    echo "❌ Error: Go source file '$SOURCE_FILE' not found."
    exit 1
fi

# builds logger.go
echo "📦 Compiling logger..."
if ! go build -o "$COMPILED_OUTPUT" "$SOURCE_FILE"; then
    echo "❌ Compilation failed."
    exit 1
fi
echo "✅ Compilation successful."

# cleans up temporary logging files
cleanup() {
    rm -f "$COMPILED_OUTPUT" "$TMP_FILE" || true
}
trap cleanup EXIT

echo "🚀 Installing binary in '$INSTALL_PATH'..."
sudo mkdir -p "$(dirname "$INSTALL_PATH")"
sudo cp "$COMPILED_OUTPUT" "$INSTALL_PATH"
echo "✅ Binary installed successfully."

# sets up logging directory
mkdir -p "$LOG_DIR"

echo "🔧 Creating test log directory in '$(cat "$CONFIG_DIR/project_root")/testing/logs' ..."
mkdir -p  "$(cat "$CONFIG_DIR/project_root")/testing/logs"
echo "✅ Directory created."

if [ -f "$ZSH_RC" ]; then
    RC_FILE="$ZSH_RC"
elif [ -f "$BASH_RC" ]; then
    RC_FILE="$BASH_RC"
else
    echo "❌ Neither ~/.bashrc nor ~/.zshrc detected. Setup failed."
    exit 1
fi

# shell hook that stores last command based on terminal type
HOOK=$(cat <<'EOF'

### >>> termlogger start >>>
_termlogger_hook() {
  
    local exit_code=$?
    local last_command

    # prevent infinite recursion
    if [[ "$_termlogger_running" == "1" ]]; then
        return
    fi
    _termlogger_running=1
    
    if [ -n "$ZSH_VERSION" ]; then
        last_command=$(fc -ln -1 2>/dev/null | sed 's/^[[:space:]]*//')
    elif [ -n "$BASH_VERSION" ]; then
        last_command=$(history 1 2>/dev/null | sed 's/^[[:space:]]*[0-9]*[[:space:]]*//')
    fi

    if [[ -n "$last_command" && "$last_command" != "_termlogger_hook"* ]]; then
        
        local ts="${EPOCHSECONDS:-$(date +%s)}"
        local hostname_val="${HOSTNAME:-$(hostname)}"
        local tty_val="${TTY:-$(tty 2>/dev/null)}"
        local termlogger_args=(
            --cmd="$last_command"
            --exit="$exit_code"
            --pid="$$"
            --uptime="$SECONDS"
            --cwd="$PWD"
            --oldpwd="$OLDPWD"
            --user="$USER"
            --euid="$EUID"
            --term="$TERM"
            --tty="$TTY"
    )

      if [[ -n "$EPOCHSECONDS" ]]; then
        termlogger_args+=(--ts "$EPOCHSECONDS")
      fi
      if [[ -n "$HOSTNAME" ]]; then
        termlogger_args+=(--hostname "$HOSTNAME")
      fi
      if [[ -n "$SSH_CLIENT" ]]; then
        termlogger_args+=(--ssh "$SSH_CLIENT")
      fi

      local is_git_repo="false"
      if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
        is_git_repo="true"
        termlogger_args+=(--gitrepo ) 
        termlogger_args+=(--gitroot "$(git rev-parse --show-toplevel 2>/dev/null)")
        termlogger_args+=(--gitbranch "$(git rev-parse --abbrev-ref HEAD 2>/dev/null)")
        termlogger_args+=(--gitcommit "$(git rev-parse --short HEAD 2>/dev/null)")
        termlogger_args+=(--gitstatus "$(git status --porcelain=v1 2>/dev/null)")
      fi

      (/usr/local/bin/termlogger "${termlogger_args[@]}" &>/dev/null &)
    fi
    
    _termlogger_running=0
}

if [ -n "$ZSH_VERSION" ]; then
    if [[ -z "${precmd_functions[(r)_termlogger_hook]}" ]]; then
        precmd_functions+=(_termlogger_hook)
    fi
elif [ -n "$BASH_VERSION" ]; then
    if [[ ! "$PROMPT_COMMAND" == *"_termlogger_hook"* ]]; then
        export PROMPT_COMMAND="_termlogger_hook;$PROMPT_COMMAND"
    fi
fi
### <<< termlogger end <<<
EOF
)

# create backup of RC file
cp "$RC_FILE" "$RC_FILE.backup.$(date +%s)"

TMP_FILE=$(mktemp)
# remove existing termlogger hooks
sed '/### >>> termlogger start >>>/,/### <<< termlogger end <<</d' "$RC_FILE" > "$TMP_FILE"
mv "$TMP_FILE" "$RC_FILE"

echo "🪝 Installing/updating hook in '$RC_FILE' ..."
echo "$HOOK" >> "$RC_FILE"
echo "✅ Hook installed/updated."