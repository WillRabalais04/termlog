CREATE TABLE IF NOT EXISTS logs (
  event_id TEXT PRIMARY KEY,
  command TEXT NOT NULL,
  exit_code INTEGER,
  ts INTEGER,          
  shell_pid INTEGER,
  shell_uptime INTEGER, 
  cwd TEXT,     
  prev_cwd TEXT,   
  user_name TEXT,    
  euid INTEGER,                
  term TEXT, 
  hostname TEXT, 
  ssh_client TEXT, 
  tty TEXT,
  git_repo INTEGER, 
  git_repo_root TEXT,  
  git_branch TEXT, 
  git_commit TEXT,  
  git_status TEXT,
  logged_successfully INTEGER  
);

-- indices for fuzzy search
CREATE INDEX IF NOT EXISTS idx_logs_command_lower ON logs (LOWER(command));
CREATE INDEX IF NOT EXISTS idx_logs_cwd_lower ON logs (LOWER(cwd));
CREATE INDEX IF NOT EXISTS idx_logs_prev_cwd_lower ON logs (LOWER(prev_cwd));
CREATE INDEX IF NOT EXISTS idx_logs_user_name_lower ON logs (LOWER(user_name));
CREATE INDEX IF NOT EXISTS idx_logs_term_lower ON logs (LOWER(term));
CREATE INDEX IF NOT EXISTS idx_logs_hostname_lower ON logs (LOWER(hostname));
CREATE INDEX IF NOT EXISTS idx_logs_ssh_client_lower ON logs (LOWER(ssh_client));
CREATE INDEX IF NOT EXISTS idx_logs_tty_lower ON logs (LOWER(tty));
CREATE INDEX IF NOT EXISTS idx_logs_git_repo_root_lower ON logs (LOWER(git_repo_root));
CREATE INDEX IF NOT EXISTS idx_logs_git_branch_lower ON logs (LOWER(git_branch));
CREATE INDEX IF NOT EXISTS idx_logs_git_commit_lower ON logs (LOWER(git_commit));
CREATE INDEX IF NOT EXISTS idx_logs_git_status_lower ON logs (LOWER(git_status));