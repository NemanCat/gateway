//сервис единой точки доступа к системе
//осуществляет проверку прав доступа и проксирует запросы к соответствующему рабочему месту

package main

import (
	"context"

	"errors"
	"fmt"
	common "gateway/common"
	pb "gateway/grpc/pb"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net"

	"github.com/joho/godotenv"
	"github.com/oklog/run"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// -------------------------------------------------
const (
	//имя лога приложения
	logname = "services.gateway"
	//частота запуска чистильщика хранилища запросов, минут
	request_store_cleaner_frequency = 3
	//сдвиг минимальной границы удаления старых записей, секунд
	start_time_shift = 5
	//частота запуска чистильщика хранилища забаненных ip-адресов, минут
	banned_store_cleaner_frequency = 5
	//имя сессионного cookie
	session_cookie_name = "secured-portal-session"
	//ключ для шифрования cookie
	cookie_secret = "qJwDNFBpDNGmMMausPBvQ9Kn1oGPkZhB"
)

var (
	//http-порт сервиса
	port string
	//grpc-порт сервиса
	grpc_port string
	//адрес сервиса логирования
	logging_service string
	//url открытой части портала
	index_workplace string
	//url личного кабинета клиента
	customer_workplace string
	//хранилище клиентских запросов
	rs *RequestsStore
	//хранилище попыток аутентификации с несуществующим логином
	ulas *UnknownLoginAttemptsStore
	//хранилище забаненных ip-адресов
	bas *BannedAddressesStore
	//размер временного окна, сек
	time_window_sec int64
	//максимальное количество запросов с одного ip-адреса внутри временного окна
	capacity int
	//максимальное количество запросов на аутентификацию с несуществующим логином
	//по достижении данного количества запросов ip-адрес блокируется
	max_unknown_login_attempts int
	//продолжительность бана ip-адреса, часов
	ban_interval int
	//адрес сервиса рассылки почты
	mailer_service string
	//адрес отправителя писем
	from string
	//имя отправителя писем
	from_name string
	//адрес администратора системы
	admin_email string
)

// --------------------------------------------------
// вывод сообщения в лог
func WriteToLog(message string) error {
	fmt.Println(message)
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	conn, err := grpc.DialContext(ctx, logging_service, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewLoggingServiceClient(conn)
	request := &pb.LoggingMessage{
		Logname: logname,
		Message: message,
	}
	_, err = client.WriteToLog(ctx, request)
	return err
}

func main() {
	//локальные переменные
	var exists bool
	//----------------------------------------
	//получаем настройки приложения
	godotenv.Load()
	//http-порт сервиса
	port, exists = os.LookupEnv("port")
	if !exists {
		port = "8080"
	}
	//grpc-порт сервиса
	grpc_port, exists = os.LookupEnv("grpc_port")
	if !exists {
		grpc_port = "8080"
	}
	//размер временного окна, сек
	time_window_sec_str, exists := os.LookupEnv("time_window_sec")
	if !exists {
		time_window_sec_str = "1"
	}
	time_window_sec = int64(common.StrToInt(time_window_sec_str))
	//максимальное количество запросов с одного ip-адреса внутри временного окна
	capacity_str, exists := os.LookupEnv("capacity")
	if !exists {
		capacity_str = "1"
	}
	capacity = common.StrToInt(capacity_str)
	//максимальное количество запросов на аутентификацию с несуществующим логином
	//по достижении данного количества запросов ip-адрес блокируется
	max_unknown_login_attempts_str, exists := os.LookupEnv("max_unknown_login_attempts")
	if !exists {
		max_unknown_login_attempts_str = "1"
	}
	max_unknown_login_attempts = common.StrToInt(max_unknown_login_attempts_str)
	//продолжительность бана ip-адреса, часов
	ban_interval_str, exists := os.LookupEnv("ban_interval")
	if !exists {
		ban_interval_str = "24"
	}
	ban_interval = common.StrToInt(ban_interval_str)
	//адрес сервиса логирования
	logging_service, exists = os.LookupEnv("logging_service")
	if !exists {
		logging_service = "localhost:5000"
	}
	//url открытой части портала
	index_workplace, exists = os.LookupEnv("frontend_workplace")
	if !exists {
		index_workplace = "http://localhost:8081"
	}
	//url личного кабинета клиента
	customer_workplace, exists = os.LookupEnv("backend_workplace")
	if !exists {
		customer_workplace = "http://localhost:8082"
	}
	//адрес сервиса рассылки почты
	mailer_service, exists = os.LookupEnv("mailer_service")
	if !exists {
		mailer_service = "localhost:5001"
	}
	//адрес почтового робота
	from, exists = os.LookupEnv("from")
	if !exists {
		from = "security@orgdem.ru"
	}
	//имя почтового робота
	from_name, exists = os.LookupEnv("from_name")
	if !exists {
		from = "Система безопасности портала"
	}
	//адрес администратора системы
	admin_email, exists = os.LookupEnv("admin_email")
	if !exists {
		admin_email = "morozov@memosoft.ru"
	}
	//----------------------------------------------
	defer WriteToLog("Gateway service stopped")
	//создаём хранилище клиентских запросов
	rs = createRequestsStore()
	//создаём хранилище попыток аутентификации с несуществующим логином
	ulas = createUnknownLoginAttemptsStore()
	//создаём хранилище забаненных ip-адресов
	bas = createBannedAddressesStore()
	//создаём защищённый прокси - сервер запросов к личному кабинету пользователя
	proxy := Proxy{}
	//веб - сервер приложения
	h := &http.Server{Addr: ":" + port, Handler: bannedMiddleware(rateMiddleware(&proxy))}
	//создаём grpc-сервер сервиса
	gs := GatewayService{}
	server := grpc.NewServer()
	defer server.GracefulStop()
	pb.RegisterGatewayServiceServer(server, &gs)
	reflection.Register(server)
	listener, err := net.Listen("tcp", ":"+grpc_port)
	if err != nil {
		fmt.Println("Could not start Gateway service grpc server with error: ", err)
		fmt.Println("Gateway service stopped")
		return
	}
	//создаём канал для сигнала остановки приложения
	quit_channel := make(chan os.Signal)
	cancel := make(chan struct{})
	signal.Notify(quit_channel, syscall.SIGINT, syscall.SIGTERM)
	//создаём run group
	var g run.Group
	//добавляем в run group http - сервер приложения
	g.Add(func() error {
		return h.ListenAndServe()
	}, func(err error) {
		if err.Error() != "termination signal received" {
			WriteToLog("Could not start Gateway http server with error: " + err.Error())
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		h.Shutdown(ctx)
	})
	//добавляем в run group grpc-сервер приложения
	g.Add(func() error {
		return server.Serve(listener)
	}, func(err error) {
		if err.Error() != "termination signal received" {
			fmt.Println("Could not start Gateway service grpc server with error: ", err)
		}
		server.GracefulStop()
		listener.Close()
	})
	//добавляем в run group чистильщик хранилища запросов
	g.Add(func() error {
		ticker := time.NewTicker(time.Minute * time.Duration(request_store_cleaner_frequency))
		defer ticker.Stop()
		for {
			select {
			case _ = <-cancel:
				//прекращение работы
				return nil
			case <-ticker.C:
				//чистим хранилище запросов от старых записей
				go rs.removeOldRecords(time.Now().Unix() - start_time_shift)
			default:
			}
		}
	}, func(err error) {
	})
	//добавляем в run group чистильщик хранилища забаненных ip-адресов
	g.Add(func() error {
		ticker := time.NewTicker(time.Minute * time.Duration(banned_store_cleaner_frequency))
		defer ticker.Stop()
		for {
			select {
			case _ = <-cancel:
				//прекращение работы
				return nil
			case <-ticker.C:
				//чистим хранилище забаненных ip-адресов от старых записей
				go bas.removeOldRecords()
			default:
			}
		}
	}, func(err error) {
	})
	//добавляем в run group signal handler
	g.Add(func() error {
		WriteToLog("Gateway service started at http port " + port + " and grpc port " + grpc_port)
		<-quit_channel
		return errors.New("termination signal received")
	}, func(err error) {
		close(cancel)
	})
	g.Run()
}
