package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"

    "maxchan/handlers"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    fmt.Println(" Запуск МаксЧан...")

    // Создаем папки
    os.MkdirAll("static/uploads", 0755)
    os.MkdirAll("static/css", 0755)

    // База данных
    db, err := sql.Open("sqlite3", "./maxchan.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    db.SetMaxOpenConns(1)
    
    if err = db.Ping(); err != nil {
        log.Fatal(err)
    }
    
    // Инициализация БД
    initDB(db)

    // Инициализация обработчиков
    handlers.Init(db)

    // Статические файлы
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    // Публичные маршруты
    http.HandleFunc("/", handlers.Index)
    http.HandleFunc("/product/", handlers.ProductDetail)
    http.HandleFunc("/support", handlers.Support)

    // Админ маршруты
    http.HandleFunc("/admin", handlers.AdminLogin)
    http.HandleFunc("/admin/login", handlers.AdminLogin)
    http.HandleFunc("/admin/logout", handlers.AdminLogout)
    http.HandleFunc("/admin/dashboard", handlers.AdminDashboard)
    http.HandleFunc("/admin/create", handlers.CreateProduct)
    http.HandleFunc("/admin/edit/", handlers.EditProduct)
    http.HandleFunc("/admin/delete/", handlers.DeleteProduct)

    // Запуск сервера
    port := ":8081"
    fmt.Printf(" Сервер запущен на http://localhost%s\n", port)
    fmt.Println(" Админка: http://localhost:8081/admin")
    
    if err := http.ListenAndServe(port, nil); err != nil {
        log.Fatal(err)
    }
}

func initDB(db *sql.DB) {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            email TEXT UNIQUE,
            password TEXT,
            is_admin BOOLEAN DEFAULT FALSE,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
        
        `CREATE TABLE IF NOT EXISTS products (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT,
            price INTEGER,
            description TEXT,
            full_desc TEXT,
            image_url TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
        
        `CREATE TABLE IF NOT EXISTS reviews (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            product_id INTEGER,
            user_name TEXT,
            rating INTEGER CHECK(rating >= 1 AND rating <= 5),
            comment TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY(product_id) REFERENCES products(id)
        )`,
        
        // Создаем админа (пароль: admin123 - замените на реальный хеш)
        `INSERT OR IGNORE INTO users (email, password, is_admin) 
        VALUES ('admin@maxchan.ru', '$2a$10$YourHashedPasswordHere', TRUE)`,
    }

    for _, q := range queries {
        db.Exec(q)
    }

    // Тестовый товар
    var count int
    db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
    
    if count == 0 {
        db.Exec(`
            INSERT INTO products (name, price, description, full_desc, image_url) VALUES 
            ('Чан "Сибирский кедр"', 89900, 'Чан на 4-6 человек из сибирского кедра', 
            'Настоящий сибирский чан ручной работы. Диаметр: 180см, высота: 70см. В комплекте: печка, крышка, лестница.', 
            '/static/uploads/sample.jpg')`)
    }
}
