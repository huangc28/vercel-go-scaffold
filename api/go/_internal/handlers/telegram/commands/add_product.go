package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/fx"
)

type AddProductCommand struct {
	dao        *CommandDAO
	productDAO *ProductDAO
	botAPI     *tgbotapi.BotAPI
}

type AddProductCommandParams struct {
	fx.In

	DAO        *CommandDAO
	ProductDAO *ProductDAO
	BotAPI     *tgbotapi.BotAPI
}

// UserState represents the current state of product creation process
type UserState struct {
	Step         string      `json:"step"`
	Product      ProductData `json:"product"`
	Specs        []string    `json:"specs"`
	ImageFileIDs []string    `json:"image_file_ids"`
}

type ProductData struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Description string  `json:"description"`
}

func NewAddProductCommand(p AddProductCommandParams) *AddProductCommand {
	return &AddProductCommand{
		dao:        p.DAO,
		productDAO: p.ProductDAO,
		botAPI:     p.BotAPI,
	}
}

// Before complete creating product, user can choose
//  1. 取消 ---> Remove `add_product` type of this userID session in DB
//  2. 跳過 ---> Skip current step, give empty value
//  3. 暫存 ---> Save state, quit, user can resume by /.add_product, it will resume from the step user left off
func (c *AddProductCommand) Handle(msg *tgbotapi.Message) error {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID
	text := msg.Text

	// Retrieve or create user session state
	state, err := c.getOrCreateUserState(ctx, userID, chatID, text)
	if err != nil {
		return fmt.Errorf("failed to get user state: %w", err)
	}

	if state == nil {
		return c.sendMessage(chatID, "請使用 /add_product 開始上架商品。")
	}

	// Handle different steps in the state machine
	if err := c.handleStateStep(ctx, state, text, userID, chatID, msg); err != nil {
		return err
	}

	// Save updated state back to database
	if err := c.dao.UpdateUserSession(ctx, userID, "add_product", state); err != nil {
		return fmt.Errorf("failed to save user state: %w", err)
	}

	return nil
}

// getOrCreateUserState retrieves existing session or creates new one
func (c *AddProductCommand) getOrCreateUserState(ctx context.Context, userID int64, chatID int64, text string) (*UserState, error) {
	// Try to get existing session
	session, err := c.dao.GetUserSession(ctx, userID, "add_product")
	if err != nil {
		return nil, fmt.Errorf("failed to get user session: %w", err)
	}

	if session == nil {
		// If no session exists and command is /add_product, create new one
		state := &UserState{Step: "sku"}
		if err := c.dao.CreateUserSession(ctx, chatID, userID, "add_product", state); err != nil {
			return nil, fmt.Errorf("failed to create user session: %w", err)
		}
		// Inform user about new session
		c.sendMessage(chatID, "🆕 開始新的商品上架流程")
		c.sendMessage(chatID, "請輸入商品 SKU：")
		return state, nil
	}

	// Parse existing session state
	var state UserState
	if err := json.Unmarshal(session.State, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session state: %w", err)
	}

	// Inform user about existing session when they use /add_product command
	currentStepMsg := c.getStepDescription(state.Step)
	resumeMsg := fmt.Sprintf("📋 發現未完成的商品上架流程\n當前步驟: %s\n\n您可以:\n• 繼續輸入以完成當前步驟\n• 輸入 /cancel 取消流程\n• 輸入 /restart 重新開始", currentStepMsg)
	c.sendMessage(chatID, resumeMsg)

	// Send the prompt for current step
	stepPrompt := c.getStepPrompt(state.Step)
	if stepPrompt != "" {
		c.sendMessage(chatID, stepPrompt)
	}

	return &state, nil
}

// handleStateStep processes the current step in the state machine
func (c *AddProductCommand) handleStateStep(ctx context.Context, state *UserState, text string, userID int64, chatID int64, msg *tgbotapi.Message) error {
	// Handle special commands first
	switch text {
	case "/cancel":
		c.sendMessage(chatID, "❌ 已取消商品上架流程")
		return c.dao.DeleteUserSession(ctx, userID, "add_product")
	case "/restart":
		c.sendMessage(chatID, "🔄 重新開始商品上架流程")
		// Reset state to beginning
		state.Step = "sku"
		state.Product = ProductData{}
		state.Specs = []string{}
		state.ImageFileIDs = []string{}
		return c.sendMessage(chatID, "請輸入商品 SKU：")
	}

	switch state.Step {
	case "sku":
		state.Product.SKU = text
		state.Step = "name"
		return c.sendMessage(chatID, "請輸入商品名稱：")
	case "name":
		state.Product.Name = text
		state.Step = "category"
		return c.sendMessage(chatID, "請輸入商品類別：")
	case "category":
		state.Product.Category = text
		state.Step = "price"
		return c.sendMessage(chatID, "請輸入商品價格：")
	case "price":
		price, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return c.sendMessage(chatID, "❌ 價格格式錯誤，請輸入數字：")
		}
		state.Product.Price = price
		state.Step = "stock"
		return c.sendMessage(chatID, "請輸入商品庫存數量：")
	case "stock":
		stock, err := strconv.Atoi(text)
		if err != nil {
			return c.sendMessage(chatID, "❌ 庫存格式錯誤，請輸入整數：")
		}
		state.Product.Stock = stock
		state.Step = "description"
		return c.sendMessageWithButtons(chatID, "請輸入商品描述：", "description")
	case "description":
		state.Product.Description = text
		state.Step = "specs"
		return c.sendMessageWithButtons(chatID, "請輸入商品規格（每行一項，輸入 /done 完成）：", "specs")
	case "specs":
		if text == "/done" {
			state.Step = "images"
			return c.sendMessageWithButtons(chatID, "請上傳商品圖片（最多 5 張，輸入 /done 完成）：", "images")
		}
		state.Specs = append(state.Specs, text)
		return c.sendMessage(chatID, "✅ 規格已新增，繼續輸入或輸入 /done 完成：")
	case "images":
		if text == "/done" {
			if len(state.ImageFileIDs) == 0 {
				return c.sendMessage(chatID, "⚠️ 請至少上傳一張商品圖片，或輸入 /done 跳過此步驟")
			}
			state.Step = "confirm"
			return c.sendSummary(chatID, state)
		} else if msg.Photo != nil {
			// Check if maximum limit reached
			const maxImages = 5
			if len(state.ImageFileIDs) >= maxImages {
				return c.sendMessage(chatID, fmt.Sprintf("❌ 最多只能上傳 %d 張圖片，目前已上傳 %d 張\n輸入 /done 完成上傳", maxImages, len(state.ImageFileIDs)))
			}

			fileID := msg.Photo[len(msg.Photo)-1].FileID
			state.ImageFileIDs = append(state.ImageFileIDs, fileID)

			remaining := maxImages - len(state.ImageFileIDs)
			if remaining > 0 {
				return c.sendMessage(chatID, fmt.Sprintf("✅ 圖片已上傳 (%d/%d)，還可上傳 %d 張或輸入 /done 完成", len(state.ImageFileIDs), maxImages, remaining))
			} else {
				return c.sendMessage(chatID, fmt.Sprintf("✅ 圖片已上傳 (%d/%d)，已達上限！輸入 /done 完成", len(state.ImageFileIDs), maxImages))
			}
		}
		return c.sendMessage(chatID, fmt.Sprintf("請上傳商品圖片（最多 %d 張，目前 %d 張），輸入 /done 完成：", 5, len(state.ImageFileIDs)))
	case "confirm":
		if text == "確認" {
			if err := c.productDAO.SaveProduct(ctx, state); err != nil {
				return c.sendMessage(chatID, "❌ 儲存失敗："+err.Error())
			} else {
				c.sendMessage(chatID, "🎉 商品已成功上架！")
			}
			// Clean up session
			return c.dao.DeleteUserSession(ctx, userID, "add_product")
		} else if text == "取消" {
			c.sendMessage(chatID, "❌ 已取消上架流程。")
			return c.dao.DeleteUserSession(ctx, userID, "add_product")
		} else {
			return c.sendMessage(chatID, "請輸入「確認」或「取消」：")
		}
	}

	return nil
}

// sendMessage sends a text message to the chat
func (c *AddProductCommand) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := c.botAPI.Send(msg)
	return err
}

// sendMessageWithButtons sends a message with inline keyboard buttons
func (c *AddProductCommand) sendMessageWithButtons(chatID int64, text string, step string) error {
	msg := tgbotapi.NewMessage(chatID, text)

	// Only add buttons for steps that can be skipped
	if c.canSkipStep(step) {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
				tgbotapi.NewInlineKeyboardButtonData("⏭️ 跳過", fmt.Sprintf("skip_%s", step)),
				tgbotapi.NewInlineKeyboardButtonData("💾 暫存", "pause"),
			),
		)
		msg.ReplyMarkup = keyboard
	}

	_, err := c.botAPI.Send(msg)
	return err
}

// canSkipStep determines if a step can be skipped
func (c *AddProductCommand) canSkipStep(step string) bool {
	skippableSteps := map[string]bool{
		"description": true,
		"specs":       true,
		"images":      true,
	}
	return skippableSteps[step]
}

// sendSummary sends a product summary for confirmation
func (c *AddProductCommand) sendSummary(chatID int64, state *UserState) error {
	summary := fmt.Sprintf(
		"商品摘要：\nSKU: %s\n名稱: %s\n類別: %s\n價格: %.2f\n庫存: %d\n描述: %s\n規格: %v\n圖片數量: %d\n請選擇：",
		state.Product.SKU,
		state.Product.Name,
		state.Product.Category,
		state.Product.Price,
		state.Product.Stock,
		state.Product.Description,
		state.Specs,
		len(state.ImageFileIDs),
	)

	msg := tgbotapi.NewMessage(chatID, summary)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ 確認", "confirm"),
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := c.botAPI.Send(msg)
	return err
}

// getStepDescription returns a user-friendly description of the current step
func (c *AddProductCommand) getStepDescription(step string) string {
	descriptions := map[string]string{
		"sku":         "輸入商品 SKU",
		"name":        "輸入商品名稱",
		"category":    "輸入商品類別",
		"price":       "輸入商品價格",
		"stock":       "輸入商品庫存數量",
		"description": "輸入商品描述",
		"specs":       "輸入商品規格",
		"images":      "上傳商品圖片",
		"confirm":     "確認商品資訊",
	}

	if desc, exists := descriptions[step]; exists {
		return desc
	}
	return "未知步驟"
}

// getStepPrompt returns the prompt message for the current step
func (c *AddProductCommand) getStepPrompt(step string) string {
	prompts := map[string]string{
		"sku":         "請輸入商品 SKU：",
		"name":        "請輸入商品名稱：",
		"category":    "請輸入商品類別：",
		"price":       "請輸入商品價格：",
		"stock":       "請輸入商品庫存數量：",
		"description": "請輸入商品描述：",
		"specs":       "請輸入商品規格（每行一項，輸入 /done 完成）：",
		"images":      "請上傳商品圖片（最多 5 張，輸入 /done 完成）：",
		"confirm":     "請檢查商品資訊，輸入「確認」儲存或「取消」放棄：",
	}

	if prompt, exists := prompts[step]; exists {
		return prompt
	}
	return ""
}

// HandleCallback handles inline keyboard button presses
func (c *AddProductCommand) HandleCallback(callback *tgbotapi.CallbackQuery) error {
	ctx := context.Background()
	userID := callback.From.ID
	chatID := callback.Message.Chat.ID
	data := callback.Data

	// Get current user state
	session, err := c.dao.GetUserSession(ctx, userID, "add_product")
	if err != nil || session == nil {
		return c.sendMessage(chatID, "❌ 未找到活動會話")
	}

	var state UserState
	if err := json.Unmarshal(session.State, &state); err != nil {
		return err
	}

	switch {
	case data == "cancel":
		c.sendMessage(chatID, "❌ 已取消商品上架流程")
		return c.dao.DeleteUserSession(ctx, userID, "add_product")

	case data == "confirm":
		if err := c.productDAO.SaveProduct(ctx, &state); err != nil {
			return c.sendMessage(chatID, "❌ 儲存失敗："+err.Error())
		} else {
			c.sendMessage(chatID, "🎉 商品已成功上架！")
		}
		return c.dao.DeleteUserSession(ctx, userID, "add_product")

	case data == "pause":
		c.sendMessage(chatID, "💾 流程已暫存，您可以稍後使用 /add_product 繼續")
		return nil // Keep session, don't delete

	case len(data) > 5 && data[:5] == "skip_":
		step := data[5:] // Remove "skip_" prefix
		return c.handleSkipStep(ctx, &state, step, userID, chatID)
	}

	return nil
}

// handleSkipStep handles skipping specific steps
func (c *AddProductCommand) handleSkipStep(ctx context.Context, state *UserState, step string, userID int64, chatID int64) error {
	switch step {
	case "description":
		state.Product.Description = "" // Skip with empty value
		state.Step = "specs"
		c.sendMessage(chatID, "⏭️ 已跳過描述")
		return c.sendMessageWithButtons(chatID, "請輸入商品規格（每行一項，輸入 /done 完成）：", "specs")
	case "specs":
		state.Specs = []string{} // Skip with empty specs
		state.Step = "images"
		c.sendMessage(chatID, "⏭️ 已跳過規格")
		return c.sendMessageWithButtons(chatID, "請上傳商品圖片（最多 5 張，輸入 /done 完成）：", "images")
	case "images":
		state.ImageFileIDs = []string{} // Skip with no images
		state.Step = "confirm"
		c.sendMessage(chatID, "⏭️ 已跳過圖片")
		return c.sendSummary(chatID, state)
	}

	// Save updated state
	return c.dao.UpdateUserSession(ctx, userID, "add_product", state)
}

func (c *AddProductCommand) Command() BotCommand {
	return AddProduct
}

var _ CommandHandler = (*AddProductCommand)(nil)
