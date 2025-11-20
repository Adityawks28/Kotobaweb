🇮🇩 KotobaWeb - AI Powered Interactive Language Learning

KotobaWeb is a proof-of-concept Language Learning Application that combines Visual Novel storytelling and game with AI.

In the future I hope anyone can find an easy way to learn my mother language, Indonesian Language or Bahasa Indonesia.
Especially but not limited to people residing in Japan.

Unlike standard quizzes, this app uses Google Gemini AI to provide context-aware feedback. It acts as a "Sensei" that explains why an answer is wrong based on the current story context, and can differentiate between a grammar mistake and a context mistake (e.g., answering a food name when asked for a location).

Tech Stack
- Backend: Go (Golang) - `net/http` standard library.
- AI Engine: Google Gemini 2.5 Flash.
- Frontend: Vanilla JavaScript, TailwindCSS (No heavy frameworks, purely logic-driven).
- Data Structure: Custom Go Structs for Scene Management.

Key Features
1.  Dynamic Visual Novel Engine: Backend-driven scene flow (sprites, dialogues, choices).
2.  Hybrid Input System: Supports both multiple-choice and free-text inputs in one seamless flow.
3.  Context-Aware AI Feedback: The AI knows the specific scene context (e.g., "Sari is asking for your name") and evaluates user input based on that.
4.  Smart Asset Management: Dynamically switches between Emoji avatars and PNG Sprites based on character emotions.

Future Roadmap (Planned Features)
I am actively developing this project. Upcoming features include:
- [ ] PostgreSQL Integration: To save user progress and persistent user profiles.
- [ ] Voice Input: Allowing users to speak Indonesian directly to the AI.
- [ ] Authentication: Secure login using JWT.
- [ ] Scenario Builder: A tool to create new stories without coding.

How to Run locally

1. Clone the repository
   bash
   git clone [https://github.com/USERNAME_KAMU/kotobaweb.git](https://github.com/USERNAME_KAMU/kotobaweb.git)
   cd kotobaweb
2. Get your Gemini API Key (Google AI Studio)
3. Run Server
   Bash
   # Mac/Linux
   GEMINI_API_KEY="your_api_key_here" go run main.go

   # Windows (PowerShell)
   $env:GEMINI_API_KEY="your_api_key_here"; go run main.go
4. Open local host
5. Add photos and other assets
6. Have fun
