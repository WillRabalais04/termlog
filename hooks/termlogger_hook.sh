
### >>> logger start >>>
PAUSE_FILE="$HOME/.termlogger/.paused"
JSON_FILE="$HOME/.termlogger/.json"

log-pause() {
  if [ -f "$PAUSE_FILE" ]; then
    echo "⏸ Terminal logging is already PAUSED."
  else
    touch "$PAUSE_FILE"
    echo "⏸ Terminal logging is now PAUSED."
  fi
}
log-resume() {
  if [ -f "$PAUSE_FILE" ]; then
    rm -f "$PAUSE_FILE"
    echo "▶️ Terminal logging is now RESUMED."
  else
    echo "▶️ Terminal logging is already IN PROGRESS."
  fi
}
log-json-start() {
  if [ -f "$JSON_FILE" ]; then
    echo "📄 Logging terminal commands to json file is already IN PROGRESS."
  else
    touch "$JSON_FILE"
    echo "📄 Logging terminal commands to json file."
  fi
}
log-json-stop() {
  if [ -f "$JSON_FILE" ]; then
    rm -f "$JSON_FILE"
    echo "📄 Logging terminal commands to json file now STOPPED."
  else
    echo "📄 Logging terminal commands to json file is already STOPPED."
  fi
}

_termlogger_hook() {
    if [ -f "$PAUSE_FILE" ]; then
        return
    fi

    local exit_code=$?
    local last_command
    local log_dir="$HOME/.termlogger"
    local log_file="$log_dir/bin.log"
    mkdir -p "$log_dir"

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

    if [[ -n "$last_command" && "$last_command" != "_termlogger_hook"* && "$last_command" != "log-pause"* && "$last_command" != "log-resume"* && "$last_command" != "log-json-start"* && "$last_command" != "log-json-stop"* ]]; then        
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
      if [ -f "$JSON_FILE" ]; then
        termlogger_args+=(--json)
      fi

      local it_repo="false"
      if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
        git_repo="true"
        termlogger_args+=(--gitrepo ) 
        termlogger_args+=(--gitroot "$(git rev-parse --show-toplevel 2>/dev/null)")
        termlogger_args+=(--gitbranch "$(git rev-parse --abbrev-ref HEAD 2>/dev/null)")
        termlogger_args+=(--gitcommit "$(git rev-parse --short HEAD 2>/dev/null)")
        termlogger_args+=(--gitstatus "$(git status --porcelain=v1 2>/dev/null)")
      fi
      ( /usr/local/bin/termlogger "${termlogger_args[@]}" &>> "$log_file" & )
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
### <<< logger end <<<