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
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Function interface {
	ResponseJSON(w http.ResponseWriter, objResponse interface{}, status string, error error, metrics interface{}) (err error)
	RunProcess(path, config, command, mode string) (pid int, err error)
	ReadConf(configfile string) (conf map[string]string, confjson string, err error)
	RootDir() (rootDir string, err error)
	RootDirAction(currentDir string) (rootDir string, err error)
	Hash(str string) (result string, err error)
	PanicOnErr(err error)
	UUID() (result string)
}

type libFunction struct {
	Function
}

// если status не из списка, то вставляем статус - 501 и Descraption из статуса
func (f *libFunction) ResponseJSON(w http.ResponseWriter, objResponse interface{}, status string, error error, metrics interface{}) (err error) {

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

	return
}

// стартуем сервис из конфига
func (f *libFunction) RunProcess(path, config, command, mode string) (pid int, err error) {
	var cmd *exec.Cmd
	var l libFiles

	if config == "" {
		return 0, fmt.Errorf("%s", "Configuration file is not found")
	}
	if command == "" {
		command = "start"
	}

	cmd = exec.Command(path, command, "--config", config, "--mode", mode)
	if mode == "debug" {
		t := time.Now().Format("2006.01.02-15-04-05")
		s := strings.Split(path, sep)
		srv := s[len(s)-1]

		err = l.CreateDir("debug" + sep + srv, 0777)
		config_name := strings.Replace(config, "-", "", -1)

		f, _ := os.Create(  "debug" + sep + srv + sep + config_name + "_" + fmt.Sprint(t) + ".log")

		cmd.Stdout = f
		cmd.Stderr = f
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err != nil {
		return 0, err
	}
	pid = cmd.Process.Pid

	return
}

// читаем файл конфигурации и возвращаем
// объект конфига, джейсон-конфига и ошибку
// ЗАГЛУШКА ДЛЯ PS и LS
func (f *libFunction) ReadConf(configfile string) (conf map[string]string, confjson string, err error) {
	//
	//	if configfile == "" {
	//		return nil, "", err
	//	}
	//
	//	// дополняем название файла раcширением
	//	if !strings.Contains(configfile, ".json") {
	//		configfile += ".json"
	//	}
	//
	//	rootDir, err := RootDir()
	//	if err != nil {
	//		return
	//	}
	//	startDir := rootDir + string(filepath.Separator) + "upload"
	//	fileName, err := ReadConfAction(startDir, configfile, false)
	//	if err != nil {
	//		return nil, "", err
	//	}
	//
	//	confJson, err := ReadFile(fileName)
	//	if err != nil {
	//		return nil, "", err
	//	}
	//
	//	err = json.Unmarshal([]byte(confJson), &conf)
	//	if err != nil {
	//		return nil, "", err
	//	}
	//
	return conf, confjson, err
}

// корневую директорию (проверяем признаки в текущей директории + шагом вверх)
// входные: currentDir - текущая папка, level - глубина (насколько уровеней вверх проверяем)
// вниз не проверяем, потому что вряд ли кто будет запускать выше корневой папки
// но если надо, то можно и доделать
func (f *libFunction) RootDir() (rootDir string, err error) {
	file, err := filepath.Abs(os.Args[0])
	if err != nil {
		return
	}

	cdir := path.Dir(file)
	rootDir, err = f.RootDirAction(cdir)
	if err != nil {
		fmt.Println("Error calculation RootDir. File: ", file, "; Error: ", err)
	}

	return
}

// получаем путь от переданной директории
func (f *libFunction) RootDirAction(currentDir string) (rootDir string, err error) {

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
		rootDir, err = f.RootDirAction(currentDir)
	} else {
		rootDir = currentDir
	}

	return rootDir, err
}

func (f *libFunction) Hash(str string) (result string, err error) {
	h := sha1.New()
	h.Write([]byte(str))
	result = hex.EncodeToString(h.Sum(nil))

	return
}

func (f *libFunction) PanicOnErr(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		panic(err)
	}
}

func (f *libFunction) UUID() (result string) {
	stUUID := uuid.NewV4()
	return stUUID.String()
}
