package utils

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

func PromptString(scanner *bufio.Scanner, message string, defaultValue string) string {
	var prompt string
	if defaultValue != "" {
		prompt = fmt.Sprintf("%s [%s]: ", message, defaultValue)
	} else {
		prompt = fmt.Sprintf("%s: ", message)
	}

	fmt.Print(prompt)
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input != "" {
			return input
		}
	}
	return defaultValue
}

func PromptInt(scanner *bufio.Scanner, message string, defaultValue int) int {
	defaultStr := strconv.Itoa(defaultValue)
	inputStr := PromptString(scanner, message, defaultStr)

	value, err := strconv.Atoi(inputStr)
	if err != nil {
		fmt.Printf("Error: Invalid value. Using default: %d\n", defaultValue)
		return defaultValue
	}
	return value
}
