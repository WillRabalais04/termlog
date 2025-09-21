package database

import (
	"fmt"

	"github.com/WillRabalais04/terminalLog/db"
)

func GetLocalRepo(cachePath string) (*LogRepo, error) {
	cache, err := NewRepo(&Config{
		Driver:       "sqlite",
		DataSource:   cachePath,
		SchemaString: db.SqliteSchema,
	})
	if err != nil {
		return nil, fmt.Errorf("could not init cache repo (sqlite): %v", err)
	}
	return cache, nil
}

func GetRemoteRepo(dataSource string) (*LogRepo, error) {
	remote, err := NewRepo(&Config{
		Driver:       "pgx",
		DataSource:   dataSource,
		SchemaString: db.PostgresSchema,
	})
	if err != nil {
		return nil, fmt.Errorf("could not init remote repo (postgres): %v", err)
	}
	return remote, nil
}

func GetMultiRepo(cachePath, dataSource string) (*MultiRepo, error) {
	local, err := GetLocalRepo(cachePath)
	if err != nil {
		return nil, fmt.Errorf("multirepo init failed: %v", err)
	}
	remote, err := GetRemoteRepo(dataSource)
	if err != nil {
		return nil, fmt.Errorf("multirepo init failed: %v", err)
	}
	return NewMultiRepo(local, remote), nil
}
