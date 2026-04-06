from typing import List, Optional

from pydantic import BaseModel, Field


class TopicInfo(BaseModel):
    slug: str = Field(description="Уникальный идентификатор (URI) топика из графа")
    name: str = Field(description="Человекочитаемое название топика")
    description: str = Field(description="Подробное описание топика для контекста LLM")


class GeneratedMCQ(BaseModel):
    question: str = Field(
        description="Текст вопроса (stem). Должен быть понятен без вариантов ответа."
    )
    image_url: Optional[str] = Field(
        default=None,
        description="Ссылка на картинку, если она есть в контексте. Иначе null.",
    )
    option_a: str = Field(description="Вариант ответа A")
    option_b: str = Field(description="Вариант ответа B")
    option_c: str = Field(description="Вариант ответа C")
    option_d: str = Field(description="Вариант ответа D")
    correct_key: str = Field(description="Строго одна буква: A, B, C или D")
    justification: str = Field(
        description="Краткое объяснение, почему дистракторы неверны, а ответ верен. (Помогает модели лучше 'думать')"
    )


class ValidationResult(BaseModel):
    is_faithful: bool = Field(
        description="True, если ответ можно вывести строго из предоставленного контекста графа."
    )
    is_single_correct: bool = Field(
        description="True, если только один вариант ответа является верным."
    )
    is_good_distractors: bool = Field(
        description="True, если дистракторы правдоподобны, уникальны и нет запрещенных фраз ('Всё вышеперечисленное')."
    )
    is_stem_clear: bool = Field(
        description="True, если вопрос имеет смысл без чтения вариантов ответа."
    )
    is_grammatically_correct: bool = Field(
        description="True, если нет ошибок и опечаток."
    )

    is_approved: bool = Field(
        description="True, ТОЛЬКО ЕСЛИ все предыдущие 5 флагов равны True. Иначе False."
    )
    feedback: str = Field(
        description="Если is_approved == False, напиши КРАТКУЮ причину, что именно нужно исправить. Если True, напиши 'Одобрено'."
    )


class MatchingPair(BaseModel):
    left: str = Field(description="Элемент левого столбца (например, картина)")
    right: str = Field(
        description="Соответствующий правильный элемент правого столбца (например, автор)"
    )


class GeneratedMatching(BaseModel):
    stem: str = Field(
        description="Текст задания (например: 'Сопоставьте картины с их создателями:')"
    )
    pairs: List[MatchingPair] = Field(description="Список правильных пар")
    distractors: List[str] = Field(
        description="Дополнительные элементы для правого столбца (для усложнения), ни к чему не подходящие"
    )
    justification: str = Field(description="Краткое объяснение логики сопоставления")


class MatchingValidationResult(BaseModel):
    is_homogeneous: bool = Field(
        description="True, если элементы в левом столбце однородны, и элементы в правом - тоже."
    )
    is_mutually_exclusive: bool = Field(
        description="True, если нет неоднозначностей (каждый элемент слева подходит СТРОГО к одному элементу справа)."
    )
    is_stem_clear: bool = Field(description="True, если задание понятно.")
    is_approved: bool = Field(
        description="True ТОЛЬКО если все проверки выше пройдены."
    )
    feedback: str = Field(
        description="Причина отклонения, если False. Иначе 'Одобрено'."
    )
