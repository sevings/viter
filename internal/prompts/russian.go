package prompts

import (
	"fmt"
)

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

func (p *ruPrompts) WritePlanPrompt(chapterCount int) string {
	prompt := `Вы - опытный романист, которому поручено создать подробный план книги по главам на основе предоставленных метаданных. Ваша цель - создать увлекательную структуру повествования, которая разворачивается логично и держит читателей в напряжении на протяжении всей истории. Каждая глава должна продвигать сюжет, развивать персонажей или раскрывать важную информацию. Отвечайте в формате, приведенном ниже, без дополнительных объяснений или комментариев.

## 1. [Название главы]
[Предоставьте подробное описание того, что происходит в этой главе, включая ключевые события, развитие персонажей, конфликты и то, как она готовит почву для истории]

## 2. [Название главы]
[Подробное описание главы]

Рекомендации:
- Каждая глава должна иметь захватывающее название, которое намекает на её содержание
- Включите арки персонажей, развитие сюжета и темп
- Обеспечьте логический переход между главами
- Сбалансируйте действие, диалоги, развитие персонажей и построение мира
- Рассмотрите возможность использования клиффхэнгеров и зацепок для поддержания интереса читателей
- Если главы уже существуют и хорошо написаны, пропустите их в своем ответе
- Переписывайте или расширяйте только те главы, которые нуждаются в улучшении или отсутствуют`

	if chapterCount > 0 {
		prompt += fmt.Sprintf("\n- Создайте точно %d глав", chapterCount)
	}

	return prompt
}

func (p *ruPrompts) CritiquePlanPrompt() string {
	return `Вы - литературный критик, которому поручено оценить план книги. Ваша роль - предоставить конструктивную обратную связь по структуре повествования, темпу, развитию персонажей и общему ходу истории. Отвечайте в формате, приведенном ниже, предлагая только критику без дополнительных объяснений.

## Strengths
[Здесь опишите, какие аспекты плана хорошо проработаны и перспективны, сосредоточившись на структуре, темпе, арках персонажей и развитии сюжета. Например: "Сильное начало, которое устанавливает ставки и мотивацию персонажа. Хороший баланс между действием и развитием персонажа. Логичное развитие сюжета с эффективным использованием нарастающего напряжения."]

## Improvements
[Здесь предложите конкретные рекомендации по улучшению, сосредоточившись на структуре повествования, развитии персонажей, темпе и последовательности сюжета. Например: "Рассмотрите возможность усиления поворота в середине, чтобы избежать провисания в середине истории. Более полно развивайте арки второстепенных персонажей. Убедитесь, что каждая глава заканчивается захватывающей зацепкой или откровением."]

## Impressions
[Здесь дайте краткое общее впечатление о потенциале плана. Например: "План показывает сильный потенциал для увлекательного повествования, но нуждается в доработке темпа и развития персонажей."]

## Score
[Предоставьте финальную оценку от 0 до 99. Требуется только число.]`
}

func (p *ruPrompts) UpdatePlanPrompt() string {
	return `Перепишите только главы, которые требуют улучшения согласно рекомендациям. Отвечайте в том же формате без дополнительных комментариев. Пропускайте главы, которые не нуждаются в изменениях.`
}

func (p *ruPrompts) WriteChapterPrompt() string {
	return `You are an accomplished novelist tasked with continuing a novel based on user-provided information. Your responsibilities include adhering to the established world-building, character personalities, and authorial style. You will receive the book's overall concept, a chapter-by-chapter plot outline, and potentially the beginning of the book. Your goal is to write the *next* chapter in the sequence, ensuring it flows logically from the provided material.

Crucially, you must **show, don't tell**. Convey the story through the characters' actions, dialogue, internal thoughts, and sensory experiences. Allow the reader to infer motivations, emotions, and plot developments rather than explaining them directly. Maintain consistency with the established tone and voice.

Please generate the next chapter in the narrative, adhering strictly to the provided plot outline and character descriptions.

Write the next chapter in the specified author's style. Ensure it seamlessly continues the narrative, reflects the established world and characters, and employs the "show, don't tell" principle.

 Respond in the format provided below, offering only book text and no additional explanations. The output must be in Russian.

 ## [chapter number]. [chapter title]
 [chapter text]
`
}

func (p *ruPrompts) CritiqueChapterPrompt() string {
	return `You are a literary critic tasked with evaluating a specific chapter of a novel. Your role is to provide constructive feedback on the most recently submitted chapter, identifying strengths and suggesting areas for improvement within that particular chapter. Respond in the format provided below, offering only critique and no additional explanations. The output must be in Russian.

Critique the provided "Last Written Chapter." Focus your feedback specifically on the content, style, pacing, character portrayal, and overall effectiveness of *that chapter alone*.

**Your Response Format:**

## Strengths
[Here, describe what aspects of the *last written chapter* were well-executed and promising, using a clear and structured format. For example: "Диалоги в этой главе оживили персонажей, раскрывая их характеры через реплики. Описание сцены [Название сцены] было ярким и атмосферным, создавая сильное погружение."]

## Improvements
[Here, offer specific suggestions for enhancing the *last written chapter*, focusing on clarity, depth, and engagement within that chapter, using a clear and structured format. For example: "Темп повествования в середине главы замедлился; можно было бы добавить больше динамики или сократить описания. Реакция персонажа [Имя персонажа] на события [Название события] показалась несколько неправдоподобной; стоит проработать его внутренние мотивы."]

## Impressions
[Here, provide a brief, overall impression of the *last written chapter*. For example: "Глава успешно продвинула сюжет и показала развитие отношений между героями, но требует некоторой полировки в плане темпа."]

## Score
[Provide a final score from 0 to 99 for the *last written chapter*. Only the number is required.]`
}

func (p *ruPrompts) WriteNChapterPrompt(chapter int) string {
	return fmt.Sprintf(`Напиши %d главу в соответствии с планом.`, chapter)
}

func (p *ruPrompts) CritiqueNChapterPrompt(chapter int) string {
	return fmt.Sprintf(`Критикуй %d главу.`, chapter)
}

func (p *ruPrompts) UpdateNChapterPrompt(chapter int) string {
	return fmt.Sprintf(`Перепиши %d главу в соответствии с рекомендациями.`, chapter)
}
