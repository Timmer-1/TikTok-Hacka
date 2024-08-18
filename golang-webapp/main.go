package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var (
	db   *sql.DB
	mu   sync.Mutex
	tmpl *template.Template
	rdb  *redis.Client
	ctx  = context.Background()
)

type Message struct {
	ID        int
	Content   string
	CreatedAt string
}

type requestData struct {
	MsgID string `json:"msg_id"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis server address
	})
	defer rdb.Close()

	// Initialize database connection
	var err error
	db, err = sql.Open("mysql", "root:Dragon1491@tcp(127.0.0.1:3306)/tiktok_db") // Update UserName and Password
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Load the HTML template
	tmpl = template.Must(template.ParseFiles("static/templates/index.html"))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/notifications", notificationHandler)
	http.HandleFunc("/submitRecommend", submitRecommendedHandler)
	http.HandleFunc("/recommend", getRecommendedHandler)
	http.HandleFunc("/deleteFavorite", deleteFavoriteHandler) // New handler for deleting from favorites

	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	rows, err := db.Query("SELECT id, content, created_at FROM messages ORDER BY created_at DESC")
	mu.Unlock()

	if err != nil {
		http.Error(w, "Error retrieving messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			http.Error(w, "Error scanning message", http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}

	tmpl.Execute(w, messages)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}

	mu.Lock()
	_, err := db.Exec("INSERT INTO messages (content) VALUES (?)", content)
	mu.Unlock()

	if err != nil {
		http.Error(w, "Error saving message", http.StatusInternalServerError)
		return
	}

	// Publish notification to Redis
	err = rdb.Publish(ctx, "notifications", content).Err()
	if err != nil {
		log.Printf("Error publishing to Redis: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func notificationHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	pubsub := rdb.Subscribe(ctx, "notifications")
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Println("Error receiving message from Redis:", err)
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
		if err != nil {
			log.Println("Error writing to WebSocket:", err)
			return
		}
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	log.Println("Received request to delete message with ID:", id) // Log received ID
	if id == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	// Convert id to int and check for errors
	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid ID format:", err)
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Delete the message from the database
	mu.Lock()
	result, err := db.Exec("DELETE FROM messages WHERE id = ?", intID)
	mu.Unlock()

	if err != nil {
		log.Println("Error deleting message:", err) // Log deletion error
		http.Error(w, "Error deleting message", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Error checking deletion result:", err) // Log error when checking rows affected
		http.Error(w, "Error checking deletion result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Println("No message found with the given ID:", id) // Log if no rows were affected
		http.Error(w, "No message found with the given ID", http.StatusNotFound)
		return
	}

	log.Println("Message with ID:", id, "deleted successfully.") // Confirm deletion
	w.WriteHeader(http.StatusOK)
}

func getRecommendedHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("recommend.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mu.Lock()
	rows, err := db.Query(`
		SELECT m.id, m.content, m.created_at
		FROM messages m
		JOIN favorites as f ON m.id = f.message_id
	`)
	mu.Unlock()

	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving messages: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			http.Error(w, "Error scanning message", http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}

	html_err := tmpl.Execute(w, messages)
	if html_err != nil {
		http.Error(w, "Error rendering recommended webpage", http.StatusInternalServerError)
		return
	}
}

func submitRecommendedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Decode the JSON request body
	var data requestData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	msgID := data.MsgID
	if msgID == "" {
		http.Error(w, "msg_id cannot be empty", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(msgID)
	if err != nil {
		http.Error(w, "Invalid msg_id format", http.StatusBadRequest)
		return
	}

	// Insert the new favorite into the database
	mu.Lock()
	_, err = db.Exec("INSERT INTO favorites (message_id) VALUES (?)", id)
	mu.Unlock()
	if err != nil {
		log.Println("Failed to add favorite:", err)
		http.Error(w, "Failed to add favorite", http.StatusInternalServerError)
		return
	}

	log.Println("Successfully added message ID:", id, "to favorites")

	// Return a JSON success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func deleteFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	// Delete the favorite from the database
	mu.Lock()
	result, err := db.Exec("DELETE FROM favorites WHERE message_id = ?", id)
	mu.Unlock()

	if err != nil {
		http.Error(w, "Error removing favorite", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking deletion result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "No favorite found with the given ID", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
