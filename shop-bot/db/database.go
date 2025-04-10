package db

import (
	"database/sql"
	"fmt"
	"shop-bot/models"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() error {
	var err error
	DB, err = sql.Open("sqlite3", "./shop.db")
	if err != nil {
		return err
	}

	// Создание таблиц
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			name_tag TEXT,
			balance REAL DEFAULT 0.0,
			history TEXT DEFAULT ''
		);
		CREATE TABLE IF NOT EXISTS goods (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			value REAL,
			descr TEXT
		);
		CREATE TABLE IF NOT EXISTS services (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			value REAL,
			descr TEXT
		);
	`)
	return err
}

func AddUser(user models.User) error {
	_, err := DB.Exec("INSERT OR IGNORE INTO users (id, name_tag, balance, history) VALUES (?, ?, ?, ?)",
		user.ID, user.NameTag, user.Balance, user.History)
	if err != nil {
		fmt.Println("Error adding user:", err)
	}
	return err
}

func AddGood(good models.Good) error {
	_, err := DB.Exec("INSERT INTO goods (name, value, descr) VALUES (?, ?, ?)",
		good.Name, good.Value, good.Descr)
	if err != nil {
		fmt.Println("Error adding good:", err)
	}
	return err
}

func AddService(service models.Service) error {
	_, err := DB.Exec("INSERT INTO services (name, value, descr) VALUES (?, ?, ?)",
		service.Name, service.Value, service.Descr)
	if err != nil {
		fmt.Println("Error adding service:", err)
	}
	return err
}

func GetGoods() ([]models.Good, error) {
	rows, err := DB.Query("SELECT id, name, value, descr FROM goods")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goods []models.Good
	for rows.Next() {
		var g models.Good
		if err := rows.Scan(&g.ID, &g.Name, &g.Value, &g.Descr); err != nil {
			return nil, err
		}
		goods = append(goods, g)
	}
	return goods, nil
}

func GetServices() ([]models.Service, error) {
	rows, err := DB.Query("SELECT id, name, value, descr FROM services")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var s models.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Value, &s.Descr); err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, nil
}

func GetUserBalance(userID int64) (float64, error) {
	var balance float64
	err := DB.QueryRow("SELECT balance FROM users WHERE id = ?", userID).Scan(&balance)
	if err == sql.ErrNoRows {
		// Пользователь не найден, добавляем его
		_, err := DB.Exec("INSERT INTO users (id, name_tag, balance, history) VALUES (?, ?, ?, ?)",
			userID, fmt.Sprintf("user_%d", userID), 0.0, "")
		if err != nil {
			fmt.Println("Error adding user in GetUserBalance:", err)
			return 0, err
		}
		return 0.0, nil
	}
	if err != nil {
		fmt.Println("Error getting user balance:", err)
		return 0, err
	}
	return balance, nil
}

func UpdateUserBalance(userID int64, newBalance float64) error {
	if newBalance < 0 {
		return fmt.Errorf("баланс не может быть отрицательным")
	}
	result, err := DB.Exec("UPDATE users SET balance = ? WHERE id = ?", newBalance, userID)
	if err != nil {
		fmt.Println("Error updating user balance:", err)
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Пользователь не найден, добавляем его
		_, err := DB.Exec("INSERT INTO users (id, name_tag, balance, history) VALUES (?, ?, ?, ?)",
			userID, fmt.Sprintf("user_%d", userID), newBalance, "")
		if err != nil {
			fmt.Println("Error adding user in UpdateUserBalance:", err)
			return fmt.Errorf("ошибка при добавлении пользователя: %v", err)
		}
	}
	return nil
}

func GetPurchaseHistory(userID int64) ([]int, error) {
	var history string
	err := DB.QueryRow("SELECT history FROM users WHERE id = ?", userID).Scan(&history)
	if err == sql.ErrNoRows {
		// Пользователь не найден, добавляем его
		_, err := DB.Exec("INSERT INTO users (id, name_tag, balance, history) VALUES (?, ?, ?, ?)",
			userID, fmt.Sprintf("user_%d", userID), 0.0, "")
		if err != nil {
			fmt.Println("Error adding user in GetPurchaseHistory:", err)
			return nil, err
		}
		return []int{}, nil
	}
	if err != nil {
		fmt.Println("Error getting purchase history:", err)
		return nil, err
	}

	if history == "" {
		return []int{}, nil
	}

	// Разделяем строку истории на массив ID
	historyIDs := strings.Split(history, ",")
	var purchaseIDs []int
	for _, idStr := range historyIDs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		purchaseIDs = append(purchaseIDs, id)
	}
	return purchaseIDs, nil
}

func AddToPurchaseHistory(userID int64, goodID int) error {
	// Получаем текущую историю
	var history string
	err := DB.QueryRow("SELECT history FROM users WHERE id = ?", userID).Scan(&history)
	if err == sql.ErrNoRows {
		// Пользователь не найден, добавляем его
		_, err := DB.Exec("INSERT INTO users (id, name_tag, balance, history) VALUES (?, ?, ?, ?)",
			userID, fmt.Sprintf("user_%d", userID), 0.0, "")
		if err != nil {
			fmt.Println("Error adding user in AddToPurchaseHistory:", err)
			return err
		}
		history = ""
	} else if err != nil {
		fmt.Println("Error getting history in AddToPurchaseHistory:", err)
		return err
	}

	// Добавляем новый ID товара в историю
	if history == "" {
		history = strconv.Itoa(goodID)
	} else {
		history += "," + strconv.Itoa(goodID)
	}

	// Обновляем историю в базе
	_, err = DB.Exec("UPDATE users SET history = ? WHERE id = ?", history, userID)
	if err != nil {
		fmt.Println("Error updating history in AddToPurchaseHistory:", err)
	}
	return err
}
