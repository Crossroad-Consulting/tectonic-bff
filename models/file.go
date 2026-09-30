package models

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type File struct {
	Name string
	Size int64
	Path string
}

type FileService struct {
	BasePath     string
	GeneralDir   string
	SafetyDir    string
	SchedulesDir string
}

func (service *FileService) ListGeneral() ([]File, error) {
	glob := path.Join(service.BasePath, service.GeneralDir, "*")
	files, err := service.listFiles(glob)
	if err != nil {
		return nil, fmt.Errorf("list general files: %w", err)
	}

	return files, nil
}

func (service *FileService) ListSafety() ([]File, error) {
	glob := path.Join(service.BasePath, service.SafetyDir, "*")
	files, err := service.listFiles(glob)
	if err != nil {
		return nil, fmt.Errorf("list safety files: %w", err)
	}

	return files, nil
}

func (service *FileService) ListSchedules(SFID string) ([]File, error) {
	glob := path.Join(service.BasePath, service.SchedulesDir, SFID+"-*")
	files, err := service.listFiles(glob)
	if err != nil {
		return nil, fmt.Errorf("list schedule files: %w", err)
	}

	return files, nil
}

func (service *FileService) listFiles(glob string) ([]File, error) {
	var files []File

	allFiles, err := filepath.Glob(glob)
	if err != nil {
		return nil, fmt.Errorf("list files for glob %q: %w", glob, err)
	}

	for _, file := range allFiles {
		info, err := os.Stat(file)
		if err != nil {
			log.Printf("failed to stat file %q: %v", file, err)
			continue
		}

		if !info.IsDir() {
			files = append(files, File{
				Name: info.Name(),
				Size: info.Size(),
				Path: file,
			})
		}
	}

	return files, nil
}

func (service *FileService) GetGeneral(filename string) (File, error) {
	filePath := path.Join(service.BasePath, service.GeneralDir, filename)
	file, err := service.getFile(filePath)
	if err != nil {
		return File{}, fmt.Errorf("get general file: %w", err)
	}

	return file, nil
}

func (service *FileService) GetSafety(filename string) (File, error) {
	filePath := path.Join(service.BasePath, service.SafetyDir, filename)
	file, err := service.getFile(filePath)
	if err != nil {
		return File{}, fmt.Errorf("get safety file: %w", err)
	}

	return file, nil
}

func (service *FileService) GetSchedule(SFID string, filename string) (File, error) {
	if !strings.HasPrefix(filename, SFID) {
		return File{}, fmt.Errorf("get schedule file: %w", ErrUnauthorized)
	}

	filePath := path.Join(service.BasePath, service.SchedulesDir, filename)
	file, err := service.getFile(filePath)
	if err != nil {
		return File{}, fmt.Errorf("get schedule file: %w", err)
	}

	return file, nil
}

func (service *FileService) getFile(filePath string) (File, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return File{}, ErrNotFound
		}

		return File{}, fmt.Errorf("querying for file: %w", err)
	}

	return File{
		Name: info.Name(),
		Size: info.Size(),
		Path: filePath,
	}, nil
}
