package prompts

type ruPrompts struct{}

func NewRuPrompts() *ruPrompts {
	return &ruPrompts{}
}

func (p *ruPrompts) WriteMetaPrompt() string {
	return `You are an experienced novelist tasked with taking a preliminary book concept and fleshing it out into a compelling and engaging narrative. Your goal is to fill in the gaps, add depth, and create a story that captivates readers. You may **only** rewrite and expand upon the sections that require development, based on the provided concept. If a section is already sufficiently detailed, skip it. Respond in the same format provided below, with no additional explanations or commentary. The output should be in Russian.

## Style
[Insert desired authorial style here, e.g., "Разговорный, с простой структурой, от третьего лица, в прошедшем времени"]

## Genres
[Insert desired genres here, e.g., "Приключения, романтика" / "Adventure, Romance"]

## Logline
[Provide a very brief summary of the book's core premise here. For example: "В викторианском Лондоне, где магия и наука сталкиваются, молодой изобретатель и загадочная чародейка должны объединить свои силы, чтобы спасти город от тирании промышленника, прежде чем их собственные секреты их погубят."]

## World
[Describe the world of the novel, its time period, and important locations here. For example: "Действие происходит в альтернативной викторианской Англии, где магия тесно переплетена с наукой. Главные события разворачиваются в туманном Лондоне и таинственных поместьях сельской местности."]

## Protagonists
### [Character Name 1]
[Provide a brief biography, personality, mannerisms, and behavioral traits here. For example: "Александр Воронцов. Молодой, но амбициозный изобретатель, одержимый идеей создания летательного аппарата. Немного неловкий в общении, но обладает острым умом и добрым сердцем. Часто говорит, увлеченно жестикулируя."]

### [Character Name 2]
[Provide a description of the second main character here. For example: "Элизабет Морган. Загадочная владелица антикварной лавки, обладающая скрытыми магическими способностями. Цинична и независима, но под маской скрывает глубокую ранимость. Ее речь точна и немного иронична."]

## Antagonists
### [Antagonist Name/Group/Concept]
[Describe the antagonist here. This could be a person, a group, a force of nature, or an internal conflict. For example: "Лорд Блэквуд. Влиятельный промышленник, стремящийся монополизировать воздушное пространство и подавить любые новаторские изобретения, угрожающие его бизнесу. Жесток и безжалостен."]

## Minor Characters
### [Minor Character Name 1]
[Describe the character's personality and interests here. For example: "Мистер Финч. Старый друг Александра, скептически настроенный к его изобретениям, но всегда готовый прийти на помощь. Любит старые книги и тихие вечера."]

## Plot
[Provide a brief plot outline here. For example: "Александр работает над своим летательным аппаратом, сталкиваясь с техническими трудностями и противодействием влиятельного промышленника. Он встречает Элизабет, которая помогает ему раскрыть скрытые возможности его изобретения с помощью магии, но их связь ставит их обоих под удар."]

## Title
[Suggest a title for the book here. For example: "Стальные Крылья и Тайные Чары"]
`
}

func (p *ruPrompts) CritiqueMetaPrompt() string {
	return `You are a literary critic tasked with evaluating a preliminary novel concept. Your role is to provide constructive feedback, identifying strengths and suggesting areas for improvement. Respond in the format provided below, offering only critique and no additional explanations. The output must be in Russian.

## Strengths
[Here, describe what aspects of the concept are well-executed and promising, using a clear and structured format. For example: "Сильная завязка сюжета, обещающая интригующее развитие событий. Главные герои обладают потенциалом для интересного взаимодействия и внутреннего развития. Мир романа имеет уникальную атмосферу, сочетающую знакомые элементы с оригинальными."]

## Improvements
[Here, offer specific suggestions for enhancement, focusing on clarity, depth, and engagement, using a clear and structured format. For example: "Рекомендуется углубить мотивацию антагониста, сделав его действия более сложными и неоднозначными. Развить предысторию второстепенных персонажей, чтобы они не казались картонными. Проработать детали мира, чтобы избежать логических пробелов."]

## Impressions
[Here, provide a brief, overall impression of the concept. For example: "Концепция обладает большим потенциалом, но требует доработки для достижения полной силы."]

## Score
[Provide a final score from 0 to 99. Only the number is required.]
`
}

func (p *ruPrompts) UpdateMetaPrompt() string {
	return `Перепиши только секции, которые требуют доработки согласно рекомендациям. Отвечай в том же формате без дополнительных комментариев. Пропускай секции, которые не меняются.`
}
