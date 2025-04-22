// Функция для создания Access токена типа JWT по алгоритму SHA-512 c id, ip клиента и со сроком действия (10 мин)

package main

import (
	"time" // срок действия токена

	"github.com/golang-jwt/jwt/v5" // создание JWT токенов
)

// пока жестко прописанная подпись
var jwtSecret = []byte("secretkey")

func generateAccessToken(userID, ip string) (string, error) { // принимает: id и ip клиента, возвращает: сам JWT токен и ошибку (если что-то пошло не так)
	payload := jwt.MapClaims{
		"sub": userID,                                  // заголовок - id клиента
		"ip":  ip,                                      // ip - IP-адрес клиента
		"exp": time.Now().Add(10 * time.Minute).Unix(), // срок действия - 10 мин
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, payload)

	return token.SignedString(jwtSecret)
}
