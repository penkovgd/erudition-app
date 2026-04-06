import logging
from langchain_community.chat_models import ChatYandexGPT
from langchain.output_parsers import PydanticOutputParser
from pydantic import SecretStr

from src.models import GeneratedMatching
from src.prompts.matching import matching_generator_template

logger = logging.getLogger(__name__)


class MatchingGenerator:
    def __init__(
        self,
        api_key: str,
        folder_id: str,
        model_name: str = "yandexgpt-5-lite/latest",
        temperature: float = 0.3,
    ):
        # Температура 0.3 - чтобы формулировки заданий были живыми, но пары оставались точными
        self.llm = ChatYandexGPT(
            model_uri=f"gpt://{folder_id}/{model_name}",
            temperature=temperature,
            api_key=SecretStr(api_key),
        )
        self.parser = PydanticOutputParser(pydantic_object=GeneratedMatching)
        self.chain = matching_generator_template | self.llm | self.parser

    async def agenerate(
        self,
        topic_name: str,
        relation_label: str,
        relation_desc: str,
        pairs_context: str,
    ) -> GeneratedMatching:
        """
        Асинхронно генерирует вопрос на сопоставление (Matching).
        """
        logger.debug(f"Генерация Matching для отношения: {relation_label}")

        result: GeneratedMatching = await self.chain.ainvoke(
            {
                "topic_name": topic_name,
                "relation_label": relation_label,
                "relation_desc": relation_desc,
                "pairs_context": pairs_context,
                "format_instructions": self.parser.get_format_instructions(),
            }
        )

        if not isinstance(result, GeneratedMatching):
            raise TypeError(f"Ожидался GeneratedMatching, получен {type(result)}")

        return result
