import os
import logging
import asyncpg
from pathlib import Path

logger = logging.getLogger(__name__)

MIGRATIONS_DIR = Path(__file__).parent.parent / "migrations"


async def run_migrations(dsn: str):
    """
    Проверяет и накатывает новые SQL-миграции при запуске приложения.
    """
    logger.info("Проверка миграций базы данных...")

    conn = await asyncpg.connect(dsn)

    try:
        await conn.execute("""
            CREATE TABLE IF NOT EXISTS schema_migrations (
                version VARCHAR(255) PRIMARY KEY,
                applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
            );
        """)

        applied_records = await conn.fetch("SELECT version FROM schema_migrations;")
        applied_migrations = {record["version"] for record in applied_records}

        if not MIGRATIONS_DIR.exists():
            logger.warning(f"Папка с миграциями не найдена: {MIGRATIONS_DIR}")
            return

        migration_files = sorted(
            [f for f in os.listdir(MIGRATIONS_DIR) if f.endswith(".sql")]
        )

        for filename in migration_files:
            if filename not in applied_migrations:
                filepath = MIGRATIONS_DIR / filename
                with open(filepath, "r", encoding="utf-8") as f:
                    sql_commands = f.read()

                logger.info(f"Применение миграции: {filename}...")

                async with conn.transaction():
                    await conn.execute(sql_commands)
                    await conn.execute(
                        "INSERT INTO schema_migrations (version) VALUES ($1)", filename
                    )

                logger.info(f"Миграция {filename} успешно применена.")

        logger.info("База данных актуальна.")

    except Exception as e:
        logger.error(f"Ошибка при выполнении миграций: {e}")
        raise
    finally:
        await conn.close()
