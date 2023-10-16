package main

import (
	"context"
	_ "embed"
	"errors"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	ctx := context.Background()
	env, err := newEnv()
	if err != nil {
		handleError(err)
	}
	config := openai.DefaultAzureConfig(env.azureOpenAiKey, env.azureOpenAiEndpoint)
	client := openai.NewClientWithConfig(config)

	prePrompt, err := env.getPrePrompt()
	if err != nil {
		handleError(err)
	}

	diff, err := env.getDiff()
	if err != nil {
		handleError(err)
	}

	prompts := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: prePrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: diff,
		},
	}

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    env.azureOpenAIModelDeployName,
			Messages: prompts,
		},
	)
	if err != nil {
		handleError(err)
	}

	message := ""
	for _, choice := range resp.Choices {
		message += choice.Message.Content
	}

	os.Stdout.WriteString(message)
}

func handleError(err error) {
	panic(err)
}

type env struct {
	azureOpenAiKey             string
	azureOpenAiEndpoint        string
	azureOpenAIModelDeployName string
	prePromptPath              string
	gitDiff                    string
	gitDiffPath                string
}

var (
	errAzureOpenAIKeyRequired      = errors.New("AZURE_OPEN_AI_KEY is required")
	errAzureOpenAIEndpointRequired = errors.New("AZURE_OPEN_AI_ENDPOINT is required")
	errAzureOpenAIModelDeployName  = errors.New("AZURE_OPEN_AI_MODEL_DEPLOY_NAME is required")
	errGitDiffRequired             = errors.New("GIT_DIFF or GIT_DIFF_PATH is required")
	errGitDiffContentRequired      = errors.New("GIT_DIFF content is required")

	//go:embed pre_prompt.md
	PrePrompt string
)

func newEnv() (*env, error) {
	e := &env{
		azureOpenAiKey:             os.Getenv("AZURE_OPEN_AI_KEY"),
		azureOpenAiEndpoint:        os.Getenv("AZURE_OPEN_AI_ENDPOINT"),
		azureOpenAIModelDeployName: os.Getenv("AZURE_OPEN_AI_MODEL_DEPLOY_NAME"),
		prePromptPath:              os.Getenv("PRE_PROMPT_PATH"),
		gitDiff:                    os.Getenv("GIT_DIFF"),
		gitDiffPath:                os.Getenv("GIT_DIFF_PATH"),
	}
	if err := e.validate(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *env) validate() error {
	if e.azureOpenAiKey == "" {
		return errAzureOpenAIKeyRequired
	}
	if e.azureOpenAiEndpoint == "" {
		return errAzureOpenAIEndpointRequired
	}
	if e.azureOpenAIModelDeployName == "" {
		return errAzureOpenAIModelDeployName
	}
	if e.gitDiffPath == "" && e.gitDiff == "" {
		return errGitDiffRequired
	}
	return nil
}

func (e *env) getPrePrompt() (string, error) {
	path := e.prePromptPath
	if path != "" {
		s, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(s), nil
	}
	return PrePrompt, nil
}

func (e *env) getDiff() (string, error) {
	if e.gitDiff != "" {
		return e.gitDiff, nil
	}
	s, err := os.ReadFile(e.gitDiffPath)
	if err != nil {
		return "", err
	}
	content := string(s)
	if content == "" {
		return "", errGitDiffContentRequired
	}
	return content, nil
}
