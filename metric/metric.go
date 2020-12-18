package metric

import (
	"context"
	bblog "github.com/buildboxapp/lib/log"
	bbstate "github.com/buildboxapp/lib/state"
	"net/http"
	"sync"
	"time"
)

type serviceMetric struct {
	Connections int `json:"connection"`				// количество открытых соединений за весь период учета
	connectionOpen int `json:"connection_current"`	// текущее кол-во открытых соединений (+ при запрос - при ответе)
	Queue []int `json:"queue"`						// массив соединений в очереди (не закрытых) см.выше
	StateHost bbstate.StateHost
	mux *sync.Mutex
	ctx context.Context
}

type ServiceMetric interface {
	SetState()
	SetConnectionIncrement()
	SetConnectionDecrement()
	Get() serviceMetric
	Middleware(next http.Handler) http.Handler
}

func (s *serviceMetric) SetState(){

	s.mux.Lock()
	s.StateHost.Tick()
	s.mux.Unlock()

	return
}

// формируем временной ряд количества соединений
// при начале запроса увеличиваем, при завершении уменьшаем
func (s *serviceMetric) SetConnectionIncrement(){
	go func() {
		s.mux.Lock()
		s.connectionOpen = s.connectionOpen + 1
		s.Queue = append(s.Queue,  s.connectionOpen)
		s.mux.Unlock()
	}()

	return
}

func (s *serviceMetric) SetConnectionDecrement(){
	go func() {
		s.mux.Lock()
		if s.connectionOpen != 0 {
			s.connectionOpen = s.connectionOpen - 1
		}
		s.Queue = append(s.Queue,  s.connectionOpen)
		s.mux.Unlock()
	}()

	return
}

func (s *serviceMetric) Get() (result serviceMetric) {
	result.StateHost = s.StateHost
	result.Connections = s.Connections
	result.Queue = s.Queue

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
		Connections: 0,
		Queue: []int{},
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
				// расчитываем значения метрик

				// сохраняем расчитанные значения в памяти для пинга и затираем текущие данные

				// сохраняем значение метрик в лог
				logger.Trace(m.Get())

				ticker = time.NewTicker(interval)
			}
		}
}