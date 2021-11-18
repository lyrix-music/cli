package service

import (
	_ "embed"
	"fmt"
	"github.com/lyrix-music/cli/meta"
	"os"
	"path/filepath"
)

//go:embed lyrix-windows_exporter.exe
var f []byte

func Run() {
	exe, err := os.Executable()
	if err != nil {
		logger.Fatal("Couldn't resolve base directory")
		panic(err)
	}
	data := filepath.Join(filepath.Dir(exe), fmt.Sprintf("lyrix-windows_exporter%s.exe", meta.BuildVersion))
	if _, err := os.Stat(data); os.IsNotExist(err) {
		_ = os.WriteFile(data, f, 0755)
	}

}
