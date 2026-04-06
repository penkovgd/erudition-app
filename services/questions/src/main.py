import os
import asyncpg
import logging
from typing import Optional, List, Dict, Any
from contextlib import asynccontextmanager
from fastapi import FastAPI, BackgroundTasks, status, Query
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from dotenv import load_dotenv
from uuid import UUID
from datetime import datetime

from src.db import get_questions
from src.migrator import run_migrations

# Импортируем оба таска из сервиса
from src.services.question_service import generate_mcq_task, generate_matching_task

logging.basicConfig(level=logging.INFO, format="%(levelname)s: %(message)s")

load_dotenv()

# === Конфигурация из переменных окружения ===
NEO4J_URI = os.getenv("NEO4J_URI", "bolt://localhost:7687")
NEO4J_USER = os.getenv("NEO4J_USER", "neo4j")
NEO4J_PASS = os.getenv("NEO4J_PASS", "neo4j")

PG_DSN = os.getenv("PG_DSN", "postgresql://user:password@localhost/db")

YC_FOLDER = os.getenv("YC_FOLDER", "your-folder-id")
YC_API_KEY = os.getenv("YC_API_KEY", "your-api-key")

# Группируем аргументы для удобной передачи в таски
neo4j_args = {
    "neo4j_uri": NEO4J_URI,
    "neo4j_user": NEO4J_USER,
    "neo4j_pass": NEO4J_PASS,
}

llm_args = {
    "llm_api_key": YC_API_KEY,
    "llm_folder_id": YC_FOLDER,
}

# Глобальная переменная для пула БД
db_pool: Optional[asyncpg.Pool] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    await run_migrations(PG_DSN)

    global db_pool
    logging.info("Подключение к PostgreSQL...")
    db_pool = await asyncpg.create_pool(dsn=PG_DSN, min_size=1, max_size=10)

    yield  # Сервер работает

    logging.info("Закрытие соединений с PostgreSQL...")
    if db_pool:
        await db_pool.close()


app = FastAPI(title="Erudition Question Generator API", lifespan=lifespan)


app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Разрешает запросы с любых доменов
    allow_credentials=False,  # При allow_origins=["*"] это значение должно быть False
    allow_methods=[
        "*"
    ],  # Разрешает все методы (GET, POST, PUT, DELETE, OPTIONS и т.д.)
    allow_headers=["*"],  # Разрешает все заголовки
)
# === Схемы запросов на генерацию ===


class GenerateMCQRequest(BaseModel):
    topic_slug: str = Field(..., description="Слаг топика")
    anchors_limit: int = Field(
        default=5, ge=1, le=1000, description="Сколько якорей взять"
    )
    paths_per_anchor: int = Field(
        default=2, ge=1, le=10, description="Сколько путей для якоря"
    )


class GenerateMatchingRequest(BaseModel):
    topic_slug: str = Field(..., description="Слаг топика")
    pairs_count: int = Field(
        default=4, ge=2, le=10, description="Сколько пар в одном вопросе (обычно 4-5)"
    )


class SuccessResponse(BaseModel):
    message: str
    topic_slug: str


# === Эндпоинты Генерации ===


@app.post(
    "/api/v1/generate/mcq",
    status_code=status.HTTP_202_ACCEPTED,
    response_model=SuccessResponse,
)
async def trigger_mcq_generation(
    req: GenerateMCQRequest, background_tasks: BackgroundTasks
):
    """Запускает фоновую генерацию вопросов типа MCQ (множественный выбор)."""
    if db_pool is None:
        raise ValueError("Database pool is not initialized")

    background_tasks.add_task(
        generate_mcq_task,
        db_pool=db_pool,
        neo4j_args=neo4j_args,
        llm_args=llm_args,
        topic_slug=req.topic_slug,
        anchors_limit=req.anchors_limit,
        paths_per_anchor=req.paths_per_anchor,
    )

    return SuccessResponse(
        message="Задача на генерацию MCQ успешно добавлена в очередь.",
        topic_slug=req.topic_slug,
    )


@app.post(
    "/api/v1/generate/matching",
    status_code=status.HTTP_202_ACCEPTED,
    response_model=SuccessResponse,
)
async def trigger_matching_generation(
    req: GenerateMatchingRequest, background_tasks: BackgroundTasks
):
    """Запускает фоновую генерацию вопросов типа Matching (сопоставление)."""
    if db_pool is None:
        raise ValueError("Database pool is not initialized")

    background_tasks.add_task(
        generate_matching_task,
        db_pool=db_pool,
        neo4j_args=neo4j_args,
        llm_args=llm_args,
        topic_slug=req.topic_slug,
        pairs_count=req.pairs_count,
    )

    return SuccessResponse(
        message="Задача на генерацию Matching успешно добавлена в очередь.",
        topic_slug=req.topic_slug,
    )


# === Схемы ответов (Получение из БД) ===


class QuestionResponse(BaseModel):
    id: UUID
    topic_slug: str
    q_type: str  # Тип вопроса (MCQ или MATCHING)
    bloom_level: int
    anchor: str
    context_path: str
    context_hash: str
    stem: str
    payload: Dict[
        str, Any
    ]  # Универсальное поле для данных ответа (варианты, пары, картинки)
    justification: Optional[str]
    created_at: datetime


# === Эндпоинт Получения вопросов ===


@app.get(
    "/api/v1/questions",
    status_code=status.HTTP_200_OK,
    response_model=List[QuestionResponse],
)
async def fetch_questions(
    topic_slug: Optional[str] = Query(
        default=None, description="Слаг топика. Если не передан, вернутся все."
    ),
    quantity: int = Query(
        default=10, ge=0, description="Количество вопросов (0 - все)."
    ),
):
    """Возвращает список сгенерированных вопросов из базы данных."""
    if db_pool is None:
        raise ValueError("Database pool is not initialized")

    questions = await get_questions(
        pool=db_pool, topic_slug=topic_slug, quantity=quantity
    )
    return questions
