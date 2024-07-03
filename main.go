package main

type Config struct {
	SourceDir string `env:"SOURCE_DIR" env-required:"true"`
	ImageDir  string `env:"IMAGE_DIR" env-required:"true"`
}

var imageExtensions = []string{".jpg", ".png", ".svg", ".jpeg"}

func main() {

}
