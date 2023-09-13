package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

const AUTH_TOKEN_ENV = "OPENAI_AUTH_TOKEN"

func getAuthToken() string {
	token, ok := os.LookupEnv(AUTH_TOKEN_ENV)
	if !ok {
		log.Fatalf("Please set %s", AUTH_TOKEN_ENV)
	}

	return token
}

func main() {
	client := openai.NewClient(getAuthToken())

	filename := filepath.Join(".", "history", time.Now().Format("2006-01-02-15-04-05")+".txt")

	// file to write history to
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	functions := []openai.FunctionDefinition{{
		Name:        "createTask",
		Description: "Create a new task",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"contents": {
					Type:        jsonschema.String,
					Description: "Contents of the task",
				},
			},
			Required: []string{"contents"},
		},
	}, {
		Name:        "deleteTask",
		Description: "Delete a task",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"filter": {
					Type:        jsonschema.String,
					Description: "Filter for the task",
				},
			},
			Required: []string{"filter"},
		},
	}, {
		Name:        "getTasks",
		Description: "Get a list of all tasks",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"filter": {
					Type:        jsonschema.String,
					Description: "Filter tasks by regex",
				},
			},
		},
	}}
	messages := []openai.ChatCompletionMessage{{
		Role:    openai.ChatMessageRoleSystem,
		Content: "you are a component of Tudu, a todo list app. please help the user manage their tasks.",
	}, {
		Role:    openai.ChatMessageRoleSystem,
		Content: "always prompt the user for input.",
	}, {
		Role: openai.ChatMessageRoleSystem,
		Content: "to create a task, use the createTask function, with the contents of the task as the argument. " +
			"the task will be created with the current timestamp as an ID. this ID will be returned to you." +
			"always verify the task list after creating tasks.",
	}, {
		Role: openai.ChatMessageRoleSystem,
		Content: "to edit: " +
			"1) getTask with a filter for the old task. " +
			"2) delete the old task exactly. " +
			"3) create the new task. " +
			"you may allow the user to do this in separate steps.",
	}, {
		Role: openai.ChatMessageRoleSystem,
	}, {
		Role:    openai.ChatMessageRoleSystem,
		Content: "always verify the task list before telling the user what tasks they have.",
	}, {
		Role:    openai.ChatMessageRoleSystem,
		Content: "please welcome the user to Tudu.",
	}}

	reader := bufio.NewReader(os.Stdin)

	ids := []int64{}
	tasks := map[int64]string{}

	for {
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:     openai.GPT3Dot5Turbo0613,
				Messages:  messages,
				Functions: functions,
			},
		)
		if err != nil {
			log.Printf("ChatCompletion error: %v", err)
			continue
		}

		msg := resp.Choices[0].Message
		messages = append(messages, msg)

		if msg.Content != "" {
			fmt.Println(msg.Content)

			// write history to file
			_, err = f.WriteString("gpt|" + msg.Content + "\n")
			if err != nil {
				panic(err)
			}
		}

		if msg.FunctionCall != nil {
			log.Println("Function call", msg.FunctionCall.Name, msg.FunctionCall.Arguments)

			// write history to file
			_, err = f.WriteString("fnc|" + msg.FunctionCall.Name + "|" + msg.FunctionCall.Arguments + "\n")
			if err != nil {
				panic(err)
			}

			switch msg.FunctionCall.Name {
			case "createTask":
				args := struct {
					Name string `json:"contents"`
				}{}

				err = json.Unmarshal([]byte(msg.FunctionCall.Arguments), &args)
				if err != nil {
					log.Printf("Error parsing arguments: %v", err)
					continue
				}

				now := time.Now()
				id := now.UnixNano()
				tasks[id] = args.Name
				ids = append(ids, id)

				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: fmt.Sprintf("task %q has been created.", args.Name),
				})
				continue
			case "deleteTask":
				args := struct {
					ID int64 `json:"id"`
				}{}

				err = json.Unmarshal([]byte(msg.FunctionCall.Arguments), &args)
				if err != nil {
					log.Printf("Error parsing arguments: %v", err)
					continue
				}

				for i, id := range ids {
					if id == args.ID {
						ids = append(ids[:i], ids[i+1:]...)
						break
					}
				}

				task := tasks[args.ID]
				delete(tasks, args.ID)

				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: fmt.Sprintf("task %d has been deleted. (%q)", args.ID, task),
				})
				continue
			case "getTasks":
				args := struct {
					Filter string `json:"filter"`
				}{}

				err = json.Unmarshal([]byte(msg.FunctionCall.Arguments), &args)
				if err != nil {
					log.Printf("Error parsing arguments: %v", err)
					continue
				}

				filter, err := regexp.Compile(args.Filter)
				if err != nil {
					log.Printf("Error compiling regex: %v", err)
					continue
				}

				content := "tasks:\n"

				if len(tasks) == 0 {
					content = "no tasks. "
				}

				for id, task := range tasks {
					if !filter.MatchString(task) {
						continue
					}

					content += fmt.Sprintf("id:%d contents:%q\n", id, task)
				}

				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: content,
				})
			default:
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: fmt.Sprintf("unknown function %q, args %q", msg.FunctionCall.Name, msg.FunctionCall.Arguments),
				})
			}
		}

		if messages[len(messages)-1].Role == openai.ChatMessageRoleSystem {
			continue
		}

		fmt.Printf("[%d] -> ", len(messages))

		now := time.Now()

		text, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("datetime: %s", now.Format(time.DateTime)),
		}, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: strings.TrimSpace(text),
		})
	}
}
