// обертка для логирования, которая дополняем аттрибутами логируемого процесса logrus
// дополняем значениями, идентифицирующими запущенный сервис UID,Name,Service

package log

import (
	"github.com/sirupsen/logrus"
	"context"
	"strings"
	"time"

	"fmt"
	"io"
	"os"
	"math/rand"
	"path/filepath"
)

var logrusB = logrus.New()
var sep = string(filepath.Separator)

type Log struct {

	// куда логируем? stdout/;*os.File на файл, в который будем писать логи
	Output io.Writer `json:"output"`
	//Debug:
	// сообщения отладки, профилирования.
	// В production системе обычно сообщения этого уровня включаются при первоначальном
	// запуске системы или для поиска узких мест (bottleneck-ов).

	//Info: - логировать процесс выполнения
	// обычные сообщения, информирующие о действиях системы.
	// Реагировать на такие сообщения вообще не надо, но они могут помочь, например,
	// при поиске багов, расследовании интересных ситуаций итд.

	//Warning: - логировать странные операции
	// записывая такое сообщение, система пытается привлечь внимание обслуживающего персонала.
	// Произошло что-то странное. Возможно, это новый тип ситуации, ещё не известный системе.
	// Следует разобраться в том, что произошло, что это означает, и отнести ситуацию либо к
	// инфо-сообщению, либо к ошибке. Соответственно, придётся доработать код обработки таких ситуаций.

	//Error: - логировать ошибки
	// ошибка в работе системы, требующая вмешательства. Что-то не сохранилось, что-то отвалилось.
	// Необходимо принимать меры довольно быстро! Ошибки этого уровня и выше требуют немедленной записи в лог,
	// чтобы ускорить реакцию на них. Нужно понимать, что ошибка пользователя – это не ошибка системы.
	// Если пользователь ввёл в поле -1, где это не предполагалось – не надо писать об этом в лог ошибок.

	//Panic: - логировать критические ошибки
	// это особый класс ошибок. Такие ошибки приводят к неработоспособности системы в целом, или
	// неработоспособности одной из подсистем. Чаще всего случаются фатальные ошибки из-за неверной конфигурации
	// или отказов оборудования. Требуют срочной, немедленной реакции. Возможно, следует предусмотреть уведомление о таких ошибках по SMS.
	// указываем уровни логирования Error/Warning/Debug/Info/Panic

	//Trace: - логировать обработки запросов

	// можно указывать через | разные уровени логирования, например Error|Warning
	// можно указать All - логирование всех уровней
	Levels string `json:"levels"`
	// uid процесса (сервиса), который логируется
	UID string `json:"uid"`
	// имя процесса (сервиса), который логируется
	Name string `json:"name"`
	// название сервиса (app/gui...)
	Service string `json:"service"`
	Dir string `json:"dir"`
}

func (c *Log) Trace(args ...interface{}) {
	if strings.Contains(c.Levels, "Trace") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})

		logrusB.WithFields(logrus.Fields{
			"name": c.Name,
			"uid":  c.UID,
			"srv":  c.Service,
		}).Trace(args...)
	}
}

func (c *Log) Debug(args ...interface{}) {
	if strings.Contains(c.Levels, "Debug") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})

		// Only log the warning severity or above.
		//logrusB.SetLevel(logrus.InfoLevel)

		logrusB.WithFields(logrus.Fields{
			"name": c.Name,
			"uid":  c.UID,
			"srv":  c.Service,
		}).Debug(args...)
	}
}

func (c *Log) Info(args ...interface{}) {
	if strings.Contains(c.Levels, "Info") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})

		logrusB.WithFields(logrus.Fields{
			"name": c.Name,
			"uid":  c.UID,
			"srv":  c.Service,
		}).Info(args...)
	}
}

func (c *Log) Warning(args ...interface{}) {
	if strings.Contains(c.Levels, "Warning") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})

		logrusB.WithFields(logrus.Fields{
			"name": c.Name,
			"uid":  c.UID,
			"srv":  c.Service,
		}).Warn(args...)
	}
}

func (c *Log) Error(err error, args ...interface{}) {
	if strings.Contains(c.Levels, "Error") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})

		logrusB.WithFields(logrus.Fields{
			"name":  c.Name,
			"uid":   c.UID,
			"srv":   c.Service,
			"error": fmt.Sprint(err),
		}).Error(args...)
	}
}

func (c *Log) Panic(err error, args ...interface{}) {
	if strings.Contains(c.Levels, "Fatal") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})

		logrusB.WithFields(logrus.Fields{
			"name":  c.Name,
			"uid":   c.UID,
			"srv":   c.Service,
			"error": fmt.Sprint(err),
		}).Panic(args...)
	}
}

// внутренняя ф-ция логирования и прекращения работы программы
func (c *Log) Exit(err error, args ...interface{}) {
	if strings.Contains(c.Levels, "Fatal") {
		logrusB.SetOutput(c.Output)
		logrusB.SetFormatter(&logrus.JSONFormatter{})

		logrusB.WithFields(logrus.Fields{
			"name":  c.Name,
			"uid":   c.UID,
			"srv":   c.Service,
			"error": fmt.Sprint(err),
		}).Fatal(args...)
	}
}

// Переинициализация файла логирования
func (c *Log) RotateInit(ctx context.Context)  {
	var delayReload time.Duration = 10		// минут - перезагрузка лога
	var delayClear time.Duration = 30		// минут - удаляем старые файлы

	// попытка обновить файл (раз в 10 минут)
	go func() {
		ticker := time.NewTicker(delayReload * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <- ctx.Done():
				return
			case <- ticker.C:
				b := New(c.Dir, c.Levels, fmt.Sprint(rand.Int()), c.Name, c.Service)
				c.Output = b.Output
				ticker = time.NewTicker(delayReload * time.Second)
			}
		}
	}()

	// попытка очистки старый файлов (каждый час)
	go func() {
		ticker := time.NewTicker(delayClear * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <- ctx.Done():
				return
			case <- ticker.C:
				oneMonthAgo := time.Now().AddDate(0, -1, 0) // minus 1 месяц
				fileMonthAgoDate := oneMonthAgo.Format("2006.01.02")

				// пробегаем директорию и читаем все файлы, если имя меньше текущее время - месяц = удаляем
				directory, _ := os.Open(c.Dir)
				objects, err := directory.Readdir(-1)
				if err != nil {
					c.Error(err, "Error read directory: ", directory)
					return
				}

				for _, obj := range objects {
					filename := obj.Name()
					filenameMonthAgoDate := c.Service + "_" + fileMonthAgoDate

					if filenameMonthAgoDate > filename {
						pathFile := c.Dir + sep + filename
						err = os.Remove(pathFile)
						if err != nil {
							c.Error(err, "Error deleted file: ", pathFile)
							return
						}
					}
				}
				ticker = time.NewTicker(delayClear * time.Second)
			}
		}
	}()



}

func New(logsDir, level, uid, name, srv string) *Log {
	var output io.Writer
	var err error
	var mode os.FileMode

	datefile := time.Now().Format("2006.01.02")
	logName := srv + "_" + datefile + ".log"

	// создаем/открываем файл логирования и назначаем его логеру
	mode = 0711
	err = os.MkdirAll(logsDir, mode)
	if err != nil {
		logrus.Error(err, "Error creating directory")
		return nil
	}

	output, err = os.OpenFile(logsDir+"/"+logName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logrus.Panic(err, "error opening file")
		return nil
	}

	return &Log{
		Output:       output,
		Levels: 	  level,
		UID:          uid,
		Name:         name,
		Service:      srv,
		Dir: 		  logsDir,
	}
}