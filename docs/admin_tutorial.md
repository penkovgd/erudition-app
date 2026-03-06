# Руководство админа

## Как составлять SPARQL запросы

### 1. составить SELECT запрос, чтобы понять, какие данные нужны

Переходим на https://query.wikidata.org/

Пример запроса: https://w.wiki/JCb8

```sql
SELECT ?russianWikipediaLink ?paintingLabel ?creatorLabel ?inception ?image ?movementLabel ?sitelinks WHERE {
  ?painting wdt:P31 wd:Q3305213;
            wdt:P135 ?movement;
            wdt:P571 ?inception;
            wdt:P170 ?creator;
            wdt:P18 ?image.
  ?painting wikibase:sitelinks ?sitelinks. 
  
  ?russianWikipediaLink schema:about ?painting;
    schema:isPartOf <https://ru.wikipedia.org/>.
  
  SERVICE wikibase:label { bd:serviceParam wikibase:language "ru". }
}
ORDER BY DESC(?sitelinks)
LIMIT 100
```

Что учитывать в запросе:

- Поиск только в русской википедии.

Так как приложение будет поддрежривать только русский язык (пока), все знания должны быть на русском языке.

```sql
?russianWikipediaLink schema:about ?painting;
    schema:isPartOf <https://ru.wikipedia.org/>.
```

Поэтому можно у Q-сущности потребовать ее ссылку на русскую вики. Таким образом, мы получаем как ссылку на вики, и фильтруем, что все лейблы будут на русском.

- Русские лейблы сущностей.

```sql
  SERVICE wikibase:label { bd:serviceParam wikibase:language "ru". }
```

Сервис лейблов позволяет из сущности (пр. ?painting -> Q123) получать ее лейбл и описание на разных языках (?paintingLabel -> "Мона Лиза") автоматически.

- Сайтлинки в википедии

```sql
  ?painting wikibase:sitelinks ?sitelinks. 
```

Один из самых простых способов проверить известноть - посчитать, сколько у сущности в wikidata есть сайтлинков. Сайтлинки (Sitelinks) — это ссылки на статьи в других проектах Викимедиа, связанные с данным элементом.

Например, у Моны Лизы есть 144 сайтлинков - это разные статьи википедии на разных языках и на другие сайты проекта вики.

### 2. Проверям запрос

Запускаем запрос на WDQS и смотрим, все ли хорошо. Напримре, может быть такое, что некоторые лейблы все-таки не на русском языке.

### 3. на основе SELECT запроса составляем CONSTRUCT запрос.

select выдает просто табличку с данными, а нам нужен граф (триплеты). Это делает CONSTRUCT запрос.

Пример: https://w.wiki/JChJ

```sql
CONSTRUCT {
  ?painting rdf:type wd:Q3305213;
            rdfs:label ?paintingLabel;
            wdt:P135 ?movement;
            wdt:P571 ?inception;
            wdt:P170 ?creator;
            wdt:P18 ?image.
  ?creator rdfs:label ?authorLabel.
  ?movement rdfs:label ?movementLabel.          
}
WHERE {
  ?painting wdt:P31 wd:Q3305213;
            wdt:P135 ?movement;
            wdt:P571 ?inception;
            wdt:P170 ?creator;
            wdt:P18 ?image.
  
  ?painting wikibase:sitelinks ?sitelinks. 
  
  ?russianWikipediaLink schema:about ?painting;
    schema:isPartOf <https://ru.wikipedia.org/>.
  
  SERVICE wikibase:label { bd:serviceParam wikibase:language "ru". }
}
ORDER BY DESC(?sitelinks)
LIMIT 3
```

Что учесть:

- блок CONSTRUCT: это те тройки, которые мы составляем. Подгоняем граф для neo4j:
    - *wdt:P31 (instanceOf) -> rdf:type. Таким образом, neo4j сможет дать сущности класс Q3305213 (Painting).
    - rdfs:label пойдет в label в neo4j. чтобы отображалось не Q123 а "Мона Лиза".

- блок WHERE, ORDER, LIMIT: копируем из SELECT-а предыдущего запроса.

