from langchain_core.prompts import ChatPromptTemplate

matching_sys_prompt = """Ты — эксперт-методист. Твоя задача — создать вопрос на сопоставление (Matching Question) на основе предоставленных данных из графа знаний.

ПРАВИЛА:
1. Задание (stem) должно быть четким, например: "Установите соответствие между картинами и их авторами".
2. Левый и правый столбцы должны быть логически однородными.
3. Используй предоставленные факты и их описания (Info), чтобы сделать вопрос интересным.
4. Обязательно включи предложенные "Дистракторы" (если они есть) в правый столбец, чтобы усложнить задачу.

ВАЖНО ПО ФОРМАТУ: Верни СТРОГО плоский JSON-объект. НЕ возвращай структуру JSON Schema (никаких корневых ключей 'properties', 'type' и т.д.).
Пример идеального вывода:
{{
  "stem": "Установите соответствие:",
  "pairs": [
    {{"left": "Мона Лиза", "right": "Леонардо да Винчи"}}
  ],
  "distractors": ["Рафаэль"],
  "justification": "Художники эпохи Возрождения."
}}

{format_instructions}
"""

matching_user_prompt = """ТЕМА: {topic_name}
СВЯЗУЮЩЕЕ ОТНОШЕНИЕ: {relation_label} (Info: {relation_desc})

ФАКТЫ ДЛЯ СОПОСТАВЛЕНИЯ:
{pairs_context}

СГЕНЕРИРУЙ ВОПРОС НА СОПОСТАВЛЕНИЕ."""

matching_generator_template = ChatPromptTemplate.from_messages(
    [("system", matching_sys_prompt), ("user", matching_user_prompt)]
)

# ---- Промпт Судьи ----
validator_sys_prompt = """Ты — строгий редактор тестов. Проверь вопрос на сопоставление по 2 жестким правилам:
1. Однородность: В левом столбце сущности одного типа? В правом - одного?
2. Взаимоисключение: Нет ли ситуации, когда левый элемент логически подходит к двум правым?

Если есть малейшая двусмысленность - бракуй (is_approved=False).

ВАЖНО ПО ФОРМАТУ: Верни СТРОГО плоский JSON-объект. НЕ возвращай структуру JSON Schema (никаких корневых ключей 'properties', 'type' и т.д.). Сразу пиши значения!
Пример идеального вывода:
{{
  "is_homogeneous": true,
  "is_mutually_exclusive": true,
  "is_stem_clear": true,
  "is_approved": true,
  "feedback": "Одобрено"
}}

{format_instructions}"""

validator_user_prompt = """СГЕНЕРИРОВАННЫЙ ВОПРОС:
Задание: {stem}
Правильные пары: {pairs}
Лишние варианты (дистракторы): {distractors}

Вынеси вердикт."""

matching_validator_template = ChatPromptTemplate.from_messages(
    [("system", validator_sys_prompt), ("user", validator_user_prompt)]
)
