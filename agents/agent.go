package agents

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/genai"
)

type ScreenshotAgent struct {
	agent          agent.Agent
	runner         *runner.Runner
	sessionService session.Service
}

func NewScreenshotAgent(ctx context.Context) (*ScreenshotAgent, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY not set")
	}

	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini model: %w", err)
	}

	instruction := `You are a screenshot organizer. 
	Analyze the image path or description and return ONLY a single word representing the category folder where it should be placed (e.g., "Code", "Finance", "Gaming", "Social", "Work").
	Do not provide explanations, just the category name. You MUST use the tool to see the folder names that already exist and see the current scenario.`

	a, err := llmagent.New(llmagent.Config{
		Name:        "Screenshot Organizer",
		Description: "Organizes screenshots into folders based on their content.",
		Model:       model,
		Instruction: instruction,
		Tools:       []tool.Tool{getReadDirectoriesNameTool()},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	sessionService := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:        "Screenshot_Organizer",
		Agent:          a,
		SessionService: sessionService,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	return &ScreenshotAgent{
		agent:          a,
		runner:         r,
		sessionService: sessionService,
	}, nil
}

func (s *ScreenshotAgent) Organize(ctx context.Context, filePath string) (string, error) {
	userID := "SYSTEM_USER"
	sessionID := "MAIN_SESSION"

	_, err := s.sessionService.Get(ctx, &session.GetRequest{
		AppName:   "Screenshot_Organizer",
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		_, err = s.sessionService.Create(ctx, &session.CreateRequest{
			AppName:   "Screenshot_Organizer",
			UserID:    userID,
			SessionID: sessionID,
		})
		if err != nil {
			return "", fmt.Errorf("failed to create session: %w", err)
		}
	}

	imgBytes, err := os.ReadFile(filePath)

	if err != nil {
		log.Fatal(err)
	}

	mimetype := "image/png"
	userMessage := &genai.Content{
		Role: genai.RoleUser,
		Parts: []*genai.Part{
			genai.NewPartFromText("Analyze this screenshot, present in " + filePath + " and categorize it."),
			genai.NewPartFromBytes(imgBytes, mimetype),
		},
	}

	var llmResponse string
	for event, err := range s.runner.Run(ctx, userID, sessionID, userMessage, agent.RunConfig{}) {
		if err != nil {
			log.Fatal(err)
		}

		if event.LLMResponse.Content != nil {
			for _, part := range event.LLMResponse.Content.Parts {
				if part.Text != "" {
					llmResponse += part.Text
				}
			}
		}
	}

	return llmResponse, nil
}

func getReadDirectoriesNameTool() tool.Tool {
	tool, err := functiontool.New(
		functiontool.Config{
			Name:        "read_directories_name_tool",
			Description: "Lists the names of existing folders in the screenshot directory to provide context for categorization.",
		},
		readDirectoriesNameTool,
	)

	if err != nil {
		log.Fatal(err)
	}

	return tool
}
