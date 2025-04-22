// Функция для создания Refresh токена произвольного типа

package main

import (
	"crypto/rand"     // генерация случайных байтов
	"encoding/base64" // кодирование Refresh токена в формат base64

	"golang.org/x/crypto/bcrypt" // bcrypt-хеширование токена
)

// генерация случайной строки и ее кодировка в base64
func generateString(n int) (string, error) { // принимает: длина строки в байтах, возвращает: закодированная в base64 строка
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// генерация Refresh токена и его хеша в bcrypt виде
func generateRefreshToken() (original string, hashed string, err error) { // ничего не принимает, возвращает: Refresh токен в оригинальном и хешированном виде, ошибка
	original, err = generateString(32)
	if err != nil {
		return "", "", err
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(original), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	return original, string(hashedBytes), nil
}
