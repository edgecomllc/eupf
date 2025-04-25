package file

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/edgecomllc/eupf/cli/config"
	"github.com/edgecomllc/eupf/cli/domain"
	"gopkg.in/yaml.v3"
)

type FileRepository struct {
	backupDir string
}

func NewFileRepository(cfg *config.Config) domain.EupfLocalRepository {
	return &FileRepository{
		backupDir: cfg.BackupPath,
	}
}

func (r *FileRepository) BackupList(ctx context.Context) ([]domain.BackupRecord, error) {
	entry, err := os.ReadDir(r.backupDir)
	if err != nil {
		return nil, err
	}

	var backups []domain.BackupRecord
	for _, e := range entry {
		if e.IsDir() {
			continue
		}

		if !strings.HasSuffix(e.Name(), ".zip") {
			continue
		}

		info, _ := e.Info()
		name := strings.TrimSuffix(info.Name(), ".zip")

		backups = append(backups, domain.BackupRecord{
			Name:      name,
			Timestamp: info.ModTime(),
		})
	}

	return backups, nil
}

func (r *FileRepository) SaveUpfConfig(ctx context.Context, config *domain.UpfConfig) error {
	err := os.MkdirAll(r.backupDir, os.ModePerm)
	if err != nil {
		return err
	}

	timestamp := time.Now().Unix()
	zipPath := fmt.Sprintf("%s/%d.zip", r.backupDir, timestamp)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	configFile, err := zipWriter.Create("config.yaml")
	if err != nil {
		return err
	}

	configData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	_, err = configFile.Write(configData)
	return err
}

func (r *FileRepository) ReadUpfConfig(ctx context.Context, name string) (*domain.UpfConfig, error) {
	if !strings.HasSuffix(name, ".zip") {
		name = name + ".zip"
	}

	zipPath := fmt.Sprintf("%s/%s", r.backupDir, name)

	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()

	var configFile *zip.File
	for _, file := range zipReader.File {
		if file.Name == "config.yaml" {
			configFile = file
			break
		}
	}

	if configFile == nil {
		return nil, fmt.Errorf("config.yaml not found in archive %s", name)
	}

	rc, err := configFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	configData, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	var config domain.UpfConfig
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
