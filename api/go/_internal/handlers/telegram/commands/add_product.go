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
	dao *CommandDAO
}

type AddProductCommandParams struct {
	fx.In

	DAO *CommandDAO
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
	return &AddProductCommand{dao: p.DAO}
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
		return c.sendMessage(chatID, "請輸入商品描述：")
	case "description":
		state.Product.Description = text
		state.Step = "specs"
		return c.sendMessage(chatID, "請輸入商品規格（每行一項，輸入 /done 完成）：")
	case "specs":
		if text == "/done" {
			state.Step = "images"
			return c.sendMessage(chatID, "請上傳商品圖片（可多張，輸入 /done 完成）：")
		}
		state.Specs = append(state.Specs, text)
		return c.sendMessage(chatID, "✅ 規格已新增，繼續輸入或輸入 /done 完成：")
	case "images":
		if text == "/done" {
			state.Step = "confirm"
			return c.sendSummary(chatID, state)
		} else if msg.Photo != nil {
			fileID := msg.Photo[len(msg.Photo)-1].FileID
			state.ImageFileIDs = append(state.ImageFileIDs, fileID)
			return c.sendMessage(chatID, "✅ 圖片已上傳，繼續上傳或輸入 /done 完成：")
		}
		return c.sendMessage(chatID, "請上傳圖片或輸入 /done 完成：")
	case "confirm":
		if text == "確認" {
			if err := c.saveProduct(ctx, state); err != nil {
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

// sendMessage sends a text message to the chat (implement this method)
func (c *AddProductCommand) sendMessage(chatID int64, text string) error {
	// TODO: Implement actual message sending via bot API
	// This will depend on your bot setup
	fmt.Printf("Sending message to chat %d: %s\n", chatID, text)
	return nil
}

// sendSummary sends a product summary for confirmation
func (c *AddProductCommand) sendSummary(chatID int64, state *UserState) error {
	summary := fmt.Sprintf(
		"商品摘要：\nSKU: %s\n名稱: %s\n類別: %s\n價格: %.2f\n庫存: %d\n描述: %s\n規格: %v\n圖片數量: %d\n請輸入「確認」儲存或「取消」放棄：",
		state.Product.SKU,
		state.Product.Name,
		state.Product.Category,
		state.Product.Price,
		state.Product.Stock,
		state.Product.Description,
		state.Specs,
		len(state.ImageFileIDs),
	)
	return c.sendMessage(chatID, summary)
}

// saveProduct saves the product to the database using raw SQL
func (c *AddProductCommand) saveProduct(ctx context.Context, state *UserState) error {
	// Create product using raw SQL
	query := `
		INSERT INTO products (sku, name, price, category, stock_count, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var productID int64
	err := c.dao.db.QueryRow(query,
		state.Product.SKU,
		state.Product.Name,
		state.Product.Price,
		state.Product.Category,
		state.Product.Stock,
		state.Product.Description,
	).Scan(&productID)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	// Create product specs
	for i, spec := range state.Specs {
		// Assume spec format is "name:value"
		specQuery := `
			INSERT INTO product_specs (product_id, spec_name, spec_value, sort_order)
			VALUES ($1, $2, $3, $4)
		`
		// Simple parsing - you might want to improve this
		specName := spec
		specValue := ""
		if len(spec) > 0 {
			specName = spec
			specValue = spec // For now, store the whole string as both name and value
		}

		_, err := c.dao.db.Exec(specQuery, productID, specName, specValue, i)
		if err != nil {
			return fmt.Errorf("failed to create product spec: %w", err)
		}
	}

	// Create product images
	for i, fileID := range state.ImageFileIDs {
		imageQuery := `
			INSERT INTO product_images (product_id, url, alt_text, is_primary, sort_order)
			VALUES ($1, $2, $3, $4, $5)
		`

		isPrimary := i == 0 // First image is primary
		altText := fmt.Sprintf("%s image %d", state.Product.Name, i+1)
		url := fmt.Sprintf("telegram_file://%s", fileID)

		_, err := c.dao.db.Exec(imageQuery, productID, url, altText, isPrimary, i)
		if err != nil {
			return fmt.Errorf("failed to create product image: %w", err)
		}
	}

	return nil
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
		"images":      "請上傳商品圖片（可多張，輸入 /done 完成）：",
		"confirm":     "請檢查商品資訊，輸入「確認」儲存或「取消」放棄：",
	}

	if prompt, exists := prompts[step]; exists {
		return prompt
	}
	return ""
}

func (c *AddProductCommand) Command() BotCommand {
	return AddProduct
}

var _ CommandHandler = (*AddProductCommand)(nil)
