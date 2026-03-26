# Screenshot Organizer Agent (Work in Progress)

An intelligent AI-powered screenshot organizer. This project monitors your screenshot folder in real-time and uses the multimodal power of **Gemini** to analyze the visual content of each image and categorize it automatically.

---

## Features

- **Real-Time Monitoring:** Instantly detects new screenshots in the configured folder.
- **Multimodal Analysis:** Sends the image bytes directly to the Gemini API to "see" the content.
- **Automatic Categorization:** Suggests categories such as *Code*, *Finance*, *Gaming*, *Social*, *Work*, etc.
- **Agent-Based Architecture:** Built with the **Google ADK (Agent Development Kit).**

## Technologies Used

- **Language:** [Go](https://go.dev/)
- **AI:** [Google Gemini 2.5 Flash](https://ai.google.dev/)
- **Agent Framework:** [Google ADK](https://github.com/google/adk)

## Installation and Configuration

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-user/screenshot_organizer.git
    cd screenshot_organizer
    ```

2.  **Configure the API Key:**
    Create a `.env` file in the project root:
    ```env
    GOOGLE_API_KEY=your_key_here
    ```

3.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

4.  **Run the project:**
    ```bash
    go run cmd/main.go
    ```

## How It Works?

1.  The program starts monitoring your Screenshots folder.
2.  Whenever a new screenshot is detected, the `ScreenshotAgent` reads the image bytes.
3.  The image is sent to Gemini with a system instruction for visual analysis.
4.  The agent returns the ideal category based on what it "saw" in the image.

---
Developed by Yuri Filgueira.
