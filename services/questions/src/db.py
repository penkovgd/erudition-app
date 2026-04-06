import json
import hashlib
import logging
import asyncpg
from typing import List, Dict, Optional, Any

from src.models import TopicInfo

logger = logging.getLogger(__name__)


async def upsert_topic(pool: asyncpg.Pool, topic: TopicInfo):
    """
    Создает или обновляет топик в PostgreSQL из DTO.
    """
    query = """
        INSERT INTO topics (slug, name, description)
        VALUES ($1, $2, $3)
        ON CONFLICT (slug) DO UPDATE 
        SET name = EXCLUDED.name,
            description = EXCLUDED.description;
    """
    try:
        async with pool.acquire() as conn:
            await conn.execute(query, topic.slug, topic.name, topic.description)
            logger.info(f"Топик '{topic.slug}' синхронизирован с БД.")
    except Exception as e:
        logger.error(f"Ошибка при синхронизации топика: {e}")
        raise


def generate_context_hash(topic: str, context: str) -> str:
    """Генерирует уникальный SHA-256 хэш на основе топика и пути."""
    raw = f"{topic}|{context}".encode("utf-8")
    return hashlib.sha256(raw).hexdigest()


async def save_approved_questions(pool: asyncpg.Pool, questions: List[dict]):
    if not questions:
        return

    query = """
        INSERT INTO questions (
            topic_slug, q_type, bloom_level, anchor, context_path, 
            context_hash, stem, payload, justification
        ) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9)
        ON CONFLICT (context_hash) DO NOTHING;
    """

    data_to_insert = []
    for q in questions:
        # Для MCQ контекст это один путь, для Matching - список пар. Хэш работает для всего.
        ctx_hash = q.get("context_hash") or generate_context_hash(
            q["topic_slug"], q["context_path"]
        )

        data_to_insert.append(
            (
                q["topic_slug"],
                q["q_type"],
                q["bloom_level"],
                q["anchor"],
                q["context_path"],
                ctx_hash,
                q["stem"],
                json.dumps(q["payload"], ensure_ascii=False),  # Универсальный JSONB
                q["justification"],
            )
        )

    try:
        async with pool.acquire() as conn:
            await conn.executemany(query, data_to_insert)
        logger.info(f"Успешно сохранено {len(questions)} вопросов.")
    except Exception as e:
        logger.error(f"Ошибка при сохранении вопросов: {e}")
        raise


async def get_questions(
    pool: asyncpg.Pool, topic_slug: Optional[str] = None, quantity: int = 10
) -> List[Dict[str, Any]]:
    """
    Получает список вопросов из БД.
    Если topic_slug не указан, ищет по всем топикам.
    Если quantity = 0, возвращает все вопросы без лимита.
    """
    query = """
        SELECT id, topic_slug, q_type, bloom_level, anchor, context_path, 
               context_hash, stem, payload, justification, created_at
        FROM questions
    """
    args = []

    # Динамическая сборка запроса
    if topic_slug:
        args.append(topic_slug)
        query += f" WHERE topic_slug = ${len(args)}"

    query += " ORDER BY RANDOM()"

    if quantity > 0:
        args.append(quantity)
        query += f" LIMIT ${len(args)}"

    try:
        async with pool.acquire() as conn:
            records = await conn.fetch(query, *args)

        result = []
        for r in records:
            row = dict(r)
            if isinstance(row["payload"], str):
                row["payload"] = json.loads(row["payload"])
            result.append(row)

        return result
    except Exception as e:
        logger.error(f"Ошибка при получении вопросов: {e}")
        raise
