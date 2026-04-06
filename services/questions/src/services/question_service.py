import logging
import asyncpg
from src.services.mcq_pipeline import (
    QuestionGenerationPipeline,
)  # Переименованный старый пайплайн
from src.services.matching_pipeline import MatchingGenerationPipeline  # Новый пайплайн
from src.db import save_approved_questions, upsert_topic

logger = logging.getLogger(__name__)


async def generate_mcq_task(
    db_pool: asyncpg.Pool,
    neo4j_args: dict,
    llm_args: dict,
    topic_slug: str,
    anchors_limit: int,
    paths_per_anchor: int,
):
    """Фоновая задача для MCQ"""
    logger.info(f"[MCQ Task] Запуск для '{topic_slug}'...")
    try:
        pipeline = QuestionGenerationPipeline(
            **neo4j_args, **llm_args, max_concurrent_requests=5
        )
        topic_info, approved_questions = await pipeline.arun(
            topic_slug=topic_slug,
            anchors_limit=anchors_limit,
            paths_per_anchor=paths_per_anchor,
        )
        await upsert_topic(db_pool, topic_info)
        if approved_questions:
            await save_approved_questions(db_pool, approved_questions)
        logger.info(f"[MCQ Task] Завершено для '{topic_slug}'!")
    except Exception as e:
        logger.error(f"[MCQ Task] Ошибка: {e}")


async def generate_matching_task(
    db_pool: asyncpg.Pool,
    neo4j_args: dict,
    llm_args: dict,
    topic_slug: str,
    pairs_count: int,
):
    """Фоновая задача для Matching"""
    logger.info(f"[Matching Task] Запуск для '{topic_slug}'...")
    try:
        pipeline = MatchingGenerationPipeline(
            **neo4j_args, **llm_args, max_concurrent=5
        )
        # Здесь свой метод arun со специфичными параметрами
        topic_info, approved_questions = await pipeline.arun(
            topic_slug=topic_slug, pairs_count=pairs_count
        )
        await upsert_topic(db_pool, topic_info)
        if approved_questions:
            await save_approved_questions(db_pool, approved_questions)
        logger.info(f"[Matching Task] Завершено для '{topic_slug}'!")
    except Exception as e:
        logger.error(f"[Matching Task] Ошибка: {e}")
