package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"maxchan/models"
)

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем правильный заголовок
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if r.Method == "GET" {
		tmpl.ExecuteTemplate(w, "admin_login.html", nil)
		return
	}

	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Простая проверка
		if email == "admin@maxchan.ru" && password == "admin123" {
			http.SetCookie(w, &http.Cookie{
				Name:     "admin_session",
				Value:    "authenticated",
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			})
			http.Redirect(w, r, "/admin/dashboard", 302)
			return
		}

		tmpl.ExecuteTemplate(w, "admin_login.html", map[string]interface{}{
			"Error": "Неверный email или пароль",
		})
	}
}

func AdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/admin", 302)
}

func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if !isAdmin(r) {
		http.Redirect(w, r, "/admin", 302)
		return
	}

	rows, _ := db.Query("SELECT id, name, price, image_url FROM products ORDER BY created_at DESC")
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		rows.Scan(&p.ID, &p.Name, &p.Price, &p.ImageURL)
		products = append(products, p)
	}

	data := map[string]interface{}{
		"Products": products,
		"IsAdmin":  true,
	}

	tmpl.ExecuteTemplate(w, "admin_dashboard.html", data)
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	if !isAdmin(r) {
		http.Redirect(w, r, "/admin", 302)
		return
	}

	if r.Method == "POST" {
		name := r.FormValue("name")
		price, _ := strconv.Atoi(r.FormValue("price"))
		desc := r.FormValue("description")
		fullDesc := r.FormValue("full_desc")

		file, handler, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "Ошибка загрузки изображения", 500)
			return
		}
		defer file.Close()

		filename := time.Now().Format("20060102150405") + "_" + handler.Filename
		fpath := filepath.Join("static/uploads", filename)
		dst, _ := os.Create(fpath)
		defer dst.Close()
		io.Copy(dst, file)

		db.Exec("INSERT INTO products (name, price, description, full_desc, image_url) VALUES (?, ?, ?, ?, ?)",
			name, price, desc, fullDesc, "/static/uploads/"+filename)

		http.Redirect(w, r, "/admin/dashboard", 302)
	}
}

func EditProduct(w http.ResponseWriter, r *http.Request) {
	if !isAdmin(r) {
		http.Redirect(w, r, "/admin", 302)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/admin/edit/")
	pid, _ := strconv.Atoi(id)

	if r.Method == "GET" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		var product models.Product
		db.QueryRow("SELECT id, name, price, description, full_desc, image_url FROM products WHERE id = ?", pid).
			Scan(&product.ID, &product.Name, &product.Price, &product.Description, &product.FullDesc, &product.ImageURL)

		data := map[string]interface{}{
			"Product": product,
			"IsAdmin": true,
		}
		tmpl.ExecuteTemplate(w, "admin_edit.html", data)
		return
	}

	if r.Method == "POST" {
		name := r.FormValue("name")
		price, _ := strconv.Atoi(r.FormValue("price"))
		desc := r.FormValue("description")
		fullDesc := r.FormValue("full_desc")

		file, handler, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			filename := time.Now().Format("20060102150405") + "_" + handler.Filename
			fpath := filepath.Join("static/uploads", filename)
			dst, _ := os.Create(fpath)
			defer dst.Close()
			io.Copy(dst, file)

			db.Exec("UPDATE products SET name=?, price=?, description=?, full_desc=?, image_url=? WHERE id=?",
				name, price, desc, fullDesc, "/static/uploads/"+filename, pid)
		} else {
			db.Exec("UPDATE products SET name=?, price=?, description=?, full_desc=? WHERE id=?",
				name, price, desc, fullDesc, pid)
		}

		http.Redirect(w, r, "/admin/dashboard", 302)
	}
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	if !isAdmin(r) {
		http.Redirect(w, r, "/admin", 302)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/admin/delete/")
	pid, _ := strconv.Atoi(id)

	db.Exec("DELETE FROM reviews WHERE product_id = ?", pid)
	db.Exec("DELETE FROM products WHERE id = ?", pid)

	http.Redirect(w, r, "/admin/dashboard", 302)
}
