package lib

import (
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"os/exec"
	"path"
	"syscall"

	"crypto/sha1"
	"encoding/hex"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//func (c *Lib) Init(output io.Writer, urlgui, urlapi string) {
//	c.Logger.Output = output
//	c.UrlGUI = urlgui
//	c.UrlAPI = urlapi
//}

// если status не из списка, то вставляем статус - 501 и Descraption из статуса
func (c *Lib) ResponseJSON(w http.ResponseWriter, objResponse interface{}, status string, error error, metrics interface{}) {

	if w == nil {
		return
	}

	errMessage := RestStatus{}
	st, found := StatusCode[status]
	if found {
		errMessage = st
	} else {
		errMessage = StatusCode["NotStatus"]
	}

	objResp := &Response{}
	if error != nil {
		errMessage.Error = fmt.Sprint(error)
	}

	// Metrics
	b1, _ := json.Marshal(metrics)
	var metricsR Metrics
	json.Unmarshal(b1, &metricsR)
	if metrics != nil {
		objResp.Metrics = metricsR
	}

	objResp.Status = errMessage
	objResp.Data = objResponse

	// формируем ответ
	out, err := json.Marshal(objResp)
	if err != nil {
		log.Printf("%s", err)
	}

	//WriteFile("./dump.json", out)

	w.WriteHeader(errMessage.Status)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(out)
}

// стартуем сервис из конфига
func (c *Lib) RunProcess(path, config, command string) (pid int, err error) {

	if config == "" {
		return 0, fmt.Errorf("%s", "Configuration file is not found")
	}
	if command == "" {
		command = "start"
	}

	cmd := exec.Command(path, command, "--config", config)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	//stdout, err := cmd.StdoutPipe()
	//if err != nil {
	//	log.Fatal(err)
	//}
	err = cmd.Start()
	if err != nil {
		return 0, err
	}

	return cmd.Process.Pid, err
}

// останавливаем сервис по порту
//func (c *Lib) StopProcess(workdir, fileConfig, message string) {
//
//	if fileConfig == "" {
//		fmt.Println(color.Red("ERROR!") + " Configuration file is not found.\n")
//		return
//	}
//
//	var err error
//	done := color.Yellow("OK")
//	fail := color.Red("FAIL")
//	fileStart := workdir + "/buildbox"
//
//	cmd := exec.Command(fileStart, "stop", "--config", fileConfig)
//	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
//	err = cmd.Start()
//	if err != nil {
//		fmt.Printf("%s Exist %s: %s\n", message, fail, err)
//		return
//	}
//
//	fmt.Printf("%s Exist %s: %s\n", done, message, cmd.Process.Pid)
//	return
//}

//////////////////////////////////////////////////////////////////
////////////////////////// СЕРВИСНЫЕ ФУНКЦИИ /////////////////////
//////////////////////////////////////////////////////////////////

// читаем файл конфигурации и возвращаем
// объект конфига, джейсон-конфига и ошибку
func (c *Lib) ReadConf(configfile string) (conf map[string]string, confjson string, err error) {

	if configfile == "" {
		return nil, "", err
	}

	// дополняем название файла разширением
	if !strings.Contains(configfile, ".json") {
		configfile += ".json"
	}

	startDir := c.RootDir() + string(filepath.Separator) + "upload"
	fileName, err := c.ReadConfAction(startDir, configfile, false)
	if err != nil {
		return nil, "", err
	}

	confJson, err := c.ReadFile(fileName)
	if err != nil {
		return nil, "", err
	}

	err = json.Unmarshal([]byte(confJson), &conf)
	if err != nil {
		return nil, "", err
	}

	return conf, confJson, err
}

// получаем путь от переданной директории
// если defConfig = true - значит ищем конфигурацию по-умолчанию
func (c *Lib) ReadConfAction(currentDir, configuration string, defConfig bool) (configPath string, err error) {
	var conf map[string]string
	directory, _ := os.Open(currentDir)
	objects, err := directory.Readdir(-1)
	if err != nil {
		return "", err
	}

	// пробегаем текущую папку и считаем совпадание признаков
	for _, obj := range objects {
		nextPath := currentDir + string(filepath.Separator) + obj.Name()
		if obj.IsDir() {
			configPath, err = c.ReadConfAction(nextPath+string(filepath.Separator)+"ini", configuration, defConfig)
			if configPath != "" {
				return configPath, err // поднимает результат наверх
			}

		} else {
			if defConfig { // проверяем на получение конфигурации по-умолчанию
				confJson, err := c.ReadFile(nextPath)
				err = json.Unmarshal([]byte(confJson), &conf)
				if err == nil {
					d := conf["default"]
					if d == "checked" {
						return obj.Name(), err
					}
				}
			} else {
				if obj.Name() == configuration {
					return nextPath, err
				}
			}
		}
	}

	return configPath, err
}

// получаем конфигурацию по-умолчанию для сервера (перебираем конфиги и ищем первый у которого default=on)
func (c *Lib) DefaultConfig() (fileConfig string, err error) {
	startDir := c.RootDir() + string(filepath.Separator) + "upload"

	return c.ReadConfAction(startDir, "", true)
}

// получаем конфигурацию по-умолчанию для сервера (перебираем конфиги и ищем первый у которого default=on)
func (c *Lib) DefaultConfig111() (fileConfig string, err error) {
	fpath := ""

	if !strings.Contains(fileConfig, "/") {
		fpath = c.RootDir() + string(filepath.Separator) + "ini"
	}

	c.Logger.Info("Search DefaultConfig from : ", fpath)

	files, err := ioutil.ReadDir(fpath)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		conf, _, err := c.ReadConf(file.Name())
		if err == nil {
			d := conf["default"]
			if d != "" {
				fileConfig = file.Name()
				continue
			}

		}
	}

	c.Logger.Info("Search DefaultConfig result : ", fileConfig)

	return fileConfig, err
}

// определяем текущий каталог для первого запуска, чтобы прочитать файл с конфигурацией
func (c *Lib) CurrentDir() string {
	// путь к шаблонам при запуске через командную строку
	var runDir, _ = os.Getwd()
	var currentDir = filepath.Dir(os.Args[0]) // если запускать с goland отдает темповую папку (заменяем)
	if currentDir != runDir {
		currentDir = runDir
	}
	return currentDir
}

// корневую директорию (проверяем признаки в текущей директории + шагом вверх)
// входные: currentDir - текущая папка, level - глубина (насколько уровеней вверх проверяем)
// вниз не проверяем, потому что вряд ли кто будет запускать выше корневой папки
// но если надо, то можно и доделать
func (c *Lib) RootDir() (rootDir string) {
	var err error
	file, _ := filepath.Abs(os.Args[0])
	cdir := path.Dir(file)

	rootDir, err = RootDirAction(cdir)
	if err != nil {
		fmt.Println("Error calculation RootDir. File: ", file, "; Error: ", err)
		c.Logger.Panic(err, "Error calculation RootDir. File: ", file)
	}

	return rootDir
}

// получаем путь от переданной директории
func RootDirAction(currentDir string) (rootDir string, err error) {

	// признаки рутовой директории - наличие файла buildbox (стартового (не меняется)
	// наличие директорий certs + dbs
	directory, _ := os.Open(currentDir)
	objects, err := directory.Readdir(-1)
	if err != nil {
		return "", err
	}

	countTrueStatus := 0
	// пробегаем текущую папку и считаем совпадание признаков
	// если их 3 - значит это корень
	for _, obj := range objects {
		if obj.IsDir() {
			if obj.Name() == "certs" {
				countTrueStatus = countTrueStatus + 1
			}
		} else {
			if obj.Name() == "buildbox" {
				countTrueStatus = countTrueStatus + 1
			}
		}
	}

	if countTrueStatus < 2 {
		sc := strings.Split(currentDir, string(filepath.Separator))
		scc := sc[:len(sc)-1]
		currentDir = strings.Join(scc, string(filepath.Separator))
		rootDir, err = RootDirAction(currentDir)
	} else {
		rootDir = currentDir
	}

	return rootDir, err
}

func (c *Lib) Hash(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	sha1_hash := hex.EncodeToString(h.Sum(nil))

	return sha1_hash
}

func (c *Lib) PanicOnErr(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		panic(err)
	}
}

func (c *Lib) UUID() string {
	stUUID := uuid.NewV4()
	return stUUID.String()
}
