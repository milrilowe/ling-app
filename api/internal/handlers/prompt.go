package handlers

import (
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RandomPromptResponse struct {
	Prompt string `json:"prompt"`
}

var prompts = []string{
	"What did you do today?",
	"Tell me about your hometown.",
	"What's your favorite food?",
	"Describe your perfect weekend.",
	"What's the last book you read?",
	"Tell me about your family.",
	"What do you like to do in your free time?",
	"What's your dream job?",
	"Where would you like to travel?",
	"Tell me about a memorable experience.",
}

func GetRandomPrompt(c *gin.Context) {
	prompt := prompts[rand.Intn(len(prompts))]

	c.JSON(http.StatusOK, RandomPromptResponse{
		Prompt: prompt,
	})
}
