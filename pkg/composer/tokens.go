package composer

import "github.com/pluqqy/pluqqy-cli/pkg/utils"

// EstimateTokens estimates the token count for a given text
func EstimateTokens(text string) int {
	return utils.EstimateTokens(text)
}