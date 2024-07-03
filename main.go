package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	SourceDir        string `env:"SOURCE_DIR"        env-required:"true"`
	ImageDir         string `env:"IMAGE_DIR"         env-required:"true"`
	VideoDir         string `env:"VIDEO_DIR"         env-required:"true"`
	MusicDir         string `env:"MUSIC_DIR"         env-required:"true"`
	DocumentsDir     string `env:"DOCUMENTS_DIR"     env-required:"true"`
	PresentationsDir string `env:"PRESENTATIONS_DIR" env-required:"true"`
	TablesDir        string `env:"TABLES_DIR"        env-required:"true"`
	TextDir          string `env:"TEXT_DIR"          env-required:"true"`
	ArchiveDir       string `env:"ARCHIVE_DIR"       env-required:"true"`
	EXEDir           string `env:"EXE_DIR"           env-required:"true"`
	OtherDir         string `env:"OTHER_DIR"         env-required:"true"`
	ScriptName       string `env:"SCRIPT_NAME"       env-required:"true"`
}

var (
	imageExtensions = []string{
		".jpg",
		".png",
		".svg",
		".jpeg",
		".gif",
		".tiff",
		".tif",
		".ico",
		".cur",
		".bmp",
		".raw",
		".jfif",
		".pjpeg",
		".pjp",
	}
	videoExtensions = []string{
		".mp4",
		".mkv",
		".webm",
		".flv",
		".vob",
		".ogg",
		".ogv",
		".drc",
		".avi",
		".mng",
		".mov",
		".qt",
		".wmv",
		".m4p",
		".m4v",
		".mpg",
		".mpeg",
		".mp2",
		".m2v",
		".m4v",
	}
	musicExtensions         = []string{".mp3", ".wav", ".mid", ".midi"}
	documentsExtensions     = []string{".docx", ".doc", ".pdf"}
	presentationsExtensions = []string{".ppt", ".pptx"}
	tablesExtensions        = []string{".xls", ".xlsx", ".csv"}
	textExtensions          = []string{".txt", ".TXT"}
	exeExtensions           = []string{".exe", ".msi"}
	archiveExtensions       = []string{".zip", ".rar", ".7z", ".gz"}
)

func getDirName(path string) string {
	return path[strings.LastIndex(path, "\\")+1:]
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err.Error())
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatal(err.Error())
	}

	dirExts := make(map[string][]string)
	dirExts[cfg.ImageDir] = imageExtensions
	dirExts[cfg.VideoDir] = videoExtensions
	dirExts[cfg.MusicDir] = musicExtensions
	dirExts[cfg.DocumentsDir] = documentsExtensions
	dirExts[cfg.PresentationsDir] = presentationsExtensions
	dirExts[cfg.TablesDir] = tablesExtensions
	dirExts[cfg.TextDir] = textExtensions
	dirExts[cfg.ArchiveDir] = archiveExtensions
	dirExts[cfg.EXEDir] = exeExtensions

	if err := filepath.Walk(cfg.SourceDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		switch info.Name() {
		case cfg.ScriptName:
			return nil
		case "Telegram Desktop":
			return filepath.SkipDir
		case getDirName(cfg.DocumentsDir):
			return filepath.SkipDir
		case getDirName(cfg.PresentationsDir):
			return filepath.SkipDir
		case getDirName(cfg.TablesDir):
			return filepath.SkipDir
		case getDirName(cfg.TextDir):
			return filepath.SkipDir
		case getDirName(cfg.ArchiveDir):
			return filepath.SkipDir
		case getDirName(cfg.EXEDir):
			return filepath.SkipDir
		case getDirName(cfg.OtherDir):
			return filepath.SkipDir
		}

		if info.IsDir() {
			return moveToOtherDir(info, cfg.SourceDir, cfg.OtherDir)
		}

		isMoved := false

		for key, value := range dirExts {
			fmt.Printf("Searching files with %v\n", value)
			isMoved, err = moveFile(info, cfg.SourceDir, key, value)
			if isMoved {
				return nil
			}
			if err != nil {
				return err
			}
		}

		return moveToOtherDir(info, cfg.SourceDir, cfg.OtherDir)
	}); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Println("SUCCESS")
}

func moveFile(
	info fs.FileInfo,
	source string,
	destination string,
	extensions []string,
) (bool, error) {
	fileMoved := false
	for _, ext := range extensions {
		if strings.HasSuffix(info.Name(), ext) {
			fileMoved = true
			filename := info.Name()
			src := filepath.Join(source, filename)
			dst := filepath.Join(destination, filename)
			if _, err := os.Stat(dst); err == nil {
				base := strings.TrimSuffix(filename, ext)
				newFilename := fmt.Sprintf("%s - copy%s", base, ext)
				dst = filepath.Join(destination, newFilename)
			}

			err := os.Rename(src, dst)
			if err != nil {
				return fileMoved, err
			}

			fmt.Printf("File moved to %s: %s\n", destination, filename)
		}
	}

	return fileMoved, nil
}

func moveToOtherDir(info fs.FileInfo, source string, destination string) error {
	name := info.Name()
	src := filepath.Join(source, name)
	dst := filepath.Join(destination, name)

	if _, err := os.Stat(src); os.IsNotExist(err) {
		fmt.Printf("Not exist: %s\n", src)
		return nil
	}

	if _, err := os.Stat(dst); err == nil {
		if info.IsDir() {
			dst = filepath.Join(destination, fmt.Sprintf("%s - copy", name))
		} else {
			ext := filepath.Ext(name)
			dst = filepath.Join(
				destination,
				fmt.Sprintf("%s - copy%s", strings.TrimSuffix(name, ext), ext),
			)
		}
	}

	if info.IsDir() {
		if err := os.MkdirAll(dst, os.ModePerm); err != nil {
			return err
		}

		if err := copyDir(src, dst); err != nil {
			return err
		}

		if err := os.RemoveAll(src); err != nil {
			return err
		}

		fmt.Printf("Dir moved to %s: %s\n", destination, name)
		return filepath.SkipDir
	}

	err := os.Rename(src, dst)
	if err != nil {
		return err
	}

	fmt.Printf("File moved to %s: %s\n", destination, name)

	return nil
}

func copyDir(src, dst string) error {
	err := os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcPath := filepath.Join(src, file.Name())
		dstPath := filepath.Join(dst, file.Name())

		if file.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
