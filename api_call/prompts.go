package api_call

import "fmt"

func GetSummaryPrompt(existingSummary, newChats string) string {
	return fmt.Sprintf(`
Progressively summarize the lines of conversation provided, adding onto the previous summary returning a new summary.

Existing Summary:
%s

New Conversation:
%s

Instructions:
1. Analyze the existing summary and the new conversation.
2. Identify key points, topics, and any decisions or conclusions from the new conversation.
3. Integrate this new information with the existing summary.
4. Ensure the updated summary remains concise (aim for 3-5 sentences) while capturing all essential information.
5. Maintain a chronological flow of events and topics where appropriate.
6. If the new conversation introduces entirely new topics, add them to the summary while preserving the most crucial points from the existing summary.
7. Remove any redundant or obsolete information from the existing summary.
8. Use neutral language and avoid personal opinions or interpretations.

Provide only the updated summary as your output, without any additional text, headings, or explanations.

Updated Summary:
`, existingSummary, newChats)
}
