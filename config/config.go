package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/buildboxapp/lib"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/gommon/color"
	"os"
	"strings"
)

type config struct {}

type Config interface {
	Load(configname string) (err error)
}

var warning = color.Red("[Fail]")

// читаем конфигурации
// получаем только название конфигурации
// 1. поднимаемся до корневой директории
// 2. от нее ищем полный путь до конфига
// 3. читаем по этому пути
func (c *config) Load(configname string, cfg interface{}) (err error) {

	if err := envconfig.Process("", &cfg); err != nil {
		fmt.Printf("%s Error load default enviroment: %s\n", warning, err)
		os.Exit(1)
	}

	// 1.
	rootDir, err := lib.RootDir()

	// 2.
	confidPath, err := c.FullPathConfig(rootDir, configname)
	fmt.Println(confidPath)

	// 3.
	c.Read(confidPath)

	return err
}

// получаем путь от переданной директории
// если defConfig = true - значит ищем конфигурацию по-умолчанию
func (c *config) FullPathConfig(rootDir, configuration string) (configPath string, err error) {
	var nextPath string
	directory, err := os.Open(rootDir)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer directory.Close()

	objects, err := directory.Readdir(-1)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// пробегаем текущую папку и считаем совпадание признаков
	for _, obj := range objects {
		nextPath = rootDir + sep + obj.Name()

		if obj.IsDir() {
			dirName := obj.Name()

			// не входим в скрытые папки
			if dirName[:1] != "." {
				configPath, err = c.FullPathConfig(nextPath, configuration)
				if configPath != "" {
					return configPath, err // поднимает результат наверх
				}
			}

		} else {
			if configuration == "default" { // проверяем на получение конфигурации по-умолчанию
				if strings.Contains(nextPath, ".cfg") {
					//confJson, err := ReadFile(nextPath)
					//err = json.Unmarshal([]byte(confJson), &conf)
					//if err == nil {
					//	d := conf["default"]
					//	if d == "checked" {
					//		return nextPath, err
					//	}
					//}
				}
			} else {
				if !strings.Contains(nextPath, "/.") {
					if strings.Contains(obj.Name(), configuration) {
						return nextPath, err
					}
				}
			}
		}
	}

	return configPath, err
}

// Читаем конфигурация по заданному полному пути
func (c *config) Read(configfile string) (err error) {
	configfileSplit := strings.Split(configfile, ".")
	if len(configfile) == 0 {
		return fmt.Errorf("%s", "Error. Configfile is empty.")
	}
	if len(configfileSplit) == 1 {
		configfile = configfile + ".cfg"
	}

	if _, err = toml.DecodeFile(configfile, &c); err != nil {
		fmt.Printf("%s Error: %s (configfile: %s)\n", warning, err, configfile)
	}

	return err
}