from collections import defaultdict
import logging
import asyncio
from typing import List, Optional, Tuple
from src.services.retriever import GraphRetriever
from src.services.mcq_generator import MCQGenerator
from src.services.mcq_validator import MCQValidator
from src.models import GeneratedMCQ
from src.models import TopicInfo

logging.basicConfig(level=logging.INFO, format="%(levelname)s: %(message)s")
logger = logging.getLogger(__name__)


class QuestionGenerationPipeline:
    def __init__(
        self,
        neo4j_uri: str,
        neo4j_user: str,
        neo4j_pass: str,
        llm_api_key: str,
        llm_folder_id: str,
        generator_model: str = "yandexgpt-5-lite/latest",
        validator_model: str = "yandexgpt-5-lite/latest",
        max_concurrent_requests: int = 5,
    ):
        self.retriever = GraphRetriever(
            uri=neo4j_uri, user=neo4j_user, password=neo4j_pass
        )

        self.generator = MCQGenerator(
            api_key=llm_api_key,
            folder_id=llm_folder_id,
            model_name=generator_model,
            temperature=0.3,
        )
        self.validator = MCQValidator(
            api_key=llm_api_key, folder_id=llm_folder_id, model_name=validator_model
        )
        self.max_concurrent_requests = max_concurrent_requests

    async def _process_single_path(
        self,
        topic: TopicInfo,
        bloom_level: int,
        anchor_label: str,
        path_data: dict,
        forbidden_distractors: str,
        semaphore: asyncio.Semaphore,
    ) -> Optional[dict]:
        """
        Изолированная асинхронная задача для обработки одного пути.
        """
        async with semaphore:
            target_answer = path_data["target_answer"]

            detected_image_link = None  # Инициализируем пустую переменную

            if target_answer.startswith("http") and ("/" in target_answer):
                detected_image_link = target_answer  # Сохраняем ссылку надежно в Python
                target_answer = anchor_label
                path_context = f"Сущность [{anchor_label}] выглядит так. Вот ссылка на её изображение: {detected_image_link}"
            # ==============================================

            path_context = self.retriever.format_path_to_text(path_data)

            # logger.info(f"Генерация вопроса для пути {path_context}")

            try:
                # 1. Генерация (ждем ответа LLM)
                mcq: GeneratedMCQ = await self.generator.agenerate(
                    topic=topic,
                    path_context=path_context,
                    target_answer=target_answer,
                    bloom_level=bloom_level,
                    forbidden_distractors=forbidden_distractors,
                )

                # 2. Валидация (ждем ответа LLM)
                validation = await self.validator.aevaluate(
                    path_context=path_context,
                    mcq=mcq,
                    forbidden_distractors=forbidden_distractors,
                )

                # 3. Принятие решения
                if validation.is_approved:
                    logger.info(
                        f"""
                        ОДОБРЕНО:
                            '{mcq.question}'
                        варианты ответа:
                            a) {mcq.option_a}
                            b) {mcq.option_b}
                            c) {mcq.option_c}
                            d) {mcq.option_d}
                        путь:
                            {path_context}
                        """
                    )
                    return {
                        "topic_slug": topic.slug,
                        "q_type": "MCQ",  # Указываем тип!
                        "bloom_level": bloom_level,
                        "anchor": anchor_label,
                        "context_path": path_context,  # Переименовал context в context_path для единообразия
                        # context_hash генерируется в db.py, либо можете генерировать его тут
                        "stem": mcq.question,
                        "payload": {  # УПАКОВЫВАЕМ СПЕЦИФИЧНЫЕ ДАННЫЕ СЮДА
                            "options": {
                                "A": mcq.option_a,
                                "B": mcq.option_b,
                                "C": mcq.option_c,
                                "D": mcq.option_d,
                            },
                            "correct_key": mcq.correct_key,
                            "image_url": detected_image_link or mcq.image_url,
                        },
                        "justification": mcq.justification,
                    }
                else:
                    logger.warning(
                        f"""
                        ОТКЛОНЕНО: 
                            '{mcq.question}' 
                        варианты ответа: 
                            a) {mcq.option_a}
                            b) {mcq.option_b}
                            c) {mcq.option_c}
                            d) {mcq.option_d}
                        путь: 
                            {path_context}
                        Причина:
                            {validation.feedback}
                        """
                    )
                    return None

            except Exception as e:
                logger.error(f"Ошибка для '{anchor_label}': {e}")
                return None

    async def arun(
        self, topic_slug: str, anchors_limit: int = 0, paths_per_anchor: int = 0
    ) -> Tuple[TopicInfo, List[dict]]:

        semaphore = asyncio.Semaphore(self.max_concurrent_requests)

        topic_info: TopicInfo = self.retriever.get_topic_info(topic_slug)

        logger.info(f"Поиск якорей для топика '{topic_info.slug}'...")
        anchors = self.retriever.get_anchors(
            topic_slug=topic_info.slug, limit=anchors_limit
        )
        logger.info(f"Найдено {len(anchors)} якорей.")

        tasks = []

        for anchor in anchors:
            anchor_uri = anchor["anchor_uri"]
            anchor_label = anchor.get("anchor_label") or anchor_uri

            for bloom_level in [1, 2]:
                paths = self.retriever.get_paths_for_anchor(
                    anchor_uri=anchor_uri,
                    topic_slug=topic_info.slug,
                    hops=bloom_level,
                    limit=paths_per_anchor,
                )

                grouped_paths = defaultdict(list)
                for path_data in paths:
                    # Создаем уникальную сигнатуру пути (всё, кроме последнего узла)
                    sig_nodes = tuple(path_data["node_names"][:-1])
                    sig_rels = tuple(path_data["relations"])
                    grouped_paths[(sig_nodes, sig_rels)].append(path_data)

                for signature, group in grouped_paths.items():
                    # Собираем ВСЕ правильные ответы для этой сигнатуры
                    all_valid_targets = [p["target_answer"] for p in group]

                    for path_data in group:
                        target = path_data["target_answer"]
                        # Запрещаем использовать соседей по группе как дистракторы
                        forbidden = [t for t in all_valid_targets if t != target]
                        forbidden_str = ", ".join(forbidden) if forbidden else "Нет"

                        # Передаем forbidden_str в задачу
                        tasks.append(
                            self._process_single_path(
                                topic_info,
                                bloom_level,
                                anchor_label,
                                path_data,
                                forbidden_str,  # <-- НОВЫЙ АРГУМЕНТ
                                semaphore,
                            )
                        )

        # Запускаем все задачи конкурентно.
        # asyncio.gather вернет список результатов, сохранив порядок запуска.
        results = await asyncio.gather(*tasks)

        # Фильтруем None (отклоненные вопросы и ошибки)
        approved_questions = [r for r in results if r is not None]

        logger.info(
            f"Пайплайн завершен. Успешно создано: {len(approved_questions)} из {len(tasks)} ==="
        )
        return topic_info, approved_questions
