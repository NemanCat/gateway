//сервис единой точки доступа к системе
//реализация функционала запрета доступа с забаненных ip-адресов

package main

import (
	common "gateway/common"
	"io"
	"net/http"
)

func bannedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//накладываем ограничение только на POST-запросы
		if r.Method == "POST" {
			//ip-адрес запроса
			ip, _ := common.GetIP(r)
			//находится ли ip-адрес в списке забаненных адресов
			if bas.isBanned(ip) {
				//запрещаем доступ
				io.WriteString(w, "BANNED")
			} else {
				//разрешаем доступ
				next.ServeHTTP(w, r)
			}
		} else {
			//запросы других типов пробрасываем дальше без ограничений
			next.ServeHTTP(w, r)
		}
	})
}
