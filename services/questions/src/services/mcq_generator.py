from langchain_community.chat_models import ChatYandexGPT
from langchain.output_parsers import (
    PydanticOutputParser,
)
from pydantic import SecretStr
from src.models import GeneratedMCQ
from src.prompts.mcq_gen import mcq_prompt_template
from src.models import TopicInfo


class MCQGenerator:
    def __init__(self, api_key: str, folder_id: str, model_name: str, temperature=0.3):
        self.llm = ChatYandexGPT(
            model_uri=f"gpt://{folder_id}/{model_name}",
            temperature=temperature,
            api_key=SecretStr(api_key),
        )
        self.parser = PydanticOutputParser(pydantic_object=GeneratedMCQ)
        self.chain = mcq_prompt_template | self.llm | self.parser

    def _get_bloom_description(self, level: int) -> str:
        """Словарь глаголов и ожиданий для каждого уровня"""
        descriptions = {
            1: "Знание (Remember): Проверка фактов. Используй формулировки 'Кто', 'Что', 'Где'.",
            2: "Понимание (Understand): Понимание связей. Вопрос должен требовать прохождения по всей цепочке фактов от начала до конца.",
            3: "Анализ (Analyze): Глубокий вывод. Сформулируй вопрос так, чтобы студент должен был сопоставить факты из цепочки для получения ответа.",
        }
        return descriptions.get(level, descriptions[1])

    def generate(
        self,
        topic: TopicInfo,
        path_context: str,
        target_answer: str,
        bloom_level: int,
        forbidden_distractors: str,
    ) -> GeneratedMCQ:
        """
        Генерирует вопрос на основе одного пути.
        """
        bloom_desc = self._get_bloom_description(bloom_level)

        result = self.chain.invoke(
            {
                "topic_name": topic.name,
                "topic_description": topic.description,
                "path_context": path_context,
                "forbidden_distractors": forbidden_distractors,
                "target_answer": target_answer,
                "bloom_level": bloom_level,
                "bloom_description": bloom_desc,
                "format_instructions": self.parser.get_format_instructions(),
            }
        )

        if not isinstance(result, GeneratedMCQ):
            raise TypeError(f"Ожидался GeneratedMCQ, получен {type(result)}")

        return result

    async def agenerate(
        self,
        topic: TopicInfo,
        path_context: str,
        target_answer: str,
        bloom_level: int,
        forbidden_distractors: str,
    ) -> GeneratedMCQ:

        bloom_desc = self._get_bloom_description(bloom_level)

        # Распаковываем свойства модели прямо в промпт
        result = await self.chain.ainvoke(
            {
                "topic_name": topic.name,
                "topic_description": topic.description,
                "path_context": path_context,
                "forbidden_distractors": forbidden_distractors,
                "target_answer": target_answer,
                "bloom_level": bloom_level,
                "bloom_description": bloom_desc,
                "format_instructions": self.parser.get_format_instructions(),
            }
        )
        return result
