package main

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestValidID(t *testing.T) {
	valid := "3fa85f64-5717-4562-b3fc-2c963f66afa6"
	if !validID(valid) {
		t.Errorf("Ожидалось сооответствие GUID", valid)
	}
	invalid := "not-a-guid"
	if validID(invalid) {
		t.Errorf("Ожидалось несооответствие GUID", invalid)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	orig, hashed, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("Ошибка функции generateRefreshToken", err)
	}
	if orig == "" {
		t.Error("Ожидался непустой токен")
	}
	if cost, err := bcrypt.Cost([]byte(hashed)); err != nil {
		t.Errorf("Результат bcrypt хешированного токена невалиден", err)
	} else if cost == 0 {
		t.Error("Не ожидался пустой результат bcrypt")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(orig)); err != nil {
		t.Errorf("Токен не соответсвует значению хеша", err)
	}
}
