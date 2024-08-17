//сервис единой точки доступа к системе
//прокси запросов к системе

package main

import (
	"encoding/json"
	"fmt"
	common "gateway/common"
	"io"
	"net/http"
	"net/http/httputil"

	"net/url"
	"time"
)

// ------------------------------------------
type CookieData struct {
	Id        string
	Lastname  string
	Firstname string
	Category  string
}

// -------------------------------------------
type DebugTransport struct{}

func (DebugTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	b, err := httputil.DumpRequestOut(r, false)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))
	return http.DefaultTransport.RoundTrip(r)
}

// ---------------------------------------
// авторизация запроса пользователя
// @in r *http.Request обрабатываемый запрос
// @out authorized bool является ли запрос авторизованным
// @out need_login bool требуется перенаправление на страницу авторизации
// @out data *CookieData данные пользователя для авторизованного запроса
// пустая строка для неавторизованных запросов
func authorizeRequest(w *http.ResponseWriter, r *http.Request) (authorized bool, need_login bool, data *CookieData) {
	//проверяем наличие сессионного cookie
	cookie, err := r.Cookie(session_cookie_name)
	if err != nil {
		//неавторизованный запрос
		return false, false, nil
	}
	if cookie == nil {
		//неавторизованный запрос
		return false, false, nil
	}
	//пытаемся расшифровать данные cookie
	value, err := common.DecryptAES(cookie_secret, cookie.Value)
	if err != nil {
		//удаляем cookie
		expire := time.Now().Add(-7 * 24 * time.Hour)
		http.SetCookie(*w, &http.Cookie{
			Name:    session_cookie_name,
			Expires: expire,
		})
		//требуется переход на страницу авторизации
		return false, true, nil
	}
	//пытаемся распарсить cookie
	var d CookieData
	err = json.Unmarshal([]byte(value), &d)
	if err != nil {
		//удаляем cookie
		expire := time.Now().Add(-7 * 24 * time.Hour)
		http.SetCookie(*w, &http.Cookie{
			Name:    session_cookie_name,
			Expires: expire,
		})
		//требуется переход на страницу авторизации
		return false, true, nil
	}
	//запрос авторизован, возвращаем данные cookie
	return true, false, &d
}

// ---------------------------------------
// тип secure proxy
type Proxy struct{}

// безопасное проксирование запросов к ресурсам системы
func (p Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		target                  *url.URL
		id, lastname, firstname string
	)
	//реализация сервиса is-alive
	if r.URL.Path == "/is-alive" {
		io.WriteString(w, "Gateway service is alive at port "+port+"\r\n")
		return
	}

	//определяем ip-адрес запроса
	ip_address, _ := common.GetIP(r)
	//проводим авторизацию запроса
	authorized, need_login, data := authorizeRequest(&w, r)
	if !authorized {
		if need_login {
			//перенаправляем на страницу авторизации
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		//проксим запрос на открытую часть портала
		target, _ = url.Parse(index_workplace)
		id = ""
		lastname = ""
		firstname = ""
	} else {
		//проксим запрос на личный кабинет клиента
		target, _ = url.Parse(customer_workplace)
		id = data.Id
		firstname = data.Firstname
		lastname = data.Lastname
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	d := proxy.Director
	proxy.Director = func(r *http.Request) {
		d(r)                 // call default director
		r.Host = target.Host // set Host header as expected by target
		r.Header.Set("X-IP", ip_address)
		r.Header.Set("X-Userid", id)
		r.Header.Set("X-Lastname", lastname)
		r.Header.Set("X-Firstname", firstname)
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	}
	//--------------------------------------
	//proxy.Transport = DebugTransport{}
	//--------------------------------------
	proxy.ServeHTTP(w, r)
}
