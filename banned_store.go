// сервис единой точки доступа к системе
// реализация функционала хранилища забаненных ip-адресов
package main

import (
	"sync"
	"time"
)

type BannedAddressesStore struct {
	//хэшированный список запросов
	//ключ списка - ip-адрес клиента
	//элемент списка - дата и время окончания бана (в формате epoch)
	list map[string]int64
	//мьютекс для многопоточного доступа
	mutex *sync.Mutex
}

// --------------------------------------------------------
// создание экземпляра хранилища инцидентов
func createBannedAddressesStore() *BannedAddressesStore {
	bas := new(BannedAddressesStore)
	bas.list = make(map[string]int64)
	bas.mutex = &sync.Mutex{}
	return bas
}

// проверка наличия указанного ip-адреса в хранилище забаненных адресов
// @in ip string ip-адрес
// @out bool флаг наличия ip-адреса в списке забаненных
func (bas *BannedAddressesStore) isBanned(ip string) bool {
	bas.mutex.Lock()
	defer bas.mutex.Unlock()
	_, ok := bas.list[ip]
	return ok
}

// добавление ip-адреса в хранилище забаненных адресов
// @in ip string ip-адрес
func (bas *BannedAddressesStore) addToBan(ip string) {
	bas.mutex.Lock()
	defer bas.mutex.Unlock()
	_, ok := bas.list[ip]
	if !ok {
		bas.list[ip] = time.Now().Add(time.Hour * time.Duration(ban_interval)).Unix()
	}
}

// очистка хранилища от старых записей
func (bas *BannedAddressesStore) removeOldRecords() {
	current_time := time.Now().Unix()
	bas.mutex.Lock()
	defer bas.mutex.Unlock()
	for key, value := range bas.list {
		if value <= current_time {
			delete(bas.list, key)
		}
	}
}
