// стандартное значение возвращаемое микросервисом системы

package common

type Retvalue struct {
	//флаг успешности операции
	Success bool `json:"Success"`
	//сообщение об ошибке в случае наличия ошибок
	Message string `json:"Message"`
	//возвращаемые данные
	Data interface{} `json:"Data"`
}
