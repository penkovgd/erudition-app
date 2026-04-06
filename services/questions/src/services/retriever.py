import random
from venv import logger

from langchain_neo4j import Neo4jGraph

from src.models import TopicInfo


class GraphRetriever:
    def __init__(self, uri, user, password):
        self.graph = Neo4jGraph(url=uri, username=user, password=password)

    def get_anchors(
        self, topic_slug: str, anchor_label: str = "Resource", limit: int = 0
    ):
        """
        Ищет узлы-якоря по топику.
        Например, мы ищем все узлы, у которых в массиве topics есть наш топик.
        """
        query = f"""
            MATCH (anchor:{anchor_label})
            WHERE $topic_slug IN anchor.topics
            RETURN anchor.uri AS anchor_uri, anchor.label AS anchor_label, anchor.description AS description
            // RETURN anchor.uri AS anchor_uri, anchor.label AS anchor_label
            """
        if limit != 0:
            query += " LIMIT $limit"

        return self.graph.query(
            query, params={"topic_slug": topic_slug, "limit": limit}
        )

    def get_paths_for_anchor(
        self, anchor_uri: str, topic_slug: str, hops: int, limit: int = 0
    ):
        """
        Для конкретного якоря достает пути нужной длины.
        """
        query = f"""
            MATCH path = (start:Resource {{uri: $anchor_uri}})-[:REL*{hops}]->(end_node)

            // Гарантируем, что связи относятся к топику, и мы не зациклились
            WHERE ALL(r IN relationships(path) WHERE $topic_slug IN r.topics)
            AND start <> end_node
            // Конечный узел может быть любым из этих двух типов
            AND (end_node:Resource OR end_node:Literal)

            RETURN 
                // Формируем массив имен узлов с умным фоллбэком
                [node IN nodes(path) | 
                    CASE 
                        WHEN node:Literal THEN node.value 
                        WHEN node.label IS NOT NULL AND node.label <> "" THEN node.label 
                        ELSE node.uri 
                    END
                ] AS node_names,

                [node IN nodes(path) | coalesce(node.description, "")] AS node_descriptions,
                
                [rel IN relationships(path) | rel.label] AS relations,

                [rel IN relationships(path) | coalesce(rel.description, "")] AS rel_descriptions,
                
                CASE 
                    WHEN end_node:Literal THEN end_node.value 
                    WHEN end_node.label IS NOT NULL AND end_node.label <> "" THEN end_node.label 
                    ELSE end_node.uri 
                END AS target_answer
                
            ORDER BY rand()  // Случайный порядок для разнообразия
        """
        if limit != 0:
            query += " LIMIT $limit"
        return self.graph.query(
            query,
            params={
                "anchor_uri": anchor_uri,
                "topic_slug": topic_slug,
                "limit": limit,
            },
        )

    def format_path_to_text(self, path_data: dict) -> str:
        # nodes = path_data["node_names"]
        # descs = path_data["node_descriptions"]
        # rels = path_data["relations"]

        # if not nodes:
        #     return ""

        # # Добавляем описание к первому узлу
        # path_str = f"[{nodes[0]}] (Info: {descs[0]})"

        # for i in range(len(rels)):
        #     # Добавляем связь и следующий узел с его описанием
        #     path_str += f" <{rels[i]}> [{nodes[i + 1]}] (Info: {descs[i + 1]})"

        # return path_str

        """
        Превращает сырые данные пути в линейный читаемый текст, включая описания.
        Пример: [Мона Лиза] (Info: картина Леонардо да Винчи) <это частный случай понятия> [картина]
        """
        nodes = path_data["node_names"]
        n_descs = path_data["node_descriptions"]
        rels = path_data["relations"]
        r_descs = path_data.get(
            "rel_descriptions", []
        )  # get для обратной совместимости

        if not nodes:
            return ""

        # Вспомогательная функция для форматирования узла
        def format_node(i):
            base = f"[{nodes[i]}]"
            # Если описание есть и оно не пустое, добавляем его в скобках
            return f"{base} (Info: {n_descs[i]})" if n_descs[i] else base

        # Вспомогательная функция для форматирования связи
        def format_rel(i):
            base = f"<{rels[i]}>"
            if i < len(r_descs) and r_descs[i]:
                return f"{base} (Info: {r_descs[i]})"
            return base

        # Собираем путь
        path_str = format_node(0)
        for i in range(len(rels)):
            path_str += f" {format_rel(i)} {format_node(i + 1)}"

        logger.debug(f"Formatted path: {path_str}")
        return path_str

    def get_topic_info(self, topic_slug: str) -> TopicInfo:
        """Получает мета-информацию о топике и возвращает Pydantic модель."""
        query = """
        MATCH (t:Topic {uri: $slug})
        RETURN t.name AS name, t.description AS description
        """
        result = self.graph.query(query, params={"slug": topic_slug})

        if not result:
            raise ValueError(f"Топик '{topic_slug}' не найден в графе знаний Neo4j.")

        return TopicInfo(
            slug=topic_slug,
            name=result[0].get("name") or topic_slug,
            description=result[0].get("description") or "",
        )

    import random

    def get_matching_clusters(
        self, topic_slug: str, target_pairs: int = 4
    ) -> list[dict]:
        """
        Ищет кластеры отношений для создания вопросов на сопоставление.
        Возвращает список кластеров, где каждый кластер содержит список уникальных пар без пересечений.
        """
        # Ищем все отношения в топике, группируем по типу отношения (rel_uri)
        query = """
        MATCH (s:Resource)-[r:REL]->(o)
        WHERE $topic_slug IN r.topics AND (o:Resource OR o:Literal)
        RETURN r.uri AS rel_uri, r.label AS rel_label, coalesce(r.description, "") AS rel_desc,
               collect(DISTINCT {
                   left_label: coalesce(s.label, s.uri), 
                   left_desc: coalesce(s.description, ""),
                   right_label: CASE WHEN o:Resource THEN coalesce(o.label, o.uri) ELSE o.value END,
                   right_desc: coalesce(o.description, "")
               }) AS all_pairs
        """
        results = self.graph.query(query, params={"topic_slug": topic_slug})

        valid_clusters = []

        for record in results:
            rel_label = record["rel_label"]
            all_pairs = record["all_pairs"]

            # Если пар меньше, чем нам нужно, пропускаем это отношение
            if len(all_pairs) < target_pairs:
                continue

            # Фильтруем пересечения (Python-логика биекции)
            # Мы не можем допустить, чтобы один и тот же автор или картина встречались дважды
            selected_pairs = []
            used_lefts = set()
            used_rights = set()

            # Перемешаем для случайности генерации
            random.shuffle(all_pairs)

            for pair in all_pairs:
                if (
                    pair["left_label"] not in used_lefts
                    and pair["right_label"] not in used_rights
                ):
                    selected_pairs.append(pair)
                    used_lefts.add(pair["left_label"])
                    used_rights.add(pair["right_label"])

                if len(selected_pairs) == target_pairs:
                    break

            if len(selected_pairs) == target_pairs:
                # Ищем дистракторы (элементы справа, которые не вошли в selected_pairs)
                # Например, художники, чьих картин нет в левом столбце
                distractors = []
                for p in all_pairs:
                    if (
                        p["right_label"] not in used_rights
                        and p["right_label"] not in distractors
                    ):
                        distractors.append(
                            {"label": p["right_label"], "desc": p["right_desc"]}
                        )
                        if len(distractors) >= 1:  # Берем 1-2 дистрактора
                            break

                valid_clusters.append(
                    {
                        "relation_label": rel_label,
                        "relation_desc": record["rel_desc"],
                        "pairs": selected_pairs,
                        "distractors": distractors,
                    }
                )

        return valid_clusters
