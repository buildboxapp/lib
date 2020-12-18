package metric

import (
	"context"
	bblog "github.com/buildboxapp/lib/log"
	bbstate "github.com/buildboxapp/lib/state"
	"net/http"
	"sync"
	"time"
)

type Metrics struct {
	StateHost bbstate.StateHost
	Connections int `json:"connection"`				// количество открытых соединений за весь период учета
	AVG_Queue int `json:"avg_queue"` 				// среднее количество запросов в очереди
	QTL_Queue_90 int `json:"qtl_queue_90"` 			// квантиль 90%
	QTL_Queue_99 int `json:"qtl_queue_99"` 			// квантиль 99%
	AVG_TPR time.Duration `json:"avg_tpr"`				// Time per request - среднее время обработки запроса
	RPS	int `json:"rps"`							// Request per second - количество запросов в секунду
}

type serviceMetric struct {
	Metrics
	Stash Metrics `json:"stash"`					// карман для сохранения предыдущего значения

	connectionOpen int `json:"connection_current"`	// текущее кол-во открытых соединений (+ при запрос - при ответе)
	queue []int `json:"queue"`						// массив соединений в очереди (не закрытых) см.выше
	mux *sync.Mutex
	ctx context.Context
}

type ServiceMetric interface {
	SetState()
	SetConnectionIncrement()
	SetConnectionDecrement()
	Generate() (result Metrics)
	Get() (result Metrics)
	Clear()
	SaveToStash()
	Middleware(next http.Handler) http.Handler
}

func (s *serviceMetric) SetState(){
	s.mux.Lock()
	s.StateHost.Tick()
	s.mux.Unlock()

	return
}

// увеличиваем счетчик и добавляем в массив метрик
// формируем временной ряд количества соединений
// при начале запроса увеличиваем, при завершении уменьшаем
// запускаем в отдельной рутине, потому что ф-ция вызывается из сервиса и не должна быть блокирующей
func (s *serviceMetric) SetConnectionIncrement(){
	go func() {
		s.mux.Lock()
		s.connectionOpen = s.connectionOpen + 1
		s.queue = append(s.queue,  s.connectionOpen)
		s.mux.Unlock()
	}()

	return
}

// уменьшаем счетчик и добавляем в массив метрик
// запускаем в отдельной рутине, потому что ф-ция вызывается из сервиса и не должна быть блокирующей
func (s *serviceMetric) SetConnectionDecrement(){
	go func() {
		s.mux.Lock()
		if s.connectionOpen != 0 {
			s.connectionOpen = s.connectionOpen - 1
		}
		s.queue = append(s.queue,  s.connectionOpen)
		s.mux.Unlock()
	}()

	return
}

// сохраняем текущее значение расчитанных метрик в кармане
func (s *serviceMetric) SaveToStash() {
	s.mux.Lock()
	s.Stash.RPS = s.RPS
	s.Stash.QTL_Queue_99 = s.QTL_Queue_99
	s.Stash.QTL_Queue_90 = s.QTL_Queue_90
	s.Stash.AVG_TPR = s.AVG_TPR
	s.Stash.AVG_Queue = s.AVG_Queue
	s.mux.Unlock()
}

func (s *serviceMetric) Clear() {
	s.mux.Lock()
	s.Connections = 0
	s.connectionOpen = 0
	s.queue = []int{}
	s.mux.Unlock()

	return
}

func (s *serviceMetric) Get() (result Metrics) {
	return s.Stash
}

func (s *serviceMetric) Generate() (result Metrics) {
	s.SetState()	// получаю текущие метрики загрузки хоста
	result.StateHost = s.StateHost
	result.Connections = s.Connections	// текущее кол-во соединений

	// расчитываем значения метрик
	s.AVG_Queue = 1
	s.AVG_TPR = 10
	s.QTL_Queue_90 = 0
	s.QTL_Queue_99 = 0
	s.RPS = 100

	return result
}

func (s *serviceMetric) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// увеличиваем счетчик активных сессий
		s.SetConnectionIncrement()
		next.ServeHTTP(w, r)

		// уменьшаем счетчик активных сессий
		s.SetConnectionDecrement()
	})
}

// interval - интервалы времени, через которые статистика будет сбрасыватсья в лог
func New(ctx context.Context, logger *bblog.Log, interval time.Duration) (metrics ServiceMetric) {
	m := sync.Mutex{}
	metrics = &serviceMetric{
		connectionOpen: 0,
		queue: []int{},
		mux: &m,
		ctx: ctx,
	}

	go RunMetricLogger(ctx, metrics, logger, interval)

	return metrics
}

func RunMetricLogger(ctx context.Context, m ServiceMetric, logger *bblog.Log, interval time.Duration)  {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <- ctx.Done():
				return
			case <- ticker.C:
				// сохраняем расчитанные значения в памяти для пинга и затираем текущие данные

				// сохраняем значение метрик в лог
				m.Generate()			// сгенерировали метрики
				m.SaveToStash()			// сохранили в карман
				m.Clear()				// очистили объект метрик для приема новых данных
				logger.Trace(m.Get())	// записали в лог из кармана

				ticker = time.NewTicker(interval)
			}
		}
}