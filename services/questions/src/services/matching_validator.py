import logging
from langchain_community.chat_models import ChatYandexGPT
from langchain.output_parsers import PydanticOutputParser
from pydantic import SecretStr

from src.models import GeneratedMatching, MatchingValidationResult
from src.prompts.matching import matching_validator_template

logger = logging.getLogger(__name__)


class MatchingValidator:
    def __init__(
        self, api_key: str, folder_id: str, model_name: str = "yandexgpt-5-lite/latest"
    ):
        # Температура 0.0 для максимальной строгости и отсутствия галлюцинаций
        self.llm = ChatYandexGPT(
            model_uri=f"gpt://{folder_id}/{model_name}",
            temperature=0.0,
            api_key=SecretStr(api_key),
        )
        self.parser = PydanticOutputParser(pydantic_object=MatchingValidationResult)
        self.chain = matching_validator_template | self.llm | self.parser

    async def aevaluate(self, match_q: GeneratedMatching) -> MatchingValidationResult:
        """
        Проверяет сгенерированный вопрос на однородность и отсутствие множественных совпадений.
        """
        logger.debug(f"Валидация Matching вопроса: {match_q.stem}")

        # Подготавливаем данные из Pydantic модели в удобочитаемый текст для LLM-судьи
        pairs_str = "\n".join([f"- {p.left}  --->  {p.right}" for p in match_q.pairs])
        distractors_str = (
            ", ".join(match_q.distractors) if match_q.distractors else "Отсутствуют"
        )

        result: MatchingValidationResult = await self.chain.ainvoke(
            {
                "stem": match_q.stem,
                "pairs": pairs_str,
                "distractors": distractors_str,
                "format_instructions": self.parser.get_format_instructions(),
            }
        )

        if not isinstance(result, MatchingValidationResult):
            raise TypeError(
                f"Ожидался MatchingValidationResult, получен {type(result)}"
            )

        return result
