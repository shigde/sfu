package storage

import "fmt"

type StorageConfig struct {
	Name         string `mapstructure:"name"`
	DataSource   string `mapstructure:"dataSource"`
	LoadFixtures bool   `mapstructure:"loadFixtures"`
}

func ValidateStorageConfig(config *StorageConfig) error {

	if config.Name != "sqlite3" {
		return fmt.Errorf("store.name currently supportes only Sqlite3")
	}

	if len(config.DataSource) == 0 {
		return fmt.Errorf("store.dataSource should not be empty")
	}

	return nil

}
