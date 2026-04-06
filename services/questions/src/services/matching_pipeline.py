import asyncio
import logging
from typing import List, Optional, Tuple
from src.models import TopicInfo, GeneratedMatching
from src.services.retriever import GraphRetriever
from src.services.matching_generator import MatchingGenerator
from src.services.matching_validator import MatchingValidator

logger = logging.getLogger(__name__)


class MatchingGenerationPipeline:
    def __init__(
        self,
        neo4j_uri,
        neo4j_user,
        neo4j_pass,
        llm_api_key,
        llm_folder_id,
        max_concurrent=5,
    ):
        self.retriever = GraphRetriever(neo4j_uri, neo4j_user, neo4j_pass)
        self.generator = MatchingGenerator(
            api_key=llm_api_key,
            folder_id=llm_folder_id,
            model_name="yandexgpt-5-lite/latest",
            temperature=0.3,
        )
        self.validator = MatchingValidator(
            api_key=llm_api_key,
            folder_id=llm_folder_id,
            model_name="yandexgpt-5-lite/latest",
        )
        self.semaphore = asyncio.Semaphore(max_concurrent)

    async def _process_cluster(self, topic: TopicInfo, cluster: dict) -> Optional[dict]:
        async with self.semaphore:
            # Формируем читаемый контекст из кластера
            pairs_str = "\n".join(
                [
                    f"- Лево: [{p['left_label']}] (Info: {p['left_desc']}) ---> Право: [{p['right_label']}] (Info: {p['right_desc']})"
                    for p in cluster["pairs"]
                ]
            )

            if cluster["distractors"]:
                dist_str = "\n".join(
                    [
                        f"- [{d['label']}] (Info: {d['desc']})"
                        for d in cluster["distractors"]
                    ]
                )
                pairs_str += f"\n\nДИСТРАКТОРЫ ДЛЯ ПРАВОГО СТОЛБЦА:\n{dist_str}"

            try:
                # 1. Генерация
                match_q: GeneratedMatching = await self.generator.agenerate(
                    topic_name=topic.name,
                    relation_label=cluster["relation_label"],
                    relation_desc=cluster["relation_desc"],
                    pairs_context=pairs_str,
                )

                # 2. Валидация
                validation = await self.validator.aevaluate(match_q)

                # 3. Возврат результата
                if validation.is_approved:
                    return {
                        "topic_slug": topic.slug,
                        "q_type": "MATCHING",
                        "bloom_level": 1,
                        "anchor": cluster[
                            "relation_label"
                        ],  # Для Matching якорем будет название отношения
                        "context_path": pairs_str,
                        "stem": match_q.stem,
                        "payload": {
                            "pairs": [
                                {"left": p.left, "right": p.right}
                                for p in match_q.pairs
                            ],
                            "distractors": match_q.distractors,
                        },
                        "justification": match_q.justification,
                    }
                return None
            except Exception as e:
                logger.error(f"Ошибка: {e}")
                return None

    async def arun(
        self, topic_slug: str, pairs_count: int = 4
    ) -> Tuple[TopicInfo, List[dict]]:
        topic_info = self.retriever.get_topic_info(topic_slug)
        clusters = self.retriever.get_matching_clusters(
            topic_slug, target_pairs=pairs_count
        )

        tasks = [self._process_cluster(topic_info, cluster) for cluster in clusters]
        results = await asyncio.gather(*tasks)

        return topic_info, [r for r in results if r is not None]
