//сервис единой точки доступа к системе
//реализация функционала ограничения количества запросов от клиента в единицу времени

package main

import (
	common "gateway/common"
	"io"
	"net/http"
	"time"
)

// ----------------------------------------------------------
func rateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//накладываем ограничение только на POST-запросы
		if r.Method == "POST" {
			//вычисляем момент открытия временного окна
			start_time := time.Now().Unix() - time_window_sec
			//определяем ip-адрес запроса
			ip, _ := common.GetIP(r)
			//вычисляем количество запросов осуществлённых клиентом с момента начала временного окна
			total_requests := rs.countTotalRequests(ip, start_time)
			if total_requests >= capacity {
				//если количество запросов данного клиента превышает лимит - отказываем ему в обслуживании
				io.WriteString(w, "Too many requests from your ip address")
				return
			} else {
				//регистрируем новый запрос и отправляем его на обработку
				rs.registerRequest(ip)
				next.ServeHTTP(w, r)
			}
		} else {
			//запросы других типов пробрасываем дальше без ограничений
			next.ServeHTTP(w, r)
		}
	})
}
