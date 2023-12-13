package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// User 结构表示用户信息
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// DiaryBook 结构表示日记本
type DiaryBook struct {
	Entries []DiaryEntry `json:"entries"`
}

// DiaryEntry 结构表示日记条目
type DiaryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Content   string    `json:"content"`
}

// Block 结构表示区块链中的块
type Block struct {
	Index        int       `json:"index"`
	Timestamp    time.Time `json:"timestamp"`
	DiaryEntries DiaryBook `json:"diaryEntries"`
	PrevHash     string    `json:"prevHash"`
	Hash         string    `json:"hash"`
	Nonce        int       `json:"nonce"`
}

// Blockchain 结构表示区块链
type Blockchain struct {
	Blocks []Block `json:"blocks"`
}

var (
	currentUser User
	blockchain  Blockchain
)

const (
	passwdFile       = "./passwd"
	diaryFolder      = "./diaries/"     // 存放日记文件的文件夹
	blockchainFolder = "./blockchains/" // 存放区块链文件的文件夹
)

// 登录状态
var loggedIn bool

// 日记是否被篡改的标志
var diaryTampered bool

// 在main函数之前添加一个初始化函数
func initialize() {
	// 创建日记文件夹
	err := os.MkdirAll(diaryFolder, 0755)
	if err != nil {
		fmt.Println("Error creating diary folder:", err)
		os.Exit(1)
	}

	// 创建区块链文件夹
	err = os.MkdirAll(blockchainFolder, 0755)
	if err != nil {
		fmt.Println("Error creating blockchain folder:", err)
		os.Exit(1)
	}
}

var mutex sync.Mutex

func main() {

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/loginin", logininHandler)
	http.HandleFunc("/validateBlockchain", validateBlockchainHandler)

	fmt.Println("后端服务器启动在82端口")
	http.ListenAndServe("0.0.0.0:82", nil)
}

// 处理验证区块链的请求
func validateBlockchainHandler(w http.ResponseWriter, r *http.Request) {
	// 打印接收到的请求信息
	//log.Printf("Received request: %s %s\n", r.Method, r.URL)

	// 鉴别用户是否登录
	if !loggedIn {
		http.Error(w, "用户未登录", http.StatusUnauthorized)
		return
	}

	// 调用验证区块链是否被篡改的函数
	if isBlockchainTampered() {
		// 如果区块链被篡改，返回警告信息
		responseData := map[string]string{
			"status":  "warning",
			"message": "警告：区块链数据已被篡改！",
		}

		// 将结构体转换为 JSON 格式的字符串
		responseJSON, err := json.Marshal(responseData)
		if err != nil {
			http.Error(w, "无法序列化响应数据", http.StatusInternalServerError)
			return
		}

		// 打印要发送到前端的响应信息
		//log.Printf("Sent response (length=%d): %q\n", len(responseJSON), responseJSON)
		// 设置响应头为 application/json
		w.Header().Set("Content-Type", "application/json")
		// 发送到前端
		w.Write(responseJSON)
	} else {
		// 如果区块链未被篡改，返回成功信息
		responseData := map[string]string{
			"status":  "success",
			"message": "区块链未被篡改",
		}

		// 将结构体转换为 JSON 格式的字符串
		responseJSON, err := json.Marshal(responseData)
		if err != nil {
			http.Error(w, "无法序列化响应数据", http.StatusInternalServerError)
			return
		}

		// 打印要发送到前端的响应信息
		//log.Printf("Sent response (length=%d): %q\n", len(responseJSON), responseJSON)
		// 设置响应头为 application/json
		w.Header().Set("Content-Type", "application/json")
		// 发送到前端
		w.Write(responseJSON)
	}
}

// 登陆后的请求处理函数
func logininHandler(w http.ResponseWriter, r *http.Request) {
	// 鉴别用户是否登录
	if !loggedIn {
		http.Error(w, "用户未登录", http.StatusUnauthorized)
		return
	}

	// 打印接收到的请求信息
	//log.Printf("Received request: %s %s\n", r.Method, r.URL)

	// 定义一个结构体，用于存储要返回给前端的数据
	type Response struct {
		Status         string `json:"status"`
		BlockchainInfo string `json:"blockchainInfo"`
		DiaryInfo      string `json:"diaryInfo"`
		Refresh        string `json:"refresh"`
	}

	// 处理获取日记和区块链信息请求
	if r.Method == http.MethodGet {
		// 获取区块链信息的字符串
		blockchainInfo := viewBlockchainInfo(currentUser.Username)

		// 获取日记信息的字符串
		diaryInfo := viewDiaryEntries(currentUser.Username)

		// 构建要返回的结构体
		responseData := Response{
			Status:         "success",
			BlockchainInfo: blockchainInfo,
			DiaryInfo:      diaryInfo,
			Refresh:        "True",
		}

		// 将结构体转换为 JSON 格式的字符串
		responseJSON, err := json.Marshal(responseData)
		if err != nil {
			http.Error(w, "无法序列化响应数据", http.StatusInternalServerError)
			return
		}

		// 打印要发送到前端的响应信息
		//log.Printf("Sent response (length=%d): %q\n", len(responseJSON), responseJSON)
		// 设置响应头为 application/json
		w.Header().Set("Content-Type", "application/json")
		// 发送到前端
		w.Write(responseJSON)

		// 处理写日记请求
	} else if r.Method == http.MethodPost {
		var entry struct {
			Content string `json:"content"`
		}

		err := json.NewDecoder(r.Body).Decode(&entry)
		if err != nil {
			http.Error(w, "无效的请求数据", http.StatusBadRequest)
			return
		}

		// 提交日记条目
		submitDiaryEntry(entry.Content, w)

		// 获取更新后的区块链信息的字符串
		blockchainInfo := viewBlockchainInfo(currentUser.Username)

		// 获取更新后的日记信息的字符串
		diaryInfo := viewDiaryEntries(currentUser.Username)

		// 构建要返回的结构体
		responseData := Response{
			Status:         "success",
			BlockchainInfo: blockchainInfo,
			DiaryInfo:      diaryInfo,
			Refresh:        "True",
		}

		// 将结构体转换为 JSON 格式的字符串
		responseJSON, err := json.Marshal(responseData)
		if err != nil {
			http.Error(w, "无法序列化响应数据", http.StatusInternalServerError)
			return
		}

		// 打印要发送到前端的响应信息
		//log.Printf("Sent response (length=%d): %q\n", len(responseJSON), responseJSON)
		// 设置响应头为 application/json
		w.Header().Set("Content-Type", "application/json")
		// 发送到前端
		w.Write(responseJSON)
	} else {
		http.Error(w, "不允许的请求方法", http.StatusMethodNotAllowed)
	}
}

// 登出处理函数
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// 处理登出请求
	if r.Method == http.MethodPost {
		// 登出成功，将登录状态保存在服务器端
		loggedIn = false
		currentUser = User{}

		// 返回前端 JSON 数据
		responseData := map[string]string{
			"status":  "success",
			"message": "登出成功",
		}

		jsonResponse, err := json.Marshal(responseData)
		if err != nil {
			http.Error(w, "无法生成 JSON 响应", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)

		//fmt.Println("登出成功")
		//fmt.Println("后端返回给前端的响应：", string(jsonResponse))
	} else {
		http.Error(w, "不允许的请求方法", http.StatusMethodNotAllowed)
	}
}

// 登录请求处理函数
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// 处理登录请求
	if r.Method == http.MethodPost || r.Method == http.MethodGet {
		var credentials User
		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			http.Error(w, "无效的请求数据", http.StatusBadRequest)
			return
		}

		username := credentials.Username
		password := credentials.Password

		if validateLogin(username, password) {
			// 登录成功，将登录状态保存在服务器端
			loggedIn = true
			currentUser = User{Username: username, Password: password}

			// 加载区块链和日记信息
			loadBlockchain(username)

			// 添加CORS头部
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:80")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// 返回前端 JSON 数据
			responseData := map[string]string{
				"status":  "success",
				"message": fmt.Sprintf("登录成功，用户名：%s，密码：%s", username, password),
			}

			jsonResponse, err := json.Marshal(responseData)
			if err != nil {
				http.Error(w, "无法生成 JSON 响应", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)

			// 便于调试，将发送数据打印在终端
			//fmt.Printf("登录成功，用户名：%s，密码：%s\n", username, password)
			//fmt.Println("后端返回给前端的响应：", string(jsonResponse))
		} else {
			// 返回前端 JSON 数据
			responseData := map[string]string{
				"status": "error",
				"error":  "用户名或密码错误",
			}

			jsonResponse, err := json.Marshal(responseData)
			if err != nil {
				http.Error(w, "无法生成 JSON 响应", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)

			//fmt.Println("登录失败")
			//fmt.Println("后端返回给前端的响应：", string(jsonResponse))
		}
	} else {
		http.Error(w, "不允许的请求方法", http.StatusMethodNotAllowed)
	}
}

// 加载区块链信息和日记
func loadBlockchain(username string) {
	// 从文件中加载区块链信息
	blockchainFileName := getBlockchainFileName(username)
	blockchainData, err := ioutil.ReadFile(blockchainFileName)
	if err != nil {
		// 如果文件不存在，创建一个新文件
		if os.IsNotExist(err) {
			createGenesisBlock(username)
			return
		}
		fmt.Println("无法加载区块链文件:", err)
		return
	}

	err = json.Unmarshal(blockchainData, &blockchain)
	if err != nil {
		fmt.Println("无法解析区块链文件:", err)
		return
	}

	// 从文件中加载日记信息
	diaryFileName := getDiaryFileName(username)
	diaryData, err := ioutil.ReadFile(diaryFileName)
	if err != nil {
		fmt.Println("无法加载日记文件:", err)
		return
	}

	err = json.Unmarshal(diaryData, &getCurrentBlock().DiaryEntries)
	if err != nil {
		fmt.Println("无法解析日记文件:", err)
		return
	}
}

// 处理注册请求
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var credentials User
		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			http.Error(w, "无效的请求数据", http.StatusBadRequest)
			return
		}

		// 打印出注册所用的用户信息
		fmt.Printf("Username: %s, Password: %s\n", credentials.Username, credentials.Password)

		mutex.Lock()
		defer mutex.Unlock()

		if userExists(credentials.Username) {
			http.Error(w, "用户名已存在，请选择登录或使用其他用户名", http.StatusBadRequest)
			return
		}

		err = addUser(credentials.Username, credentials.Password)
		if err != nil {
			http.Error(w, "注册失败", http.StatusInternalServerError)
			return
		}

		// 注册成功响应
		responseData := map[string]interface{}{
			"status":  "success",
			"message": "注册成功，正在自动登录...",
		}

		// 调用初始化函数
		initialize()

		// 自动登录的逻辑
		// 设置登录状态
		loggedIn = true
		currentUser = User{Username: credentials.Username, Password: credentials.Password}

		// 创建创世区块并写入文件
		createGenesisBlock(credentials.Username)

		// 将响应数据转换为 JSON
		responseJSON, err := json.Marshal(responseData)
		if err != nil {
			http.Error(w, "无法序列化响应数据", http.StatusInternalServerError)
			return
		}

		// 打印后端响应
		//fmt.Printf("Backend Response: %s\n", string(responseJSON))

		// 将响应发送到前端
		w.Write(responseJSON)
	} else {
		http.Error(w, "请求方法无效", http.StatusMethodNotAllowed)
	}
}

// 查看区块链信息的函数，返回包含区块链信息的字符串
func viewBlockchainInfo(username string) string {
	var blockchainInfo strings.Builder

	for _, block := range blockchain.Blocks {
		// 检查区块的用户名是否与当前登录用户一致
		if block.DiaryEntries.Entries != nil && block.DiaryEntries.Entries[0].Content == fmt.Sprintf("%s的日记本", username) {
			blockchainInfo.WriteString(fmt.Sprintf("Index: %d\n", block.Index))
			blockchainInfo.WriteString(fmt.Sprintf("Timestamp: %s\n", block.Timestamp.Format("2006-01-02 15:04:05")))
			blockchainInfo.WriteString(fmt.Sprintf("Hash: %s\n", block.Hash))
			blockchainInfo.WriteString(fmt.Sprintf("PrevHash: %s\n", block.PrevHash))
			blockchainInfo.WriteString(fmt.Sprintf("Nonce: %d\n", block.Nonce))
			blockchainInfo.WriteString("Diary Entries:\n")

			for _, entry := range block.DiaryEntries.Entries {
				blockchainInfo.WriteString(fmt.Sprintf("[%s] %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"), entry.Content))
			}

			blockchainInfo.WriteString("--------------------------------------------------------------------------------------------------------------------\n")
		}
	}

	// 返回包含区块链信息的字符串
	return blockchainInfo.String()
}

func validateLogin(username, password string) bool {
	// 从密码文件中验证用户
	users, err := loadUsers()
	if err != nil {
		fmt.Println("无法验证登录:", err)
		return false
	}

	for _, user := range users {
		if user.Username == username && user.Password == password {
			currentUser = user
			return true
		}
	}

	return false
}

func userExists(username string) bool {
	// 检查用户名是否存在于密码文件中
	users, err := loadUsers()
	if err != nil {
		fmt.Println("无法检查用户名:", err)
		return false
	}

	for _, user := range users {
		if user.Username == username {
			return true
		}
	}

	return false
}

func addUser(username, password string) error {
	// 添加新用户到密码文件
	users, err := loadUsers()
	if err != nil {
		return err
	}

	newUser := User{Username: username, Password: password}
	users = append(users, newUser)

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(passwdFile, data, 0755)
	if err != nil {
		return err
	}

	return nil
}

// 修改 submitDiaryEntry 函数，以在写入区块链后进行验证
func submitDiaryEntry(content string, w http.ResponseWriter) {
	// 创建日记条目
	entry := DiaryEntry{
		Timestamp: time.Now(),
		Content:   content,
	}

	// 记录挖矿开始时间
	startTime := time.Now()

	// 挖矿创建新区块
	newBlock := mineBlock(entry)

	// 计算挖矿耗时
	elapsedTime := time.Since(startTime)

	// 添加到区块链
	blockchain.Blocks = append(blockchain.Blocks, *newBlock)

	// 更新区块链
	updateBlockchain()

	// 验证是否被篡改
	checkDiaryIntegrity(w)

	// 在写日记结束后返回添加成功以及挖矿信息
	response := fmt.Sprintf("日记添加成功！挖矿耗时：%s\n", elapsedTime)
	response += fmt.Sprintf("新区块信息：Index: %d, Timestamp: %s, Hash: %s, Nonce: %d\n",
		newBlock.Index, newBlock.Timestamp.Format("2006-01-02 15:04:05"), newBlock.Hash, newBlock.Nonce)

	w.Write([]byte(response))
}

// 验证日记的完整性
func checkDiaryIntegrity(w http.ResponseWriter) {
	if isBlockchainTampered() {
		w.Write([]byte("警告：区块链数据已被篡改！"))
		diaryTampered = true
	} else {
		w.Write([]byte("区块链未被篡改"))
	}
}

// 验证区块链是否被篡改
func isBlockchainTampered() bool {
	for i := 1; i < len(blockchain.Blocks); i++ {
		prevBlock := blockchain.Blocks[i-1]
		currentBlock := blockchain.Blocks[i]

		// 检查当前区块的前一个哈希是否等于前一个区块的哈希
		if currentBlock.PrevHash != prevBlock.Hash {
			return true
		}
	}
	return false
}

// 查看所有日记
func viewDiaryEntries(username string) string {
	seenEntries := make(map[string]struct{})
	var Diaryinfo strings.Builder

	for _, block := range blockchain.Blocks {
		// 检查区块的用户名是否与当前登录用户一致
		if block.DiaryEntries.Entries != nil && block.DiaryEntries.Entries[0].Content == fmt.Sprintf("%s的日记本", username) {
			for _, entry := range block.DiaryEntries.Entries {
				// 去重逻辑
				entryKey := entry.Timestamp.Format("2006-01-02 15:04:05") + entry.Content
				if _, seen := seenEntries[entryKey]; !seen {
					Diaryinfo.WriteString(fmt.Sprintf("[%s] %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"), entry.Content))
					seenEntries[entryKey] = struct{}{}
				}
			}
		}
	}

	// 返回包含日记信息的字符串
	return Diaryinfo.String()
}

func mineBlock(entry DiaryEntry) *Block {
	// 挖矿
	currentBlock := getCurrentBlock()
	newBlock := mine(currentBlock, entry)

	// 更新区块链
	updateBlockchain()

	return newBlock
}

func mine(prevBlock *Block, entry DiaryEntry) *Block {
	// 工作量证明（增加挖矿次数）
	maxAttempts := 1000000
	for nonce := 0; nonce < maxAttempts; nonce++ {
		hash := calculateHash(*prevBlock, nonce)
		if hashIsValid(hash) {
			// 将原有的日记条目也包含在 DiaryEntries 中
			newDiaryEntries := append(prevBlock.DiaryEntries.Entries, entry)
			return &Block{
				Index:        prevBlock.Index + 1,
				Timestamp:    time.Now(),
				DiaryEntries: DiaryBook{Entries: newDiaryEntries},
				PrevHash:     prevBlock.Hash,
				Hash:         hash,
				Nonce:        nonce,
			}
		}
	}
	// 如果未能找到符合条件的哈希值，可以考虑返回错误或者进行其他处理
	return nil
}

func hashIsValid(hash string) bool {
	// 简化版的验证条件
	return strings.HasPrefix(hash, "000")
}

func calculateDiaryHash(diary DiaryBook) string {
	data, err := json.Marshal(diary)
	if err != nil {
		fmt.Println("计算哈希值时发生错误:", err)
		return ""
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

func calculateHash(block Block, nonce int) string {
	data := fmt.Sprintf("%d%d%s%d%s", block.Index, block.Timestamp.UnixNano(), block.PrevHash, nonce, calculateDiaryHash(block.DiaryEntries))
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func getCurrentBlock() *Block {
	// 获取当前区块
	return &blockchain.Blocks[len(blockchain.Blocks)-1]
}

func updateBlockchain() {
	// 将区块链和日记写入文件
	blockchainData, err := json.MarshalIndent(blockchain, "", "  ")
	if err != nil {
		fmt.Println("更新区块链时发生错误:", err)
		return
	}

	err = ioutil.WriteFile(getBlockchainFileName(currentUser.Username), blockchainData, 0755)
	if err != nil {
		fmt.Println("写入区块链文件时发生错误:", err)
		return
	}

	diaryData, err := json.MarshalIndent(getCurrentBlock().DiaryEntries, "", "  ")
	if err != nil {
		fmt.Println("更新日记时发生错误:", err)
		return
	}

	err = ioutil.WriteFile(getDiaryFileName(currentUser.Username), diaryData, 0755)
	if err != nil {
		fmt.Println("写日记时发生错误:", err)
		return
	}
}

func loadUsers() ([]User, error) {
	// 从文件中读取用户信息
	data, err := ioutil.ReadFile(passwdFile)
	if err != nil {
		// 如果文件不存在，创建一个新文件
		if os.IsNotExist(err) {
			err := ioutil.WriteFile(passwdFile, []byte{}, 0755)
			if err != nil {
				fmt.Println("Error creating password file:", err)
				os.Exit(1)
			}
			return []User{}, nil
		}
		return nil, err
	}

	// 如果文件为空，返回空用户列表
	if len(data) == 0 {
		return []User{}, nil
	}

	var users []User
	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// 创世区块
func createGenesisBlock(username string) Block {
	// 检查文件是否已存在，如果存在则不再创建
	blockchainFileName := getBlockchainFileName(username)
	diaryFileName := getDiaryFileName(username)

	if _, err := os.Stat(blockchainFileName); err == nil {
		fmt.Println("区块链文件已存在，无需再次创建。")
		return Block{}
	}

	if _, err := os.Stat(diaryFileName); err == nil {
		fmt.Println("日记文件已存在，无需再次创建。")
		return Block{}
	}

	genesisBlock := mineGenesisBlock(username)
	writeBlockToFile(genesisBlock, username)

	// 将创世区块添加到全局的 blockchain 变量中
	blockchain.Blocks = append(blockchain.Blocks, genesisBlock)

	return genesisBlock
}

func mineGenesisBlock(username string) Block {
	// 挖矿创造创世区块
	nonce := 0
	for {
		genesisBlock := Block{
			Index:        0,
			Timestamp:    time.Now(),
			DiaryEntries: DiaryBook{Entries: []DiaryEntry{{Timestamp: time.Now(), Content: fmt.Sprintf("%s的日记本", username)}}},
			PrevHash:     "",
			Nonce:        nonce,
		}

		// 计算哈希值
		hash := calculateHash(genesisBlock, nonce)

		// 检查哈希值是否符合条件
		if hashIsValid(hash) {
			// 符合条件的哈希值找到，返回创世区块
			genesisBlock.Hash = hash
			return genesisBlock
		}

		// 尝试下一个 nonce
		nonce++
	}
}

// 写入区块和日记到文件的函数
func writeBlockToFile(block Block, username string) {
	blockchainData, err := json.MarshalIndent(Blockchain{Blocks: []Block{block}}, "", "  ")
	if err != nil {
		fmt.Println("写入区块到文件时发生错误:", err)
		return
	}

	err = ioutil.WriteFile(getBlockchainFileName(username), blockchainData, 0644)
	if err != nil {
		fmt.Println("写入区块链文件时发生错误:", err)
		return
	}

	diaryData, err := json.MarshalIndent(block.DiaryEntries, "", "  ")
	if err != nil {
		fmt.Println("写入日记到文件时发生错误:", err)
		return
	}

	err = ioutil.WriteFile(getDiaryFileName(username), diaryData, 0644)
	if err != nil {
		fmt.Println("写入日记文件时发生错误:", err)
		return
	}
}

func getDiaryFileName(username string) string {
	return diaryFolder + username + "_diary.json"
}

func getBlockchainFileName(username string) string {
	return blockchainFolder + username + "_blockchain.json"
}
