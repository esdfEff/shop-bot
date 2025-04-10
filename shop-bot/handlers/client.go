package handlers

import (
	"fmt"
	"log"
	"shop-bot/db"
	"shop-bot/models"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleClientBot(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// Обработка текстовых сообщений
		if update.Message != nil {
			userID := update.Message.From.ID
			nameTag := update.Message.From.UserName

			// Добавляем пользователя в базу
			if err := db.AddUser(models.User{ID: userID, NameTag: nameTag}); err != nil {
				log.Println("Failed to add user:", err)
			}

			switch update.Message.Text {
			case "/start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать в магазин! Выберите опцию:")
				msg.ReplyMarkup = clientMenu()
				bot.Send(msg)

			case "Товары":
				goods, err := db.GetGoods()
				if err != nil {
					log.Println("Failed to get goods:", err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении товаров.")
					bot.Send(msg)
					continue
				}
				if len(goods) == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Товаров пока нет.")
					bot.Send(msg)
					continue
				}

				// Создаем сообщение с текстом
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Доступные товары:")
				var buttons [][]tgbotapi.InlineKeyboardButton
				for _, good := range goods {
					// Создаем кнопку для каждого товара
					buttonText := fmt.Sprintf("%s (%.2f ₽)", good.Name, good.Value)
					buttonData := fmt.Sprintf("good_%d", good.ID)
					buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonData),
					))
				}
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
				bot.Send(msg)

			case "Профиль":
				// Получаем баланс пользователя
				balance, err := db.GetUserBalance(userID)
				if err != nil {
					log.Println("Failed to get user balance:", err)
					balance = 0.0
				}

				// Получаем историю покупок
				history, err := db.GetPurchaseHistory(userID)
				if err != nil {
					log.Println("Failed to get purchase history:", err)
					history = []int{}
				}

				// Форматируем профиль
				profileText := fmt.Sprintf(
					"*Ваш баланс:* %.2f ₽ 💰\n\n"+
						"🆔 ID: %d\n"+
						"🛍️ Количество покупок: %d\n\n"+
						"━━━━━━━━━━━━",
					balance,
					userID,
					len(history),
				)

				// Создаем сообщение с текстом
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, profileText)
				msg.ParseMode = "Markdown"

				// Добавляем inline-кнопки
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("💸 Пополнить баланс", "top_up_balance"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("🤑🤑 Реферальная система", "referral_system"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("🛒 История покупок", "purchase_history"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("📜 Активировать купон", "activate_coupon"),
					),
				)

				// Отправляем сообщение
				bot.Send(msg)

			case "Услуги":
				services, err := db.GetServices()
				if err != nil {
					log.Println("Failed to get services:", err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении услуг.")
					bot.Send(msg)
					continue
				}
				if len(services) == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Услуг пока нет.")
					bot.Send(msg)
					continue
				}

				// Создаем сообщение с текстом
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Доступные услуги:")
				var buttons [][]tgbotapi.InlineKeyboardButton
				for _, service := range services {
					// Создаем кнопку для каждой услуги
					buttonText := fmt.Sprintf("%s (%.2f ₽)", service.Name, service.Value)
					buttonData := fmt.Sprintf("service_%d", service.ID)
					buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonData),
					))
				}
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
				bot.Send(msg)

			case "Поддержка":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Напишите @SupportBot для помощи.")
				bot.Send(msg)
			}
		}

		// Обработка нажатий на inline-кнопки
		if update.CallbackQuery != nil {
			callback := update.CallbackQuery
			chatID := callback.Message.Chat.ID
			var response string

			switch callback.Data {
			case "top_up_balance":
				response = "Функция пополнения баланса пока в разработке! 💸"
			case "referral_system":
				response = "Реферальная система: пригласите друга и получите бонус! 🤑"
			case "purchase_history":
				// Получаем историю покупок
				userID := callback.From.ID
				history, err := db.GetPurchaseHistory(userID)
				if err != nil {
					log.Println("Failed to get purchase history:", err)
					response = "Ошибка при получении истории покупок. 🛒"
				} else if len(history) == 0 {
					response = "История покупок: у вас пока нет покупок. 🛒"
				} else {
					response = "История покупок:\n"
					goods, _ := db.GetGoods()
					for _, goodID := range history {
						for _, good := range goods {
							if good.ID == goodID {
								response += fmt.Sprintf("ID: %d, %s - %.2f\n", good.ID, good.Name, good.Value)
							}
						}
					}
				}
			case "activate_coupon":
				response = "Введите код купона для активации! 📜"
			default:
				// Обработка нажатия на кнопку товара
				if strings.HasPrefix(callback.Data, "good_") {
					goodIDStr := strings.TrimPrefix(callback.Data, "good_")
					goodID, err := strconv.Atoi(goodIDStr)
					if err != nil {
						response = "Ошибка: неверный ID товара."
					} else {
						goods, err := db.GetGoods()
						if err != nil {
							response = "Ошибка при получении информации о товаре."
						} else {
							for _, good := range goods {
								if good.ID == goodID {
									response = fmt.Sprintf(
										"Товар: %s\nЦена: %.2f ₽\nОписание: %s",
										good.Name, good.Value, good.Descr,
									)
									break
								}
							}
							if response == "" {
								response = "Товар не найден."
							}
						}
					}
				} else if strings.HasPrefix(callback.Data, "service_") {
					// Обработка нажатия на кнопку услуги
					serviceIDStr := strings.TrimPrefix(callback.Data, "service_")
					serviceID, err := strconv.Atoi(serviceIDStr)
					if err != nil {
						response = "Ошибка: неверный ID услуги."
					} else {
						services, err := db.GetServices()
						if err != nil {
							response = "Ошибка при получении информации об услуге."
						} else {
							for _, service := range services {
								if service.ID == serviceID {
									response = fmt.Sprintf(
										"Услуга: %s\nЦена: %.2f ₽\nОписание: %s",
										service.Name, service.Value, service.Descr,
									)
									break
								}
							}
							if response == "" {
								response = "Услуга не найдена."
							}
						}
					}
				} else {
					response = "Неизвестное действие."
				}
			}

			// Отправляем ответ на нажатие кнопки
			msg := tgbotapi.NewMessage(chatID, response)
			bot.Send(msg)

			// Подтверждаем обработку callback
			callbackConfig := tgbotapi.NewCallback(callback.ID, "")
			bot.Send(callbackConfig)
		}
	}
}

func clientMenu() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Товары"),
			tgbotapi.NewKeyboardButton("Профиль"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Услуги"),
			tgbotapi.NewKeyboardButton("Поддержка"),
		),
	)
}
