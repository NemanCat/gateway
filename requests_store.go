//сервис единой точки доступа к системе
//реализация функционала доступа к хранилищу клиентских запросов

package main

import (
	"sync"
	"time"
)

// хранилище запросов
type RequestsStore struct {
	//хэшированный список запросов
	//ключ списка - ip-адрес клиента
	//элемент списка - хэшированный список,
	//ключ - секунда в формате epoch, элемент - количество запросов данного клиента в эту секунду
	list map[string]map[int64]int
	//мьютекс для многопоточного доступа
	mutex *sync.Mutex
}

// --------------------------------------------------------
// создание экземпляра хранилища запросов
func createRequestsStore() *RequestsStore {
	rs := new(RequestsStore)
	rs.list = make(map[string]map[int64]int)
	rs.mutex = &sync.Mutex{}
	return rs
}

// поиск журнала запросов для указанного клиента
// ip ip-адрес клиента
func (rs *RequestsStore) findClient(ip string) *map[int64]int {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	value, ok := rs.list[ip]
	if !ok {
		return nil
	}
	return &value
}

// регистрация запроса
// ip ip-адрес клиента
func (rs *RequestsStore) registerRequest(ip string) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	//проверяем существует ли в хранилище запись для данного клиента
	client, ok := rs.list[ip]
	if !ok {
		//записи для данного клиента нет, создаём её
		rs.list[ip] = make(map[int64]int)
		client = rs.list[ip]
	}
	//текущая секунда в формате epoch
	current_sec := time.Now().Unix()
	//проверяем есть ли запись для текущей секунды
	_, ok = rs.list[ip][current_sec]
	if !ok {
		//добавляем текущую секунду в список запросов и устанавливаем список запросов в 1
		client[current_sec] = 1
	} else {
		//увеличиваем счётчик запросов для текущей секунды
		client[current_sec]++
	}
}

// подсчёт количества запросов указанного клиента начиная с указанного времени
// ip  ip-адрес клиента
// start_time начало интервала времени за который подсчитывается количество запросов
func (rs *RequestsStore) countTotalRequests(ip string, start_time int64) int {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	//проверяем существует ли в хранилище запись для данного клиента
	client, ok := rs.list[ip]
	if !ok {
		return 0
	}
	//подсчитываем количество запросов данного клиента за время прошедшее после указанного времени
	//при этом удаляем записи для запросов поступивших ранее указанного времени минус N секунд
	total_requests := 0
	for key, _ := range client {
		if key > start_time {
			total_requests += client[key]
		} else {
			if key <= start_time-start_time_shift {
				delete(client, key)
			}
		}
	}
	return total_requests
}

// удаление из хранилища старых записей
// будут удалены все записи для более старых запросов
// после чего будут удалены все записи для клиентов у которых количество запросов равно 0
// minimal_time нижняя граница интервала времени, будут удалены записи запросов за более раннее время
func (rs *RequestsStore) removeOldRecords(minimal_time int64) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	//удаляем записи для старых запросов
	for _, client := range rs.list {
		for key, _ := range client {
			if key <= minimal_time {
				delete(client, key)
			}
		}
	}
	//удаляем записи для клиентов у которых нет ни одного запроса
	for key, client := range rs.list {
		if len(client) == 0 {
			delete(rs.list, key)
		}
	}
}
