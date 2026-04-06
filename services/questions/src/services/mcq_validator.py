from langchain_community.chat_models import ChatYandexGPT
from langchain.output_parsers import PydanticOutputParser
from pydantic import SecretStr

from src.models import GeneratedMCQ, ValidationResult
from src.prompts.mcq_val import validator_prompt_template


class MCQValidator:
    def __init__(self, api_key: str, folder_id: str, model_name: str):
        self.llm = ChatYandexGPT(
            model_uri=f"gpt://{folder_id}/{model_name}",
            temperature=0.0,
            api_key=SecretStr(api_key),
        )
        self.parser = PydanticOutputParser(pydantic_object=ValidationResult)
        self.chain = validator_prompt_template | self.llm | self.parser

    def evaluate(
        self, path_context: str, mcq: GeneratedMCQ, forbidden_distractors: str
    ) -> ValidationResult:
        """
        Проверяет сгенерированный вопрос.
        """
        result = self.chain.invoke(
            {
                "path_context": path_context,
                "forbidden_distractors": forbidden_distractors,
                "question": mcq.question,
                "image_url": mcq.image_url,
                "option_a": mcq.option_a,
                "option_b": mcq.option_b,
                "option_c": mcq.option_c,
                "option_d": mcq.option_d,
                "correct_key": mcq.correct_key,
                "justification": mcq.justification,
                "format_instructions": self.parser.get_format_instructions(),
            }
        )

        if not isinstance(result, ValidationResult):
            raise TypeError(f"Ожидался ValidationResult, получен {type(result)}")

        return result

    async def aevaluate(
        self, path_context: str, mcq: GeneratedMCQ, forbidden_distractors: str
    ):
        """Асинхронная валидация вопроса."""
        # Используем ainvoke вместо invoke
        result = await self.chain.ainvoke(
            {
                "path_context": path_context,
                "forbidden_distractors": forbidden_distractors,
                "question": mcq.question,
                "image_url": mcq.image_url,
                "option_a": mcq.option_a,
                "option_b": mcq.option_b,
                "option_c": mcq.option_c,
                "option_d": mcq.option_d,
                "correct_key": mcq.correct_key,
                "justification": mcq.justification,
                "format_instructions": self.parser.get_format_instructions(),
            }
        )
        return result
