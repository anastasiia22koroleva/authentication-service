package main

import (
	"encoding/json" // кодирование структуры Token в JSON
	"fmt"
	"log" // вывод ошибок
	"net"
	"net/http" // стандартная HTTP-библиотека (для REST маршрута №1)
	"regexp"   // для проверки id на принадлежность к GUID

	"github.com/go-chi/chi/v5" // библиотека chi (для REST маршрута №2)
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Регулярное выражение для проверки id на принадлежность к GUID
var uuid_check = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

type Token struct {
	AccessToken  string `json:"a_token"` // т.е. ответ в формате JSON,
	RefreshToken string `json:"r_token"` // а поле будет называться A_token
}

func main() {
	connectDB() // настройка БД

	r := chi.NewRouter()

	r.Handle("/giveTokens", http.HandlerFunc(giveTokens)) // REST маршрут №1
	r.Post("/refreshTokens", refreshTokens)               // REST маршрут №2

	// запуск сервера
	log.Println("Запуск сервера на http://localhost:8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("Ошибка запуска сервера", err)
	}
}

// проверка, что строка валидна и является GUID
func validID(s string) bool {
	return uuid_check.MatchString(s)
}

func getIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// REST маршрут №1 (через стандартную HTTP-библиотеку) --------------------
func giveTokens(w http.ResponseWriter, r *http.Request) { // w - ответ клиента (Writer), r - запрос и его составляющие
	if r.Method != http.MethodPost {
		http.Error(w, "Метод недоступен", http.StatusMethodNotAllowed) // защита данных от изменения на стороне клиента и попыток второго использования
		return
	}

	// получение id клиента
	user_ID := r.URL.Query().Get("user_id")

	if user_ID == "" {
		http.Error(w, "id не может быть пустым", http.StatusBadRequest)
		return
	}

	// проверка id на валидность и принадлежность к GUID
	if !validID(user_ID) {
		http.Error(w, "GUID невалиден", http.StatusBadRequest)
		return
	}

	// ip := "1.2.3.4" // для проверки вывода warning
	ip := getIP(r)

	// Access токен
	accessToken, err := generateAccessToken(user_ID, ip)
	if err != nil {
		http.Error(w, "Ошибка при создании Access токена", http.StatusInternalServerError)
		return
	}

	// Refresh токен
	refreshToken, hashedRefreshToken, err := generateRefreshToken()
	if err != nil {
		http.Error(w, "Ошибка при создании Refresh токена", http.StatusInternalServerError)
		return
	}

	// SQL-запрос (новая строка в таблицу) для добавления id клиента и bcrypt-хеша Refresh токена в БД
	_, err = db.Exec(`INSERT INTO refresh_tokens (user_id, token_hash, created_at, ip_address) 
                      VALUES ($1, $2, NOW(), $3)
                      `, user_ID, hashedRefreshToken, ip)
	if err != nil {
		http.Error(w, "Ошибка при сохранении Refresh токена", http.StatusInternalServerError)
		return
	}

	// создание ответа клиенту
	response := Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	// отправка ответа клиенту в JSON
	json.NewEncoder(w).Encode(response)
} // ----------------------------------------------------------------------

// REST маршрут №2 (через библиотеку chi)----------------------------------
func refreshTokens(w http.ResponseWriter, r *http.Request) {
	type RefreshRequest struct {
		AccessToken  string `json:"a_token"`
		RefreshToken string `json:"r_token"`
	}

	// запрос клиента
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// извлечение payload из Access токена и проверки
	token, err := jwt.Parse(req.AccessToken, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS512 {
			return nil, fmt.Errorf("Подпись не поддерживается")
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Access токен невалиден", http.StatusUnauthorized)
		return
	}

	payload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Невозможно извлечь payload", http.StatusInternalServerError)
		return
	}

	// актуальные данные клиента
	userID := payload["sub"].(string)
	originalIP := payload["ip"].(string)
	currentIP := getIP(r)
	//currentIP := "1.2.3.4" // для проверки вывода warning

	// gолучение данных из БД (bcrypt-хеш Refresh токена и ip клиента)
	rows, err := db.Query("SELECT token_hash, ip_address FROM refresh_tokens WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "Ошибка чтения токена из БД", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var matchedHash string
	ipChanged := false
	found := false

	// перебор по user_id, чтобы не было дубликатов
	for rows.Next() {
		var hash, storedIP string
		rows.Scan(&hash, &storedIP)

		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.RefreshToken)) == nil {
			found = true
			matchedHash = hash

			if currentIP != originalIP {
				ipChanged = true
			}

			if currentIP != storedIP {
				ipChanged = true
			}
			break
		}
	}

	if ipChanged {
		log.Println("IP-адрес изменен. Отправка email-предупреждения.")
	}

	if !found {
		http.Error(w, "Неверный Refresh токен", http.StatusUnauthorized)
		return
	}

	// удаление bcrypt-хеша старого Refresh-токена
	_, err = db.Exec("DELETE FROM refresh_tokens WHERE token_hash = $1", matchedHash)
	if err != nil {
		log.Println("Ошибка при удалении прежнего токена", err)
	}

	// новый Acces токен
	newAccess, err := generateAccessToken(userID, currentIP)
	if err != nil {
		http.Error(w, "Ошибка при создании нового Access токена", http.StatusInternalServerError)
		return
	}

	// новый Refresh токен
	newRefresh, newHashRefresh, err := generateRefreshToken()
	if err != nil {
		http.Error(w, "Ошибка при создании нового Refresh токена", http.StatusInternalServerError)
		return
	}

	// добавление обновленных данных в БД
	_, err = db.Exec(`INSERT INTO refresh_tokens (user_id, token_hash, ip_address) VALUES ($1, $2, $3)`,
		userID, newHashRefresh, currentIP)
	if err != nil {
		http.Error(w, "Ошибка при сохранении нового Refresh токена в БД", http.StatusInternalServerError)
		return
	}

	// cоздание ответа клиенту
	response := Token{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
	}

	// // отправка ответа клиенту в JSON
	json.NewEncoder(w).Encode(response)
} // ----------------------------------------------------------------------
