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

	// Переменная для хранения состояния
	type clientState struct {
		lastMessageID int // ID последнего отправленного сообщения
	}
	state := clientState{}

	for update := range updates {
		// Обработка текстовых сообщений
		if update.Message != nil {
			userID := update.Message.From.ID
			nameTag := update.Message.From.UserName
			chatID := update.Message.Chat.ID

			// Добавляем пользователя в базу
			if err := db.AddUser(models.User{ID: userID, NameTag: nameTag}); err != nil {
				log.Println("Failed to add user:", err)
			}

			switch update.Message.Text {
			case "/start":
				// Удаляем предыдущее сообщение, если оно есть
				if state.lastMessageID != 0 {
					deleteMsg := tgbotapi.NewDeleteMessage(chatID, state.lastMessageID)
					if _, err := bot.Send(deleteMsg); err != nil {
						log.Println("Failed to delete message:", err)
					}
				}

				// Отправляем сообщение с inline-кнопками
				msg := tgbotapi.NewMessage(chatID, "Добро пожаловать в магазин! Выберите опцию:")
				msg.ReplyMarkup = clientMenu()
				sentMsg, err := bot.Send(msg)
				if err != nil {
					log.Println("Failed to send message:", err)
					continue
				}
				// Сохраняем ID отправленного сообщения
				state.lastMessageID = sentMsg.MessageID

			default:
				// Если пользователь отправил что-то, что не является командой
				msg := tgbotapi.NewMessage(chatID, "Пожалуйста, используйте кнопки для взаимодействия.")
				sentMsg, err := bot.Send(msg)
				if err != nil {
					log.Println("Failed to send message:", err)
					continue
				}
				// Сохраняем ID нового сообщения
				state.lastMessageID = sentMsg.MessageID
			}
		}

		// Обработка нажатий на inline-кнопки
		if update.CallbackQuery != nil {
			callback := update.CallbackQuery
			chatID := callback.Message.Chat.ID
			userID := callback.From.ID
			var response string
			var msg tgbotapi.MessageConfig

			// Удаляем предыдущее сообщение (например, меню или список товаров/услуг)
			if callback.Message.MessageID != 0 {
				deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
				if _, err := bot.Send(deleteMsg); err != nil {
					log.Println("Failed to delete message:", err)
				}
				// Обновляем state.lastMessageID, так как сообщение удалено
				if state.lastMessageID == callback.Message.MessageID {
					state.lastMessageID = 0
				}
			}

			switch callback.Data {
			case "back_to_menu":
				// Возвращаемся к главному меню
				response = "Добро пожаловать в магазин! Выберите опцию:"
				msg = tgbotapi.NewMessage(chatID, response)
				msg.ReplyMarkup = clientMenu()

			case "catalog":
				// Показываем inline-кнопки для выбора: Товары или Услуги
				response = "Выберите категорию:"
				msg = tgbotapi.NewMessage(chatID, response)
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Товары", "show_goods"),
						tgbotapi.NewInlineKeyboardButtonData("Услуги", "show_services"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "back_to_menu"),
					),
				)

			case "profile":
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
				response = fmt.Sprintf(
					"*Ваш баланс:* %.2f ₽ 💰\n\n"+
						"🆔 ID: %d\n"+
						"🛍️ Количество покупок: %d\n\n"+
						"━━━━━━━━━━━━",
					balance,
					userID,
					len(history),
				)
				msg = tgbotapi.NewMessage(chatID, response)
				msg.ParseMode = "Markdown"
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
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "back_to_menu"),
					),
				)

			case "info":
				response = "Напишите @SupportBot для помощи."
				msg = tgbotapi.NewMessage(chatID, response)
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "back_to_menu"),
					),
				)

			case "show_goods":
				goods, err := db.GetGoods()
				if err != nil {
					log.Println("Failed to get goods:", err)
					response = "Ошибка при получении товаров."
				} else if len(goods) == 0 {
					response = "Товаров пока нет."
				} else {
					response = "Доступные товары:"
					var buttons [][]tgbotapi.InlineKeyboardButton
					for _, good := range goods {
						buttonText := fmt.Sprintf("%s (%.2f ₽)", good.Name, good.Value)
						buttonData := fmt.Sprintf("good_%d", good.ID)
						buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonData),
						))
					}
					// Добавляем кнопку "Назад"
					buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "back_to_menu"),
					))
					msg = tgbotapi.NewMessage(chatID, response)
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
					sentMsg, err := bot.Send(msg)
					if err != nil {
						log.Println("Failed to send message:", err)
						continue
					}
					// Сохраняем ID нового сообщения
					state.lastMessageID = sentMsg.MessageID
					// Подтверждаем обработку callback
					callbackConfig := tgbotapi.NewCallback(callback.ID, "")
					bot.Send(callbackConfig)
					continue
				}

			case "show_services":
				services, err := db.GetServices()
				if err != nil {
					log.Println("Failed to get services:", err)
					response = "Ошибка при получении услуг."
				} else if len(services) == 0 {
					response = "Услуг пока нет."
				} else {
					response = "Доступные услуги:"
					var buttons [][]tgbotapi.InlineKeyboardButton
					for _, service := range services {
						buttonText := fmt.Sprintf("%s (%.2f ₽)", service.Name, service.Value)
						buttonData := fmt.Sprintf("service_%d", service.ID)
						buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonData),
						))
					}
					// Добавляем кнопку "Назад"
					buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "back_to_menu"),
					))
					msg = tgbotapi.NewMessage(chatID, response)
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
					sentMsg, err := bot.Send(msg)
					if err != nil {
						log.Println("Failed to send message:", err)
						continue
					}
					// Сохраняем ID нового сообщения
					state.lastMessageID = sentMsg.MessageID
					// Подтверждаем обработку callback
					callbackConfig := tgbotapi.NewCallback(callback.ID, "")
					bot.Send(callbackConfig)
					continue
				}

			case "top_up_balance":
				response = "Функция пополнения баланса пока в разработке! 💸"
				msg = tgbotapi.NewMessage(chatID, response)
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "back_to_menu"),
					),
				)

			case "referral_system":
				response = "Реферальная система: пригласите друга и получите бонус! 🤑"
				msg = tgbotapi.NewMessage(chatID, response)
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "back_to_menu"),
					),
				)

			case "purchase_history":
				// Получаем историю покупок
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
				msg = tgbotapi.NewMessage(chatID, response)
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "back_to_menu"),
					),
				)

			case "activate_coupon":
				response = "Введите код купона для активации! 📜"
				msg = tgbotapi.NewMessage(chatID, response)
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "back_to_menu"),
					),
				)

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
					msg = tgbotapi.NewMessage(chatID, response)
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "back_to_menu"),
						),
					)

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
					msg = tgbotapi.NewMessage(chatID, response)
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "back_to_menu"),
						),
					)

				} else {
					response = "Неизвестное действие."
					msg = tgbotapi.NewMessage(chatID, response)
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "back_to_menu"),
						),
					)
				}
			}

			// Отправляем ответ на нажатие кнопки
			sentMsg, err := bot.Send(msg)
			if err != nil {
				log.Println("Failed to send message:", err)
				continue
			}
			// Сохраняем ID нового сообщения
			state.lastMessageID = sentMsg.MessageID

			// Подтверждаем обработку callback
			callbackConfig := tgbotapi.NewCallback(callback.ID, "")
			bot.Send(callbackConfig)
		}
	}
}

func clientMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛒 Каталог", "catalog"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Профиль", "profile"),
			tgbotapi.NewInlineKeyboardButtonData("ℹ Информация", "info"),
		),
	)
}
