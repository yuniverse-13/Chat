// go-chat-backend/main.go
package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// upgrader "повышает" обычное HTTP-соединение до постоянного WebSocket-соединения.
var upgrader = websocket.Upgrader{
	// Эта функция определяет, разрешено ли соединение с данного источника (origin).
	// Для разработки мы временно разрешаем все подключения.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// handleConnections будет вызываться для каждого нового клиента, подключившегося по WebSocket.
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Повышаем HTTP-запрос до WebSocket.
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Важно закрыть соединение, когда клиент отключается.
	defer ws.Close()

	log.Println("Клиент успешно подключился!")

	// Бесконечный цикл для чтения сообщений от клиента.
	for {
		// Читаем сообщение. ReadMessage блокирует выполнение, пока не придет сообщение.
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Ошибка чтения сообщения: %v", err)
			break // Выходим из цикла, если клиент отсоединился.
		}

		// Выводим полученное сообщение в консоль сервера.
		log.Printf("Получено сообщение: %s", message)

		// Отправляем то же самое сообщение обратно клиенту (Эхо-сервер).
		err = ws.WriteMessage(messageType, message)
		if err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
			break
		}
	}
}

func main() {
	// Создаем простой файловый сервер для статики (пока не используется, но может пригодиться).
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// Регистрируем наш обработчик для WebSocket-соединений по пути "/ws".
	http.HandleFunc("/ws", handleConnections)

	// Запускаем HTTP-сервер на порту 8080.
	log.Println("HTTP-сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}