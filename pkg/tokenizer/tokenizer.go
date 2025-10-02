package tokenizer

import (
	"unicode"
)

// EstimateTokens provides a rough estimate of token count for text
// This is a simplified estimation based on word count and punctuation
// For accurate counts, integrate with tiktoken or similar libraries
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}

	// Rough approximation:
	// - Count words
	// - Add punctuation/special characters
	// - Average: ~1.3 tokens per word for English text

	words := 0
	specialChars := 0
	inWord := false

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			if !inWord {
				words++
				inWord = true
			}
		} else {
			inWord = false
			if unicode.IsPunct(r) || unicode.IsSymbol(r) {
				specialChars++
			}
		}
	}

	// Rough formula: words * 1.3 + special chars * 0.5
	// This approximates GPT-style tokenization
	tokens := int(float64(words)*1.3 + float64(specialChars)*0.5)

	// Minimum 1 token for non-empty text
	if tokens == 0 && text != "" {
		tokens = 1
	}

	return tokens
}

// EstimateTokensSimple provides a very simple token estimation
// Rule of thumb: ~4 characters per token
func EstimateTokensSimple(text string) int {
	if text == "" {
		return 0
	}

	charCount := len([]rune(text))
	tokens := charCount / 4

	if tokens == 0 && text != "" {
		tokens = 1
	}

	return tokens
}

// CountTokensForMessages estimates tokens for chat messages including structure overhead
// OpenAI/GPT models add overhead for message formatting
func CountTokensForMessages(systemPrompt, userMessage, context string) int {
	// Base tokens for message structure
	structureOverhead := 4 // <|im_start|>, <|im_end|>, role tags, etc.

	systemTokens := 0
	if systemPrompt != "" {
		systemTokens = EstimateTokens(systemPrompt) + structureOverhead
	}

	userTokens := EstimateTokens(userMessage) + structureOverhead
	contextTokens := EstimateTokens(context)

	return systemTokens + userTokens + contextTokens
}
