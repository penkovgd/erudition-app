FROM python:3.9-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

COPY services/questions/requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Копируем исходники вопросов
COPY services/questions/ ./

# В docker-compose для questions проброшен порт 8080
EXPOSE 8080

CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8080"]