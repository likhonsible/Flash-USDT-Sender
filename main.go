package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/tronprotocol/go-tron/api"
	"github.com/tronprotocol/go-tron/common"
)

// BotConfig 包含Bot的配置信息
type BotConfig struct {
	Token          string // Bot的token
	AdminChatID    int64  // 管理员的Telegram Chat ID
	RateUSDTtoTRX  float64 // 1 USDT 可以兑换多少TRX
	MaxDecimalsUSDT int    // USDT小数点后最多保留几位
}

// TronConfig 包含Tron的配置信息
type TronConfig struct {
	FullNodeAPI    string // Full Node API的地址
	SolidityAPI    string // Solidity API的地址
	DefaultAccount string // 默认账户地址
	PrivateKey     string // 默认账户私钥
}

func main() {
	// 配置Bot和Tron
	botConfig := &BotConfig{
		Token:          "YOUR_TELEGRAM_BOT_TOKEN", // 替换成你的Telegram Bot的token
		AdminChatID:    YOUR_TELEGRAM_CHAT_ID,     // 替换成你的Telegram Chat ID
		RateUSDTtoTRX:  30,                        // 1 USDT 可以兑换30 TRX
		MaxDecimalsUSDT: 4,                        // USDT小数点后最多保留4位
	}
	tronConfig := &TronConfig{
		FullNodeAPI:    "https://api.trongrid.io", // Tron Full Node API的地址
		SolidityAPI:    "https://api.trongrid.io", // Tron Solidity API的地址
		DefaultAccount: "YOUR_TRON_ACCOUNT_ADDRESS", // 替换成你的Tron账户地址
		PrivateKey:     "YOUR_TRON_ACCOUNT_PRIVATE_KEY", // 替换成你的Tron账户私钥
	}

	// 配置Bot和Tron API
	bot, err := tgbotapi.NewBotAPI(botConfig.Token)
	if err != nil {
		log.Fatal(err)
	}
	tronAPI := api.NewGrpcClient(tronConfig.FullNodeAPI, tronConfig.SolidityAPI)

	// 处理收到的消息
	updates := bot.ListenForWebhook("/")
	for update := range updates {
		if update.Message == nil { // 忽略非消息事件
			continue
		}

		// 解析命令和参数
		command, args := parseCommand(update.Message.Text)

		// 处理命令
		switch command {
		case "/start":
			sendWelcomeMessage(bot, update.Message.Chat.ID)
		case "/setrate":
			err := setExchangeRate(bot, update.Message.Chat.ID, botConfig.AdminChatID, args, botConfig)
			if err != nil {
				sendErrorMessage(bot, update.Message.Chat.ID, err.Error())
			}
		case "/sendusdt":
			err := sendUSDT(bot, update.Message.Chat.ID, tronAPI, args, botConfig, tronConfig)
			if err != nil {
				sendErrorMessage(bot, update.Message.Chat.ID, err.Error())
			
			}
		default:
			sendUnknownCommandMessage(bot, update.Message.Chat.ID)
		}
	}
}

// parseCommand 解析命令和参数
func parseCommand(text string) (command string, args []string) {
parts := strings.Split(text, " ")
if len(parts) > 0 {
command = parts[0]
}
if len(parts) > 1 {
args = parts[1:]
}
return command, args
}

// sendWelcomeMessage 发送欢迎消息
func sendWelcomeMessage(bot *tgbotapi.BotAPI, chatID int64) {
msg := tgbotapi.NewMessage(chatID, "欢迎使用我们的Bot！\n\n使用 /sendusdt 命令发送 USDT 到指定地址，并获得相应的 TRX。\n\n使用 /setrate 命令设置 USDT 和 TRX 的兑换汇率。")
bot.Send(msg)
}

// sendUnknownCommandMessage 发送未知命令消息
func sendUnknownCommandMessage(bot *tgbotapi.BotAPI, chatID int64) {
msg := tgbotapi.NewMessage(chatID, "未知命令，请使用 /start 命令查看可用命令。")
bot.Send(msg)
}

// setExchangeRate 设置 USDT 和 TRX 的兑换汇率
func setExchangeRate(bot *tgbotapi.BotAPI, chatID int64, adminChatID int64, args []string, botConfig *BotConfig) error {
// 检查是否是管理员
if chatID != adminChatID {
return errors.New("你不是管理员，无法执行此操作。")
}
// 检查参数数量
if len(args) != 1 {
	return errors.New("命令格式不正确。请使用 /setrate rate 设置 USDT 和 TRX 的兑换汇率，其中 rate 为数字。")
}

// 解析汇率
rate, err := strconv.ParseFloat(args[0], 64)
if err != nil {
	return errors.New("汇率必须是数字。")
}

// 更新汇率
botConfig.RateUSDTtoTRX = rate

// 发送成功消息
msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("USDT 和 TRX 的兑换汇率已经更新为 %v。", rate))
bot.Send(msg)

return nil
}

// sendUSDT 发送 USDT 并获取 TRX
func sendUSDT(bot *tgbotapi.BotAPI, chatID int64, tronAPI *api.GrpcClient, args []string, botConfig *BotConfig, tronConfig *TronConfig) error {
// 检查参数数量
if len(args) != 2 {
return errors.New("命令格式不正确。请使用 /sendusdt address amount 设置要发送 USDT 的地址和数量，其中 address 是 TRC20 地址，amount 是 USDT 的数量。")
}
// 解析参数
address := args[0]
amount, err := strconv.ParseFloat(args[1], 64)
if err != nil {
	return errors.New("USDT 的数量必须是数字。")
}

// 计算 TRX 的数量
trxAmount := math.Floor(amount * botConfig.RateUSDTtoTRX * math.Pow10(botConfig.MaxDecimalsUSDT)) / math
// 检查 TRX 数量是否超过限制
if trxAmount > botConfig.MaxTRXToSend {
	return errors.New(fmt.Sprintf("一次最多只能发送 %v TRX。", botConfig.MaxTRXToSend))
}

// 检查 TRX 地址是否有效
if !tronAPI.ValidateAddress(tronConfig.MainNet, tronConfig.SolidityNode, tronConfig.EventServer, tronConfig.AddressHex, address) {
	return errors.New("TRC20 地址无效。")
}

// 获取 USDT 合约地址
contractAddress := common.HexToAddress(tronConfig.USDTContractAddress)

// 获取 TRC20 接口
usdt, err := trc20.NewTRC20(contractAddress, tronAPI)
if err != nil {
	return errors.New("无法获取 TRC20 接口。")
}

// 获取 USDT 的精度
decimals, err := usdt.Decimals(nil)
if err != nil {
	return errors.New("无法获取 USDT 精度。")
}

// 将 USDT 数量转换为整数
amountInt := big.NewInt(int64(amount * math.Pow10(decimals)))

// 获取用户的 USDT 余额
balance, err := usdt.BalanceOf(nil, common.HexToAddress(address))
if err != nil {
	return errors.New("无法获取用户的 USDT 余额。")
}

// 检查余额是否充足
if balance.Cmp(amountInt) < 0 {
	return errors.New("用户的 USDT 余额不足。")
}

// 获取用户的 TRX 地址
trxAddress, err := tronAPI.GetAccountAddress(tronConfig.MainNet, tronConfig.SolidityNode, tronConfig.EventServer, tronConfig.AddressHex, address)
if err != nil {
	return errors.New("无法获取用户的 TRX 地址。")
}

// 发送 USDT
tx, err := usdt.Transfer(nil, common.HexToAddress(trxAddress), amountInt)
if err != nil {
	return errors.New("无法发送 USDT。")
}

// 等待交易确认
receipt, err := tronAPI.WaitForTransactionReceipt(tx.Hash().Hex(), tronConfig.MainNet, tronConfig.SolidityNode, tronConfig.EventServer, tronConfig.WaitTimeout)
if err != nil {
	return errors.New("USDT 交易确认失败。")
}

// 获取 TRX 的精度
trxDecimals := int64(math.Pow10(botConfig.MaxDecimalsTRX))

// 发送 TRX
trxTx, err := tronAPI.Transfer(trxAddress, botConfig.AdminAddress, trxAmount, trxDecimals, tronConfig.MainNet, tronConfig.SolidityNode, tronConfig.EventServer)
if err != nil {
	return errors.New("无法发送 TRX。")
}

// 等待交易确认
trxReceipt, err := tronAPI.WaitForTransaction(trxTx, tronConfig.MainNet, tronConfig.SolidityNode, tronConfig.EventServer, tronConfig.WaitTimeout)
if err != nil {
	return errors.New("TRX 交易确认失败。")
}

// 发送成功消息
msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("USDT 交易成功，已发送 %v USDT到地址 %v。TRX 交易成功，已发送 %v TRX 到地址 %v。", amount, address, trxAmount, botConfig.AdminAddress))
_, err = bot.Send(msg)
if err != nil {
return errors.New("无法发送消息到 Telegram。")
}
return nil
}

// handleMessage 处理来自 Telegram 的消息
func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
// 如果消息来自群组或频道，则忽略
if message.Chat.Type != "private" {
return nil
}// 解析命令
command, err := parseCommand(message.Text)
if err != nil {
	return err
}

// 处理命令
switch command.Name {
case "start":
	return handleStartCommand(bot, message.Chat.ID)
case "help":
	return handleHelpCommand(bot, message.Chat.ID)
case "setrate":
	return handleSetRateCommand(bot, message.Chat.ID, command.Args)
case "sendusdt":
	return handleSendUSDTCommand(bot, message.Chat.ID, command.Args)
default:
	return handleUnknownCommand(bot, message.Chat.ID)
}
}

// parseCommand 解析命令
func parseCommand(text string) (*Command, error) {
// 如果不是以 / 开头，则不是命令
if !strings.HasPrefix(text, "/") {
return nil, errors.New("不是有效的命令。")
}
// 按空格分割命令和参数
parts := strings.Split(text, " ")

// 如果只有命令没有参数，则返回空参数
if len(parts) == 1 {
	return &Command{Name: strings.TrimPrefix(parts[0], "/"), Args: ""}, nil
}

// 否则返回命令和参数
return &Command{Name: strings.TrimPrefix(parts[0], "/"), Args: strings.Join(parts[1:], " ")}, nil
}

// handleStartCommand 处理 /start 命令
func handleStartCommand(bot *tgbotapi.BotAPI, chatID int64) error {
msg := tgbotapi.NewMessage(chatID, "欢迎使用 USDT 转账机器人。请使用 /help 命令获取帮助。")
_, err := bot.Send(msg)
if err != nil {
return errors.New("无法发送消息到 Telegram。")
}
return nil
}

// handleHelpCommand 处理 /help 命令
func handleHelpCommand(bot *tgbotapi.BotAPI, chatID int64) error {
msg := tgbotapi.NewMessage(chatID, 使用方法： /setrate 汇率 - 设置 1 USDT 可以兑换多少 TRX。 /sendusdt 地址 数量 - 向指定的 TRC20 地址发送指定数量的 USDT。)
_, err := bot.Send(msg)
if err != nil {
return errors.New("无法发送消息到 Telegram。")
}
return nil
}

// handleSetRateCommand 处理 /setrate 命令
func handleSetRateCommand(bot *tgbotapi.BotAPI, chatID int64, args string) error {
// 检查是否为管理员
if chatID != botConfig.AdminChatID {
return errors.New("只有管理员可以设置汇率。")
}
// 解析汇率
rate, err := strconv
.Atoi(args)
if err != nil {
return errors.New("汇率必须是整数。")
}
// 保存汇率到配置文件
botConfig.ExchangeRate = rate
err = saveConfig(botConfig)
if err != nil {
	return errors.New("无法保存配置文件。")
}

// 发送成功消息
msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("汇率已设置为 %d TRX/USDT。", botConfig.ExchangeRate))
_, err = bot.Send(msg)
if err != nil {
	return errors.New("无法发送消息到 Telegram。")
}

return nil
}

// handleSendUSDTCommand 处理 /sendusdt 命令
func handleSendUSDTCommand(bot *tgbotapi.BotAPI, chatID int64, args string) error {
// 解析地址和数量
parts := strings.Split(args, " ")
if len(parts) != 2 {
return errors.New("用法：/sendusdt 地址 数量。")
}
address := parts[0]
amount, err := strconv.Atoi(parts[1])
if err != nil {
return errors.New("数量必须是整数。")
}
// 调用 USDT 转账函数
err = sendUSDT(address, amount, bot)
if err != nil {
	return err
}

// 发送成功消息
msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("已向地址 %v 发送 %v USDT。", address, amount))
_, err = bot.Send(msg)
if err != nil {
	return errors.New("无法发送消息到 Telegram。")
}

return nil
}

// handleUnknownCommand 处理未知命令
func handleUnknownCommand(bot *tgbotapi.BotAPI, chatID int64) error {
msg := tgbotapi.NewMessage(chatID, "未知命令，请使用 /help 命令获取帮助。")
_, err := bot.Send(msg)
if err != nil {
return errors.New("无法发送消息到 Telegram。")
}
return nil
}

// saveConfig 保存配置文件
func saveConfig(config *Config) error {
// 转换配置为 JSON 字符串
data, err := json.MarshalIndent(config, "", " ")
if err != nil {
return errors.New("无法转换配置文件为 JSON 格式。")
}
// 写入文件
err = ioutil.WriteFile(configFilePath, data, 0644)
if err != nil {
	return errors.New("无法保存配置文件。")
}

return nil
}

// loadConfig 加载配置文件
func loadConfig() (*Config, error) {
// 读取文件
data, err := ioutil.ReadFile(configFilePath)
if err != nil {
return nil, errors.New("无法读取配置文件。")
}
// 解析 JSON 字符串
config := &Config{}
err = json.Unmarshal(data, config)
if err != nil {
	return nil, errors.New("无法解析配置文件。")
}

return config, nil
}

// main 程序入口
func main() {
// 加载配置文件
config, err := loadConfig()
if err != nil {
log.Fatalf("无法加载配置文件：%v", err)
}
botConfig = config
// 创建 Telegram Bot 实例
bot, err := tgbotapi.NewBotAPI(botConfig.BotToken)
if err != nil {
	log.Fatalf("无法创建 Telegram Bot 实例：%v", err)
	}
	// 设置 Webhook
_, err = bot.SetWebhook(tgbotapi.NewWebhook(botConfig.WebhookURL))
if err != nil {
	log.Fatalf("无法设置 Webhook：%v", err)
}

// 启动 HTTP 服务器
http.HandleFunc("/", handleWebhook)
err = http.ListenAndServeTLS(fmt.Sprintf(":%d", botConfig.WebhookPort), botConfig.CertFile, botConfig.KeyFile, nil)
if err != nil {
	log.Fatalf("无法启动 HTTP 服务器：%v", err)
}
}

// handleWebhook 处理 Telegram Webhook
func handleWebhook(w http.ResponseWriter, r *http.Request) {
// 解析请求
update, err := tgbotapi.NewUpdateFromRequest(r)
if err != nil {
log.Printf("无法解析请求：%v", err)
return
}
// 处理更新
err = handleUpdate(update)
if err != nil {
	log.Printf("无法处理更新：%v", err)
	return
}
}

// handleUpdate 处理 Telegram 更新
func handleUpdate(update tgbotapi.Update) error {
// 处理不同类型的更新
if update.Message != nil {
return handleMessageUpdate(update.Message)
}
if update.CallbackQuery != nil {
return handleCallbackQueryUpdate(update.CallbackQuery)
}
// 忽略其它类型的更新
return nil
}

// handleMessageUpdate 处理消息更新
func handleMessageUpdate(message *tgbotapi.Message) error {
// 忽略非文本消息
if message.Text == "" {
return nil
}
// 解析命令
command, args := parseCommand(message.Text)

// 处理命令
switch command {
case "start":
	return handleStartCommand(bot, message.Chat.ID)
case "help":
	return handleHelpCommand(bot, message.Chat.ID)
case "setrate":
	return handleSetRateCommand(bot, message.Chat.ID, args)
case "sendusdt":
	return handleSendUSDTCommand(bot, message.Chat.ID, args)
default:
	return handleUnknownCommand(bot, message.Chat.ID)
}
}

// handleCallbackQueryUpdate 处理回调查询更新
func handleCallbackQueryUpdate(callbackQuery *tgbotapi.CallbackQuery) error {
// 忽略查询 ID 不变的重复查询
if lastCallbackQueryID == callbackQuery.ID {
return nil
}
lastCallbackQueryID = callbackQuery.ID
// 解析查询数据
data, err := parseCallbackQueryData(callbackQuery.Data)
if err != nil {
	return err
}

// 处理查询
switch data.Action {
case "usdttx":
	return handleUSDTTransactionCallback(bot, callbackQuery, data)
default:
	return handleUnknownCallback(bot, callbackQuery)
}
}

// parseCommand 解析命令和参数
func parseCommand(text string) (string, string) {
parts := strings.SplitN(text, " ", 2)
if len(parts) == 1 {
return parts[0], ""
}
return parts[0], parts[1]
}

// parseCallbackQueryData 解析回调查询数据
func parseCallbackQueryData(data string) (*CallbackQueryData, error) {
var result CallbackQueryData
err := json.Unmarshal([]byte(data), &result)
if err != nil {
return nil, err
}
return &result, nil
}
// handleStartCommand 处理 start 命令
func handleStartCommand(bot *tgbotapi.BotAPI, chatID int64) error {
	// 发送欢迎消息
	msg := tgbotapi.NewMessage(chatID, "欢迎使用本机器人！")
	_, err := bot.Send(msg)
	if err != nil {
	return err
	}// 发送帮助消息
return handleHelpCommand(bot, chatID)
}

// handleHelpCommand 处理 help 命令
func handleHelpCommand(bot *tgbotapi.BotAPI, chatID int64) error {
// 发送帮助消息
msg := tgbotapi.NewMessage(chatID, "本机器人支持以下命令：\n"+
"/help - 显示帮助信息\n"+
"/setrate - 设置汇率\n"+
"/sendusdt - 发送 USDT")
_, err := bot.Send(msg)
if err != nil {
return err
}
return nil
}

// handleSetRateCommand 处理 setrate 命令
func handleSetRateCommand(bot *tgbotapi.BotAPI, chatID int64, args string) error {
// 解析参数
parts := strings.SplitN(args, " ", 2)
if len(parts) != 2 {
return errors.New("参数错误")
}
usdtAmount, err := strconv.ParseFloat(parts[0], 64)
if err != nil {
return errors.New("参数错误")
}
trxAmount, err := strconv.ParseFloat(parts[1], 64)
if err != nil {
return errors.New("参数错误")
}// 设置汇率
rateMutex.Lock()
rateMap[chatID] = Rate{usdtAmount, trxAmount}
rateMutex.Unlock()

// 发送响应消息
msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("已设置汇率：1 USDT = %.6f TRX", trxAmount/usdtAmount))
_, err = bot.Send(msg)
if err != nil {
	return err
}

return nil
}

// handleSendUSDTCommand 处理 sendusdt 命令
func handleSendUSDTCommand(bot *tgbotapi.BotAPI, chatID int64, args string) error {
// 解析参数
parts := strings.SplitN(args, " ", 2)
if len(parts) != 2 {
return errors.New("参数错误")
}
address := parts[0]
amount, err := strconv.ParseFloat(parts[1], 64)
if err != nil {
return errors.New("参数错误")
}
// 获取汇率
rateMutex.Lock()
rate, ok := rateMap[chatID]
rateMutex.Unlock()
if !ok {
	return errors.New("请先设置汇率")
}

// 计算 TRX 数量
trxAmount := amount * rate.TrxPerUsdt

// 创建 USDT 交易记录
txID := uuid.New().String()
usdtTx := USDTTransaction{txID, address, amount}

// 发送 USDT 交易请求
data, err := json.Marshal(&usdtTx)
if err != nil {
	return err
}
callbackData := CallbackQueryData{"usdttx", string(data)}
callbackMsg := tgbotapi.NewMessage(chatID, "请在钱包中确认 USDT 转账请求。")
callbackMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
	tgbotapi.NewInlineKeyboardButtonData("确认", callbackData.ToString()),
	tgbotapi.NewInlineKeyboardButtonData("取消", "cancel"),
	),
	)
	callbackMsg.ReplyMarkup = &replyMarkup
	_, err = bot.Send(callbackMsg)
	if err != nil {
	return err
	}
	// 保存 USDT 交易记录
usdtMutex.Lock()
usdtMap[txID] = usdtTx
usdtMutex.Unlock()

// 保存交易状态
txStatusMutex.Lock()
txStatusMap[txID] = TransactionStatus{chatID, false, trxAmount}
txStatusMutex.Unlock()

return nil
}

// handleCallbackQuery 处理回调查询
func handleCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
// 解析回调查询
callbackQuery := update.CallbackQuery
var callbackData CallbackQueryData
err := json.Unmarshal([]byte(callbackQuery.Data), &callbackData)
if err != nil {
return err
}
// 处理回调查询
switch callbackData.Type {
case "usdttx":
	return handleUSDTTransactionCallbackQuery(bot, callbackQuery.Message.Chat.ID, callbackData.Data, callbackQuery.ID)
case "trxtx":
	return handleTRXTransactionCallbackQuery(bot, callbackQuery.Message.Chat.ID, callbackData.Data, callbackQuery.ID)
}

return nil
}

// handleUSDTTransactionCallbackQuery 处理 USDT 交易回调查询
func handleUSDTTransactionCallbackQuery(bot *tgbotapi.BotAPI, chatID int64, data string, callbackQueryID string) error {
// 解析 USDT 交易记录
var usdtTx USDTTransaction
err := json.Unmarshal([]byte(data), &usdtTx)
if err != nil {
return err
}
// 发送 TRX 交易请求
rateMutex.Lock()
rate := rateMap[chatID]
rateMutex.Unlock()
trxAmount := usdtTx.Amount * rate.TrxPerUsdt
trxTx := TRXTransaction{uuid.New().String(), usdtTx.Address, trxAmount}
data, err = json.Marshal(&trxTx)
if err != nil {
	return err
}
callbackData := CallbackQueryData{"trxtx", string(data)}
callbackMsg := tgbotapi.NewMessage(chatID, "请在钱包中确认 TRX 转账请求。")
replyMarkup := tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("确认", callbackData.ToString()),
		tgbotapi.NewInlineKeyboardButtonData("取消", "cancel"),
	),
)
callbackMsg.ReplyMarkup = &replyMarkup
_, err = bot.Send(callbackMsg)
if err != nil {
	return err
}

// 更新交易状态
txStatusMutex.Lock()
txStatus := txStatusMap[usdtTx.ID]
txStatus.TrxAmount = trxAmount
txStatusMap[usdtTx.ID] = txStatus
txStatusMutex.Unlock()

// 返回成功响应
_, err = bot.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQueryID, "请在钱包中确认 USDT 转账请求。"))
return err
}

// handleTRXTransactionCallbackQuery 处理 TRX 交易回调查询
func handleTRXTransactionCallbackQuery(bot *tgbotapi.BotAPI, chatID int64, data string, callbackQueryID string) error {
	// 解析 TRX 交易记录
	var trxTx TRXTransaction
	err := json.Unmarshal([]byte(data), &trxTx)
	if err != nil {
	return err
	}
	// 发送 TRX 转账请求
trxClient := client.NewHTTPClient()
defer trxClient.Close()
var account *account.Account
var tx *builder.TransactionBuilder
for {
	// 获取账户信息
	account, err = trxClient.GetAccount(trxTx.FromAddress)
	if err != nil {
		return err
	}

	// 构造交易
	tx, err = builder.NewTransactionBuilderFromAccount(account)
	if err != nil {
		return err
	}
	tx = tx.
		AddTransfer(trxTx.ToAddress, asset.NewAmount(trxTx.Amount*1000000)).
		SetFeeLimit(10000000).
		SetMemo(trxTx.ID).
		SetExpiration(60 * 60 * 24)
	if tx.GetTransaction().Timestamp == 0 {
		tx = tx.SetTimestamp(time.Now().UnixNano() / int64(time.Millisecond))
	}

	// 签名交易
	signedTx, err := tx.Sign(account.PrivateKey)
	if err != nil {
		return err
	}

	// 发送交易
	broadcastResult, err := trxClient.BroadcastTransaction(signedTx)
	if err != nil {
		if broadcastResult != nil && strings.Contains(broadcastResult.Error, "Duplicated transaction") {
			// 如果交易已经存在，则重新生成交易 ID，并重试
			trxTx.ID = uuid.New().String()
			data, err = json.Marshal(&trxTx)
			if err != nil {
				return err
			}
			callbackData := CallbackQueryData{"trxtx", string(data)}
			callbackMsg := tgbotapi.NewMessage(chatID, "请在钱包中确认 TRX 转账请求。")
			replyMarkup := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("确认", callbackData.ToString()),
					tgbotapi.NewInlineKeyboardButtonData("取消", "cancel"),
				),
			)
			callbackMsg.ReplyMarkup = &replyMarkup
			_, err = bot.Send(callbackMsg)
			if err != nil {
				return err
			}
			continue
		}
		return err
	}
	break
}

// 更新交易状态
txStatusMutex.Lock()
txStatus := txStatusMap[trxTx.ID]
txStatus.Successful = true
txStatusMap[trxTx.ID] = txStatus
txStatusMutex.Unlock()

// 返回成功响应
_, err = bot.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQueryID, "TRX 转账成功！"))
return err
}

// handleCancelCallbackQuery 处理取消回调查询
func handleCancelCallbackQuery(bot *tgbotapi.BotAPI, callbackQueryID string) error {
_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQueryID, "操作已取消。"))
return err
}

// handleCommand 处理命令
func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
// 解析命令
command := update.Message.Command()
args := update.Message.CommandArguments()
// 处理命令
switch command {
case "setrate":
	return
// handleCommand 处理命令
func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	// 解析命令
	command := update.Message.Command()
	args := update.Message.CommandArguments()
	// 处理命令
switch command {
case "setrate":
	// 检查管理员权限
	if update.Message.Chat.ID != adminChatID {
		return errors.New("you are not authorized to perform this operation")
	}

	// 解析汇率
	rate, err := strconv.ParseFloat(args, 64)
	if err != nil {
		return err
	}

	// 更新汇率
	rateMutex.Lock()
	usdtTrxRate = rate
	rateMutex.Unlock()

	// 返回成功响应
	_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("汇率已设置为 %.6f TRX/USDT。", rate)))
	return err
case "getrate":
	// 查询汇率
	rateMutex.RLock()
	rate := usdtTrxRate
	rateMutex.RUnlock()

	// 返回汇率
	_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("当前汇率为 %.6f TRX/USDT。", rate)))
	return err
case "start":
	// 发送欢迎消息
	_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "欢迎使用 USDT 转账机器人！请发送 /help 获取帮助信息。"))
	return err
case "help":
	// 发送帮助消息
	helpMessage := "本机器人支持以下命令：\n\n"
	helpMessage += "/setrate <rate> - 设置 TRX/USDT 汇率（管理员专用）\n"
	helpMessage += "/getrate - 查询 TRX/USDT 汇率\n"
	helpMessage += "/deposit - 充值 USDT\n"
	helpMessage += "/withdraw <address> <amount> - 提现 USDT\n"
	helpMessage += "/balance - 查询余额\n"
	helpMessage += "\n发送 USDT 到本机器人的地址即可充值 USDT。"
	_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, helpMessage))
	return err
case "deposit":
	// 生成新的 USDT 地址并返回给用户
	address, err := generateUSDTAddress()
	if err != nil {
		return err
	}
	_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("请将 USDT 发送到地址 %s。", address)))
	return err
case "withdraw":
	// 解析提现地址和金额
	argsArray := strings.Split(args, " ")
	if len(argsArray) != 2 {
		return errors.New("invalid arguments")
	}
	toAddress := argsArray[0]
	amount, err := strconv.ParseFloat(argsArray[1], 64)
	if err != nil {
		return err
	}

	// 检查余额
	usdtBalance, err := getUSDTBalance(usdtAddress)
	if err != nil {
		return err
	}
	if amount > usdtBalance
	{
		return errors.New("insufficient balance")
		}
		
	// 计算 TRX 金额
	trxAmount := amount * usdtTrxRate

	// 调用 TRX 转账 API
	err = sendTRX(toAddress, trxAmount)
	if err != nil {
		return err
	}

	// 扣除 USDT 余额
	err = sendUSDT(usdtAddress, toAddress, amount)
	if err != nil {
		return err
	}

	// 返回成功响应
	_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("已向地址 %s 转账 %.2f TRX。", toAddress, trxAmount)))
	return err
case "balance":
	// 查询 USDT 余额
	balance, err := getUSDTBalance(usdtAddress)
	if err != nil {
		return err
	}
	_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("USDT 余额为 %.2f。", balance)))
	return err
default:
	// 未知命令，发送帮助消息
	helpMessage := "未知命令，请使用 /help 获取帮助信息。"
	_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, helpMessage))
	return err
}
}
func main() {
	// 初始化 Telegram Bot
	bot, err := tgbotapi.NewBotAPI("YOUR_TELEGRAM_BOT_TOKEN")
	if err != nil {
	log.Fatal(err)
	}
	// 设置 Webhook，当有消息到达时触发 handleMessage 函数处理
_, err = bot.SetWebhook(tgbotapi.NewWebhookWithCert("YOUR_WEBHOOK_URL", "YOUR_SSL_CERT_PATH"))
if err != nil {
	log.Fatal(err)
}

// 启动 HTTP 服务器，监听 Webhook 回调
http.HandleFunc("/"+bot.Token, func(w http.ResponseWriter, r *http.Request) {
	update, err := bot.HandleUpdate(r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	if update.Message == nil {
		return
	}

	err = handleMessage(bot, update)
	if err != nil {
		log.Println(err)
		return
	}
})

err = http.ListenAndServeTLS("YOUR_LISTEN_ADDRESS", "YOUR_SSL_CERT_PATH", "YOUR_SSL_KEY_PATH", nil)
if err != nil {
	log.Fatal(err)
}
}		
