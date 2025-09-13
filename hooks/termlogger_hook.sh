### >>> logger start >>>
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

      local it_repo="false"
      if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
        git_repo="true"
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
### <<< logger end <<<