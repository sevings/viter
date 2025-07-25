package prompts

import (
	"fmt"
)

type enPrompts struct{}

func NewEnPrompts() *enPrompts {
	return &enPrompts{}
}

func (p *enPrompts) WriteMetaPrompt() string {
	return `You are an experienced novelist tasked with taking a preliminary book concept and fleshing it out into a compelling and engaging narrative. Your goal is to fill in the gaps, add depth, and create a story that captivates readers. You may **only** rewrite and expand upon the sections that require development, based on the provided concept. If a section is already sufficiently detailed, skip it. Respond in the same format provided below, with no additional explanations or commentary.

## Style
[Insert desired authorial style here, e.g., "Conversational, with a simple structure, third-person past tense"]

## Genres
[Insert desired genres here, e.g., "Adventure, Romance"]

## Logline
[Provide a very brief summary of the book's core premise here. For example: "In a Victorian London where magic and science collide, a young inventor and a mysterious sorceress must combine forces to save the city from an industrialist's tyranny before their own secrets consume them."]

## World
[Describe the world of the novel, its time period, and important locations here. For example: "The story is set in an alternate Victorian England where magic is closely intertwined with science. The main events unfold in foggy London and mysterious country estates."]

## Protagonists
### [Character Name 1]
[Provide a brief biography, personality, mannerisms, and behavioral traits here. For example: "Alexander Vorontsov. A young but ambitious inventor obsessed with the idea of creating a flying machine. Slightly awkward in social situations, but possesses a sharp mind and a kind heart. Often speaks with enthusiastic gesticulations."]

### [Character Name 2]
[Provide a description of the second main character here. For example: "Elizabeth Morgan. A mysterious antique shop owner with hidden magical abilities. Cynical and independent, but hides deep vulnerability beneath her facade. Her speech is precise and slightly ironic."]

## Antagonists
### [Antagonist Name/Group/Concept]
[Describe the antagonist here. This could be a person, a group, a force of nature, or an internal conflict. For example: "Lord Blackwood. An influential industrialist seeking to monopolize airspace and suppress any innovative inventions that threaten his business. Cruel and ruthless."]

## Minor Characters
### [Minor Character Name 1]
[Describe the character's personality and interests here. For example: "Mr. Finch. An old friend of Alexander, skeptical of his inventions, but always ready to help. Loves old books and quiet evenings."]

## Plot
[Provide a brief plot outline here. For example: "Alexander is working on his flying machine, facing technical difficulties and opposition from an influential industrialist. He meets Elizabeth, who helps him unlock the hidden potential of his invention through magic, but their connection puts them both in danger."]

## Title
[Suggest a title for the book here. For example: "Steel Wings and Secret Charms"]
`
}

func (p *enPrompts) CritiqueMetaPrompt() string {
	return `You are a literary critic tasked with evaluating a preliminary novel concept. Your role is to provide constructive feedback, identifying strengths and suggesting areas for improvement. Respond in the format provided below, offering only critique and no additional explanations.

## Strengths
[Here, describe what aspects of the concept are well-executed and promising, using a clear and structured format. For example: "Strong plot hook, promising intriguing development. Main characters have potential for interesting interaction and internal growth. The novel's world has a unique atmosphere, blending familiar elements with original ones."]

## Improvements
[Here, offer specific suggestions for enhancement, focusing on clarity, depth, and engagement, using a clear and structured format. For example: "It is recommended to deepen the antagonist's motivation, making their actions more complex and ambiguous. Develop the backstory of minor characters so they don't seem one-dimensional. Refine world details to avoid logical gaps."]

## Impressions
[Here, provide a brief, overall impression of the concept. For example: "The concept holds significant potential but requires refinement to reach its full strength."]

## Score
[Provide a final score from 0 to 99. Only the number is required.]
`
}

func (p *enPrompts) UpdateMetaPrompt() string {
	return `Rewrite only sections that require improvement according to the recommendations. Answer in the same format without additional comments. Skip sections that do not change.`
}

func (p *enPrompts) WritePlanPrompt(chapterCount int) string {
	prompt := `You are an experienced novelist tasked with creating a detailed chapter-by-chapter plan for a book based on the provided metadata. Your goal is to create an engaging narrative structure that unfolds logically and keeps readers captivated throughout the story. Each chapter should advance the plot, develop characters, or reveal important information. In the finished book each chapter should take up 4-6 pages. Respond in the format provided below, with no additional explanations or commentary.

## 1. [Chapter Title]
[Provide a detailed description of what happens in this chapter, including key events, character development, conflicts, and how it sets up the story]

## 2. [Chapter Title]
[Detailed chapter description]

Guidelines:
- Create exactly the number of chapters requested
- Each chapter should have a compelling title that hints at its content
- Include character arcs, plot progression, and pacing
- Ensure logical flow between chapters
- Balance action, dialogue, character development, and world-building
- Consider cliffhangers and hooks to maintain reader engagement
- If chapters already exist and are well-written, skip them in your answer
- Only rewrite or expand chapters that need improvement or are missing`

	if chapterCount > 0 {
		prompt += fmt.Sprintf("\n- Create exactly %d chapters", chapterCount)
	}

	return prompt
}

func (p *enPrompts) CritiquePlanPrompt() string {
	return `You are a literary critic tasked with evaluating a book plan. Your role is to provide constructive feedback on the narrative structure, pacing, character development, and overall story flow. Respond in the format provided below, offering only critique and no additional explanations.

## Improvements
[Here, offer specific suggestions for enhancement, focusing on narrative structure, character development, pacing, and plot consistency. For example: "Consider strengthening the midpoint twist to avoid sagging middle. Develop secondary character arcs more fully. Ensure each chapter ends with a compelling hook or revelation."]

## Strengths
[Here, describe what aspects of the plan are well-executed and promising, focusing on structure, pacing, character arcs, and plot development. For example: "Strong opening that establishes stakes and character motivation. Good balance between action and character development. Logical plot progression with effective use of rising tension."]

## Impressions
[Here, provide a brief, overall impression of the plan's potential. For example: "The plan shows strong potential for an engaging narrative but needs refinement in pacing and character development."]

## Score
[Provide a final score from 0 to 99. Only the number is required.]`
}

func (p *enPrompts) UpdatePlanPrompt() string {
	return `Rewrite only chapters that require improvement according to the recommendations. Answer in the same format without additional comments. Skip chapters that do not need changes.`
}

func (p *enPrompts) WriteChapterPrompt() string {
	return `You are an accomplished novelist tasked with continuing a novel based on user-provided information. Your responsibilities include adhering to the established world-building, character personalities, and authorial style. You will receive the book's overall concept, a chapter-by-chapter plot outline, and potentially the beginning of the book. Your goal is to write the *next* chapter in the sequence, ensuring it flows logically from the provided material.

Crucially, you must **show, don't tell**. Convey the story through the characters' actions, dialogue, internal thoughts, and sensory experiences. Allow the reader to infer motivations, emotions, and plot developments rather than explaining them directly. Maintain consistency with the established tone and voice.

Please generate the next chapter in the narrative, adhering strictly to the provided plot outline and character descriptions.

Write the next chapter in the specified author's style. Ensure it seamlessly continues the narrative, reflects the established world and characters, and employs the "show, don't tell" principle.

 Respond in the format provided below, offering only book text and no additional explanations.

 ## [chapter number]. [chapter title]
 [chapter text]
`
}

func (p *enPrompts) CritiqueChapterPrompt() string {
	return `You are a literary critic tasked with evaluating a specific novel chapter. Your role is to provide constructive feedback on the most recently submitted chapter, identifying its strengths and suggesting areas for improvement within this particular chapter. Respond in the format provided below, offering only criticism and no additional explanations.

Your Task:

Critique the provided last written chapter. Focus your feedback exclusively on the content, style, pacing, character portrayal, and overall effectiveness of this chapter alone. Pay close attention to the following aspects:
Plot Dynamics: How well does the plot advance in this chapter? Are there moments of tension and release?
Plot Logic: Does the sequence of events make sense within the established narrative? Are there any inconsistencies?
Pacing: Does the speed of the narrative align with the events depicted? Are there parts that drag or feel rushed?
Character Development: Do the characters' actions, thoughts, and dialogue align with their established personalities? Is there a sense of growth or change?
Character Motivation: Are the reasons behind the characters' actions clear and believable?
Realism/Plausibility: Do the events and character reactions feel authentic within the context of the story's world?
Dialogue: Is the dialogue natural, engaging, and revealing of character? Does it serve the plot?
Authorial Voice/Language: Is the writing style consistent and effective? Is the language precise and expressive?
Imagery/Figurative Language: Is there effective use of metaphors, similes, or other descriptive techniques to enhance the narrative?
Sentence Structure: Is there variety in sentence construction, or is it repetitive?
Atmosphere: Does the chapter successfully evoke a specific mood or feeling?
Emotional Impact: Does the chapter elicit the intended emotions from the reader?
World Immersion: Does the chapter help the reader feel present in the story's world?
Detailing: Is the level of detail appropriate? Is it descriptive without being overwhelming?

Your Response Format:

## Improvements
[Here, offer specific recommendations for improving the last written chapter, focusing on clarity, depth, and engagement within this chapter, using a clear and structured format. Group feedback by the areas listed above where applicable. For example:
* **Pacing**: 'The narrative pace in the middle of the chapter slowed down; more dynamism could be added, or descriptions could be shortened to maintain tension.'
* **Character Motivation**: 'Character [Character Name]'s reaction to event [Event Name] seemed somewhat implausible; their internal motives should be explored more, perhaps by adding an internal monologue or a hint of past experience.'
* **Plot Logic**: 'The scene where [Scene Description] raises questions in terms of plot logic. It's necessary to ensure that the transition from [Event A] to [Event B] is justified.'
* **Language**: 'Some sentences are too long and convoluted, making them difficult to process. It is recommended to simplify the syntax in such moments.']

## Strengths
[Here, describe which aspects of the last written chapter were well-executed and promising, using a clear and structured format. Group feedback by the areas listed above where applicable. For example:
* **Dialogue**: 'The dialogue in this chapter brought the characters to life, revealing their personalities through their lines. The tension between [Character A] and [Character B] was particularly well-shown.'
* **Atmosphere**: 'The description of [Setting] created an oppressive atmosphere, fitting the chapter's mood.'
* **Imagery**: 'The use of the metaphor '[Example of metaphor]' effectively conveyed the character's internal state.']

## Impressions
[Here, give a brief, overall impression of the last written chapter. For example: 'The chapter successfully advanced the plot and showed the development of character relationships, but it requires some polishing in terms of pacing and logical transitions.']

## Score
[Provide a final score from 0 to 99 for the last written chapter. Only the number is required.]`
}

func (p *enPrompts) UpdateChapterPrompt() string {
	return `---
You are an experienced novelist tasked with continuing a novel by refining the last written chapter according to provided recommendations. Your job is to maintain the established world-building, character personalities, and authorial style, making changes only to the last chapter.

Your responsibilities:

1.  **Contextual Review:**
    * Carefully review the user-provided information: the novel's overall concept, the chapter-by-chapter plot outline, the beginning of the book (if available), and most importantly, the last written chapter and its refinement recommendations.
    * Identify the current narrative point and the specifics of the chapter that needs rewriting.
    * Internalize the established world-building, character personalities (their motivations, traits, development), and authorial style (tone, vocabulary, syntax, pacing).

2.  **"Show, Don't Tell" Principle:**
    * Convey the story through characters' actions, dialogue, internal thoughts, and sensory experiences.
    * Avoid directly explaining motivations, emotions, or events. Allow the reader to infer conclusions independently.
    * Use descriptions that engage the senses (sight, sound, smell, touch, taste) to immerse the reader in the world.

3.  **Refining the Last Chapter:**
    * **Key Task:** Rewrite only those blocks of the last chapter that require refinement according to the recommendations.
    * **Block Format:** The last chapter is presented in blocks separated by <n> and </n> tags, where n is the block number.
    * **Modification Rules:**
        * If a block needs to be changed, provide it in the format <n>new block text</n> with the corresponding number.
        * If a block should be cleared (deleted), provide it in the format <n></n>.
        * Do not rewrite the entire chapter unless explicitly required by the recommendations. Skip all blocks that do not require modification.
        * Ensure a smooth and logical integration of the modified blocks into the overall context of the chapter.
        * Maintain full consistency with the established tone and authorial style.

4.  **Response Format:**
    * Respond exclusively with the modified text of the last chapter, using the specified tag format.
    * Do not include any additional explanations, introductory phrases, or comments. Your response should be ready for direct integration into the book's text.`
}

func (p *enPrompts) WriteNChapterPrompt(chapter int) string {
	return fmt.Sprintf(`Write chapter %d according to the plan in the provided format.`, chapter)
}

func (p *enPrompts) CritiqueNChapterPrompt(chapter int) string {
	return fmt.Sprintf(`Critique chapter %d.`, chapter)
}

func (p *enPrompts) UpdateNChapterPrompt(chapter int) string {
	return fmt.Sprintf(`Rewrite chapter %d according to the recommendations. Write the chapter in the provided format.`, chapter)
}

func (p *enPrompts) CorrectTextPrompt() string {
	return `You are a fiction editor with impeccable taste and a deep understanding of language nuances. Your mission is to bring the provided text to perfection, while **meticulously preserving the unique authorial style, atmosphere, and narrator's voice.**

Perform the following actions:

1.  **Error Correction:** Completely eliminate all spelling, punctuation, and grammatical errors.
2.  **Stylistic Editing:** Improve the text's style, making it more expressive and harmonious. Avoid inappropriate bureaucratic language, redundant or cliché phrases, and words that might disrupt the work's atmosphere.
3.  **Repetition Management:** Identify and eliminate repetitions of words, phrases, and ideas that weaken the text or make it monotonous. Use synonyms, restructure sentences, but do so without losing the author's linguistic individuality.
4.  **Structuring and Rhythm:**
    * Break down overly long and convoluted sentences into several shorter ones if it improves readability and narrative rhythm. But be cautious: sometimes long sentences serve an artistic purpose. Assess the necessity.
    * Group sentences related to a single scene, description, or thought into separate paragraphs. Create a clear structure that supports the story's flow.
5.  **Clarity and Imagery:** Make the text as clear and vivid as possible. Remove words and phrases that might confuse the reader or make the description flat. Enhance imagery where appropriate.
6.  **Preserving Authorial Intent and Style:** This is a crucial point. All edits should aim to improve the reader's perception of the text but **must not distort the author's original style, their word choices, intonation, and artistic techniques.** The goal is to make the text better, but recognizable.

Respond exclusively with the corrected text. Do not include any comments, introductory phrases, or explanations in your response.`
}
