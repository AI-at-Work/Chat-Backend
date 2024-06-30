package api_call

import "fmt"

func GetSummaryPrompt(existingSummary, chats string) string {
	return fmt.Sprintf(
		"You are given an existing summary of a user's conversion and a new conversation. "+
			"Your task is to integrate the new conversion into existing summary, coherent summary that combines both the existing and new information. "+
			"The goal is to create a concise and updated summary. Note: Only give the summary do not include any headings; give text output only."+
			"\n\n**Existing Summary:**\n"+
			"%s\n"+
			"\n\n**New Conversation:**\n"+
			"%s\n\n", existingSummary, chats)
}
