package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"maxchan/models"
)

var db *sql.DB
var tmpl *template.Template

func Init(database *sql.DB) {
	db = database

	funcMap := template.FuncMap{
		"formatPrice": func(price int) string {
			return fmt.Sprintf("%d ₽", price)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("02.01.2006")
		},
		"avgRating": func(reviews []models.Review) float64 {
			if len(reviews) == 0 {
				return 0
			}
			var sum int
			for _, r := range reviews {
				sum += r.Rating
			}
			return float64(sum) / float64(len(reviews))
		},
		// Добавляем новые функции
		"repeat": func(s string, count int) string {
			result := ""
			for i := 0; i < count; i++ {
				result += s
			}
			return result
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}

	var err error
	tmpl, err = template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Ошибка загрузки шаблонов:", err)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	rows, err := db.Query("SELECT id, name, price, description, image_url FROM products ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		rows.Scan(&p.ID, &p.Name, &p.Price, &p.Description, &p.ImageURL)
		products = append(products, p)
	}

	data := map[string]interface{}{
		"Products": products,
		"IsAdmin":  isAdmin(r),
	}

	tmpl.ExecuteTemplate(w, "index.html", data)
}

func ProductDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	id := strings.TrimPrefix(r.URL.Path, "/product/")
	pid, _ := strconv.Atoi(id)

	var product models.Product
	err := db.QueryRow("SELECT id, name, price, description, full_desc, image_url FROM products WHERE id = ?", pid).
		Scan(&product.ID, &product.Name, &product.Price, &product.Description, &product.FullDesc, &product.ImageURL)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Получаем отзывы
	rows, _ := db.Query("SELECT id, user_name, rating, comment, created_at FROM reviews WHERE product_id = ? ORDER BY created_at DESC", pid)
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var r models.Review
		rows.Scan(&r.ID, &r.UserName, &r.Rating, &r.Comment, &r.CreatedAt)
		reviews = append(reviews, r)
	}

	data := map[string]interface{}{
		"Product": product,
		"Reviews": reviews,
		"IsAdmin": isAdmin(r),
	}

	tmpl.ExecuteTemplate(w, "product.html", data)
}

func Support(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := map[string]interface{}{
		"IsAdmin": isAdmin(r),
	}

	err := tmpl.ExecuteTemplate(w, "support.html", data)
	if err != nil {
		log.Printf("Ошибка отображения support.html: %v", err)
		http.Error(w, err.Error(), 500)
	}
}

// Проверка админа
func isAdmin(r *http.Request) bool {
	cookie, err := r.Cookie("admin_session")
	if err != nil {
		return false
	}
	return cookie.Value == "authenticated"
}
