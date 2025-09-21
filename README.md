# termlog
- add a shell hook that logs all of your terminal commands in either a local sqlite database or in a remote postgres database.
- the remote db can be accessed by api
# setup:
- 'make setup'
- source your .zshrc or .bashrc
- configure the env file to meet your needs
- to pick a mode see the .env file
# local mode
- local mode stores your logs in a local sqlite db located at '$HOME/.termlogger/cache.db'
# org mode 
- org mode allows you to host your data on a postgres server and you can access it via api
- has local cache of logs stored at $HOME/.termlogger/cache.db if logs can't be pushed to remote
- local stores the logs in a sqlite 
- to run in org-mode run 'make start-server' which builds and starts the server
# server commands
- 'make start-server' builds and runs the server
- once the server is running, to see the server interactions in real time run 'make logs-server'
- 'make stop-server' stops the server
# uninstall: 
- 'make uninstall'
# clean 
- to clean logs run 'make clean' or see 'make help' for more details
# testing: 
the testing approach is broadly speaking to have two types of tests — unit/integration tests and a test harness
- unit/integration tests are located at ./test/
 - 'go test ./test/...' or 'go test -v ./test/...'
- test harness are located at ./cmd/test/
    - before using test harness run 'make start-db-test' and after 'make stop-db-test'
    - run with 'go run cmd/test/*' eg. 'go run cmd/test/dummy-repo/main.go'
- you can log outputs to a json file at test/logs/logs.json for debugging by running 'log-json-start' or 'logs-json-stop' in any terminal
# pause logging
- you can pause and resume logging by entering 'log-pause' and 'log-resume' respectively in your terminal anywhere
# help
to see more run 'make help'