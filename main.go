package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// --- GLOBAL VARIABLES ---
// Keeping these global for now so we can access them anywhere.
// In a massive production app, we might inject these as dependencies, but for learning? This works.
var aiClient *genai.Client
var aiModel *genai.GenerativeModel

// --- STRUCTS (THE BLUEPRINTS) ---
// Go is statically typed, so we have to define the "shape" of our JSON data upfront.
// The text inside `...` tells Go how to map JSON keys to these struct fields.

// Module: Represents a single learning unit (like "Greetings" or "Numbers")
type Module struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Icon  string `json:"icon"`
	Done  bool   `json:"done"`
	Color string `json:"color"`
}

// Profile: The user's stats.
type Profile struct {
	DisplayName string  `json:"displayName"`
	Username    string  `json:"username"`
	Level       int     `json:"level"`
	Progress    float64 `json:"progress"`
}

// AppState: The big container for the dashboard data.
type AppState struct {
	StreakDays    int      `json:"streakDays"`
	XP            int      `json:"xp"`
	CurrentLesson string   `json:"currentLesson"`
	Modules       []Module `json:"modules"` // A slice (array) of modules
	Profile       Profile  `json:"profile"` // Nested struct
}

// ChatRequest: What the frontend sends us when talking to the bot.
type ChatRequest struct {
	Context   string `json:"context"`   // The story situation (Sari asking name, etc.)
	UserQuery string `json:"userQuery"` // What the user actually typed
}

// ChatResponse: What we send back to the frontend.
type ChatResponse struct {
	Reply string `json:"reply"`
}

// Option: Represents a choice button in the UI.
type Option struct {
	Text      string `json:"text"`
	Reaction  string `json:"reaction"`
	IsCorrect bool   `json:"isCorrect"`
	// 'omitempty' is cool: if this string is empty, it won't even appear in the JSON. Saves bandwidth!
	ReactionImage string `json:"reactionImage,omitempty"`
}

// Scene: One screen in our visual novel.
type Scene struct {
	InputType     string   `json:"inputType"` // Determines if we show buttons ('choice') or a text box ('text')
	CharacterName string   `json:"characterName"`
	CharacterMood string   `json:"characterMood"` // Now stores the filename, e.g., "Muka_sari_senang.png"
	Dialogue      string   `json:"dialogue"`
	Options       []Option `json:"options"`
}

// Lesson: A collection of scenes.
type Lesson struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Scenes []Scene `json:"scenes"`
}

// TextAnswerRequest: When the user types "Rendang", this catches it.
type TextAnswerRequest struct {
	SceneIndex int    `json:"sceneIndex"`
	Answer     string `json:"answer"`
}

// TextAnswerResponse: The result of our logic check (Rendang logic, etc.)
type TextAnswerResponse struct {
	IsCorrect     bool   `json:"isCorrect"`
	Reaction      string `json:"reaction"`
	Feedback      string `json:"feedback"`
	ReactionImage string `json:"reactionImage,omitempty"` // Adding reaction image from the server
}

// --- MOCK DATA ---
// Usually, this comes from a Database (PostgreSQL/MySQL).
// But since I'm just learning backend logic, storing it in memory variables is way faster to build.

var state = AppState{
	StreakDays:    5,
	XP:            1200,
	CurrentLesson: "Greetings 1 - Jakarta",
	Modules: []Module{
		{ID: "1", Title: "Greetings 1 - Jakarta", Icon: "👋", Done: false, Color: "#06b6d4"},
		{ID: "2", Title: "Numbers - Bali", Icon: "🔢", Done: true, Color: "#f59e0b"},
	},
	Profile: Profile{
		DisplayName: "Wira",
		Username:    "@DeeDotz",
		Level:       12,
		Progress:    0.62,
	},
}

var lesson1 = Lesson{
	ID:    "1",
	Title: "Meeting Sari in Jakarta",
	Scenes: []Scene{
		{
			InputType:     "choice",
			CharacterName: "Sari",
			CharacterMood: "Muka_sari_menyapa.png", // Using PNGs now instead of Emojis
			Dialogue:      "Halo! Selamat siang. Aku belum pernah melihatmu di sini. Siapa namamu?",
			Options: []Option{
				{Text: "Nama saya Wira.", Reaction: "Wah, nama yang bagus! Salam kenal, Aditya.", IsCorrect: true, ReactionImage: "Muka_sari_senang.png"},
				{Text: "Saya umur 12 tahun.", Reaction: "Eh? Aku tanya nama lho, bukan umur.", IsCorrect: false, ReactionImage: "Muka_sari_bingung.png"},
			},
		},
		{
			InputType:     "text",
			CharacterName: "Sari",
			CharacterMood: "Muka_sari_senang.png",
			Dialogue:      "Salam kenal ya. Ngomong-ngomong, kamu asalnya dari mana?",
		},
		{
			InputType:     "choice",
			CharacterName: "Sari",
			CharacterMood: "Muka_sari_kaget.png",
			Dialogue:      "Terus, sedang apa kamu di Jakarta?",
			Options: []Option{
				{Text: "Saya sedang liburan.", Reaction: "Wah asyik! Jakarta punya banyak mall lho.", IsCorrect: true, ReactionImage: "Muka_sari_senang.png"},
				// This reaction is pure gold lol
				{Text: "Saya sedang makan batu.", Reaction: "Hah?! Gigi kamu kuat banget... (先生：石を食べていますか？ほんまに？ふざけんなぁお前)", IsCorrect: false, ReactionImage: "Muka_sari_bingung.png"},
				{Text: "Saya sedang belajar Bahasa.", Reaction: "Wah rajin sekali! Semangat ya.", IsCorrect: true, ReactionImage: "Muka_sari_senang.png"},
			},
		},
		{
			InputType:     "text",
			CharacterName: "Sari",
			CharacterMood: "Muka_sari_senang.png",
			Dialogue:      "Kelihatan masih muda ya. Boleh tahu umur kamu berapa? (Tulis angka saja)",
		},
		{
			InputType:     "choice",
			CharacterName: "Sari",
			CharacterMood: "Muka_sari_menyap.png",
			Dialogue:      "Oke deh, aku harus pergi kerja dulu. Senang mengobrol denganmu!",
			Options: []Option{
				{Text: "Sampai jumpa!", Reaction: "Dah! Hati-hati di jalan.", IsCorrect: true, ReactionImage: "Muka_sari_happy.png"},
				{Text: "Kamu siapa?", Reaction: "Lho? Kan tadi kita baru kenalan...", IsCorrect: false, ReactionImage: "Muka_sari_bingung.png"},
			},
		},
	},
}

// --- [UPDATE] LOGIC GEMINI AI ---

// initGemini: Fires up the connection to Google's servers.
func initGemini(apiKey string) {
	ctx := context.Background() // Go uses 'Context' to handle timeouts and request cancellations.
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err) // If we can't connect to AI, the app crashes (Fatal).
	}
	aiClient = client

	// HEADS UP: I noticed I wrote "gemini-2.5-flash" here.
	// That model doesn't exist yet! It will crash with a 404.
	// I need to change this to "gemini-1.5-flash" or "gemini-pro" later to make it work.
	aiModel = client.GenerativeModel("gemini-2.5-flash")
}

// callRealGemini: This is where we send the prompt to the AI.
func callRealGemini(contextInfo string, userQuery string) string {
	ctx := context.Background()

	// System Prompt Engineering:
	// This is crucial. I'm telling the AI to act like a "Sensei" for Japanese speakers learning Indonesian.
	// I'm mixing ID/JP/EN instructions to make it robust.
	prompt := fmt.Sprintf(`
    Kamu adalah 'Sensei', asisten guru bahasa Indonesia yang ramah.
    Kamu tahu sangat sulit untuk belajar bahasa indonesia jika kamu adalah orang jepang, coba posisikan diri kamu
    sebagai orang jepang yang punya pengalaman mengajar bahasa indonesia.
    
    [KONTEKS CERITA SAAT INI]
    %s

    [PERTANYAAN USER]
    %s

    [INSTRUKSI]
    Jawablah pertanyaan user dengan singkat, jelas, dan ramah.
    Jika user bertanya soal bahasa/grammar, jelaskan alasannya.
    Jangan menjawab terlalu panjang (maksimal 2-3 kalimat). Kamu ditargetkan untuk orang yang belajar Bahasa Indonesia menggunakan Bahasa Jepang, jadi gunakanlah Bahasa Jepang sebagai main.
    Tetapi kadang, jelaskan juga menggunakan bahasa indonesia dan bahasa inggris. Tergantung pertanyaan mereka dalam bahasa apa duluan.
    `, contextInfo, userQuery)

	resp, err := aiModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		// If the AI errors out (like invalid key or model not found), we log it to the terminal
		// and send a fallback message to the UI so the app doesn't freeze.
		log.Println("Error Gemini:", err)
		return "Maaf, koneksi otak saya sedang terputus. Coba lagi nanti ya!"
	}

	// Parsing the complicated JSON response from Google to get just the text we need.
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])
	}

	return "Hmm, saya tidak tahu harus jawab apa."
}

// --- LOGIC TEXT EVALUATION ---
// I'm keeping this logic HARDCODED.
// Why? Because forcing specific funny responses (like "Rendang") is easier with if/else
// than convincing an AI to be funny in a specific way. It's faster too.
func evaluateTextAnswer(sceneIndex int, answer string) TextAnswerResponse {
	ans := strings.ToLower(answer) // Normalize input so "Jepang" and "jepang" match.

	// SCENE 2: Asal (Where are you from?)
	if sceneIndex == 1 {
		if strings.Contains(ans, "jepang") || strings.Contains(ans, "japan") {
			return TextAnswerResponse{IsCorrect: true, Reaction: "Wah, Jepang! 🇯🇵 Aku suka anime lho.", Feedback: "Sempurna! (完璧です！)", ReactionImage: "Muka_sari_kaget.png"}
		}
		if strings.Contains(ans, "indonesia") {
			return TextAnswerResponse{IsCorrect: true, Reaction: "Oh, orang lokal ternyata!", Feedback: "Benar! (正解です)", ReactionImage: "Muka_sari_senang.png"}
		}
		// THE FUNNY PART: Logic for "Rendang"
		if strings.Contains(ans, "rendang") {
			return TextAnswerResponse{IsCorrect: false, Reaction: "Eh rendang? Daerah mana itu... 🍛 Itu nama makanan kali!", Feedback: "Salah konteks. Rendang itu makanan wkwk. (それは食べ物です。場所ではありません ww)", ReactionImage: "Muka_sari_bingung.png"}
		}
		if strings.Contains(ans, "sabun") {
			return TextAnswerResponse{IsCorrect: false, Reaction: "...kamu gak papa? Kepalamu terbentur?", Feedback: "Jawaban sangat aneh. (変な答えです。大丈夫ですか？)", ReactionImage: "Muka_sari_kaget.png"}
		}
		// Default catch-all for wrong answers
		return TextAnswerResponse{IsCorrect: false, Reaction: "Hmm, aku belum pernah dengar nama daerah itu.", Feedback: "Coba jawab dengan nama negara atau kota. (国や都市の名前で答えてみてください)", ReactionImage: "Muka_sari_bingung.png"}
	}
	// SCENE 4: Umur (How old?)
	if sceneIndex == 3 {
		// Basic validation: just check if there's a digit.
		if strings.ContainsAny(ans, "0123456789") {
			return TextAnswerResponse{IsCorrect: true, Reaction: "Ooh segitu. Masih semangat muda ya!", Feedback: "Angka diterima. (数字が確認できました。OKです)", ReactionImage: "Muka_sari_happy.png"}
		}
		return TextAnswerResponse{IsCorrect: false, Reaction: "Itu bukan angka deh kayaknya...", Feedback: "Tulis menggunakan angka. (数字を使って書いてください)", ReactionImage: "Muka_sari_bingung.png"}
	}
	return TextAnswerResponse{IsCorrect: false, Reaction: "...", Feedback: "Error logic"}
}

func main() {
	// 1. INIT GEMINI
	//
	// SECURITY WARNING:
	// I'm hardcoding the API Key here for testing/learning purposes.
	// In a real job/production app, NEVER do this.
	// Hackers scan GitHub for keys. Use environment variables (os.Getenv) instead.
	//
	apiKey := os.Getenv(("GeMINI_API_KEY"))

	// Just a sanity check to remind myself to replace the placeholder cause i tend to forgot
	if apiKey == "" {
		fmt.Println("!WARNING: Have not set the API Key yet, check README")
	}

	initGemini(apiKey)

	// 'NewServeMux' is essentially the router. It decides which function runs
	// based on the URL the browser requests.
	mux := http.NewServeMux()

	// API Endpoint: Returns the dashboard stats (XP, streak, etc.)
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state)
	})

	// API Endpoint: Returns the lesson content (Sari's dialogue)
	mux.HandleFunc("/api/lesson/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(lesson1)
	})

	// API Endpoint: The AI Chatbot
	mux.HandleFunc("/api/ask-tutor", func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		// Decoding JSON from the request body (what the frontend sent us)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Call the AI logic
		aiReply := callRealGemini(req.Context, req.UserQuery)

		resp := ChatResponse{Reply: aiReply}
		w.Header().Set("Content-Type", "application/json")
		// Encoding our struct back to JSON to send to the browser
		json.NewEncoder(w).Encode(resp)
	})

	// API Endpoint: Validating text input (Rendang logic)
	mux.HandleFunc("/api/check-text", func(w http.ResponseWriter, r *http.Request) {
		var req TextAnswerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp := evaluateTextAnswer(req.SceneIndex, req.Answer)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Static File Server: This serves the HTML, CSS, JS, and Images
	// It looks inside the "web" folder.
	fs := http.FileServer(http.Dir("web"))
	mux.Handle("/", fs)

	// Start the server on port 8080. This loop runs forever until I Ctrl+C.
	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
