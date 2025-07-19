package prompts

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
