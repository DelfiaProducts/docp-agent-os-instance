package pkg

import (
	"os"
	"path/filepath"
)

// FileSystem is struct for execute operations in file system
type FileSystem struct{}

// NewFileSystem return instance of file system
func NewFileSystem() *FileSystem {
	return &FileSystem{}
}

// VerifyFileExist verify if file exist in host
func (fls *FileSystem) VerifyFileExist(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}
	return nil
}

// CreateOrUpdateSymlink creates or updates a symbolic link from linkPath to targetPath
func (fls *FileSystem) CreateOrUpdateSymlink(targetPath, linkPath string) error {
	// Remove existing symlink or file if it exists
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			return err
		}
	}
	err := os.Symlink(targetPath, linkPath)
	if err != nil {
		return err
	}
	return nil
}

// VerifyDirExistAndCreate verify if exist director and create
func (fls *FileSystem) VerifyDirExistAndCreate(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		errCrt := os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			return errCrt
		}
	}
	return nil
}

// GetHomeDir return home dir the host
func (fls *FileSystem) GetHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {

		return "", err
	}
	return homeDir, nil
}

// IsDirectory verify if exist directory
func (fls *FileSystem) IsDirectory(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// JoinPaths execute join with paths
func (fls *FileSystem) JoinPaths(paths []string) string {
	if len(paths) > 0 {
		return filepath.Join(paths...)
	}
	return ""
}

// VerifyDirExist verify if exist directory
func (fls *FileSystem) VerifyDirExist(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

// GetFileContent return content from file
func (fls *FileSystem) GetFileContent(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// CreateFile create file
func (fls *FileSystem) CreateFile(filePath string) error {
	f, err := os.OpenFile(filePath, os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

// CreatePathCompleted create dirs and file base
func (fls *FileSystem) CreatePathCompleted(filePath string) error {
	dirPath := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}
	finalPath := filepath.Join(dirPath, fileName)
	f, err := os.Create(finalPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

// CreateAndWriteFileContent create and write file
func (fls *FileSystem) CreateAndWriteFileContent(filePath string, content []byte) error {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

// WriteFileContent execute write the content for file
func (fls *FileSystem) WriteFileContent(filePath string, content []byte) error {
	if err := os.WriteFile(filePath, content, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// WriteBinaryContent execute write the binary for file
func (fls *FileSystem) WriteBinaryContent(filePath string, content []byte) error {
	if err := os.WriteFile(filePath, content, 0755); err != nil {
		return err
	}
	return nil
}
