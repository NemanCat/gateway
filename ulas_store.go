// сервис единой точки доступа к системе
// реализация функционала хранилища инцидентов безопасности - попыток аутентификации с несуществующим логином
package main

import (
	"sync"
)

type UnknownLoginAttemptsStore struct {
	//хэшированный список запросов
	//ключ списка - ip-адрес клиента
	//элемент списка - счётчик количества инцидентов данного типа для ip-адреса
	list map[string]int
	//мьютекс для многопоточного доступа
	mutex *sync.Mutex
}

// --------------------------------------------------------
// создание экземпляра хранилища инцидентов
func createUnknownLoginAttemptsStore() *UnknownLoginAttemptsStore {
	ulas := new(UnknownLoginAttemptsStore)
	ulas.list = make(map[string]int)
	ulas.mutex = &sync.Mutex{}
	return ulas
}

// получение значения счётчика для указанного ip-адреса
// @in ip ip-адрес запроса
// @out int текущее значение счётчика или 0 если запист для указанного ip-адреса отсутствует
func (ulas *UnknownLoginAttemptsStore) getCounter(ip string) int {
	ulas.mutex.Lock()
	defer ulas.mutex.Unlock()
	//проверяем наличие в хранилище записи для указанного ip-адреса
	value, ok := ulas.list[ip]
	if ok {
		//возвращаем текущее значение счётчика
		return value
	} else {
		//возвращаем 0
		return 0
	}
}

// увеличение значения счётчика инцидентов для указанного ip-адреса
// @in ip ip-адрес запроса
func (ulas *UnknownLoginAttemptsStore) incCounter(ip string) {
	ulas.mutex.Lock()
	defer ulas.mutex.Unlock()
	//проверяем наличие в хранилище записи для указанного ip-адреса
	_, ok := ulas.list[ip]
	if ok {
		//запись для указанного адреса существует в хранилище
		//увеличиваем счётчик на единицу
		ulas.list[ip]++
	} else {
		//запись для указанного адреса не существует в хранилище
		//добавляем запись и выставляем значение счётчика в 1
		ulas.list[ip] = 1
	}
}

// удаление записи из списка инцидентов
// @in ip ip-адрес запроса
func (ulas *UnknownLoginAttemptsStore) Delete(ip string) {
	ulas.mutex.Lock()
	defer ulas.mutex.Unlock()
	delete(ulas.list, ip)
}
