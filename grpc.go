//сервис единой точки доступа к системе
//реализация grpc-сервера сервиса

package main

import (
	"bytes"
	"context"

	pb "gateway/grpc/pb"
	"html/template"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

// -------------------------------------------------
// отправка письма через сервис рассылки
func SendEmailMessage(to string, from string, from_name string, subject string, body string) error {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	conn, err := grpc.DialContext(ctx, mailer_service, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewMailerServiceClient(conn)
	req := &pb.MailerMessage{
		To:       to,
		From:     from,
		Fromname: from_name,
		Subject:  subject,
		Body:     body,
	}
	_, err = client.SendMail(ctx, req)
	return err
}

// -------------------------------------------------
// grpc-сервис приложения
type GatewayService struct {
	pb.UnimplementedGatewayServiceServer
}

// -------------------------------------------------
// реализация методов grpc-сервера
// приём сообщения об инциденте безопасности - попытка аутентификации с несуществующим логином
func (gs *GatewayService) UnknownLoginAttempt(ctx context.Context,
	req *pb.StringSecurityMessage) (*empty.Empty, error) {
	//ip-адрес запроса
	ip_address := req.GetValue()
	//проверяем текущее состояние счётчика для данного ip-адреса в хранилище инцидентов
	counter := ulas.getCounter(ip_address)
	if counter < max_unknown_login_attempts {
		//счётчик инцидентов ещё не достиг максимального значения
		//увеличиваем счётчик на 1
		ulas.incCounter(ip_address)
	} else {
		//количество инцидентов превысило  максимальное допустимое значение
		//добавляем ip-адрес запроса в список забаненных адресов
		bas.addToBan(ip_address)
		//отправляем администраторам системы письмо об инциденте
		var doc bytes.Buffer
		tmpl := template.New("unknown_login_ban_template")
		t, err := tmpl.Parse(unknown_login_ban_template)
		if err == nil {
			data := struct {
				Ip_address string
			}{
				ip_address,
			}
			err = t.Execute(&doc, data)
			if err == nil {
				SendEmailMessage(admin_email, from, from_name,
					"Инцидент безопасноcти на портале Оргдиагностика-ЭМ", doc.String())
			}
		}

		//удаляем ip-адрес запроса из списка инцидентов данного типа
		ulas.Delete(ip_address)
	}
	return &empty.Empty{}, nil
}
