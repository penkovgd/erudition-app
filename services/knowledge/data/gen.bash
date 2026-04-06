#!/bin/bash

# Указываем целевую директорию
TARGET_DIR="topics"

# Создаем папку, если ее нет
mkdir -p "$TARGET_DIR"

echo "Начинаем генерацию 50 файлов топиков в папку $TARGET_DIR/..."

# Используем awk для парсинга и нарезки файлов
awk -v dir="$TARGET_DIR" '
/^# [a-z0-9-]+\.yaml/ {
    if (file) close(file);
    file = dir "/" $2;
    print "Создан: " file;
    next; # Переходим к следующей строке, само название файла внутрь не пишем
}
{
    # Пишем в файл только если мы уже знаем имя файла
    # Игнорируем декоративные разделители
    if (file && $0 !~ /^---/ && $0 !~ /^# ===/ && $0 !~ /^# КАТЕГОРИЯ/) {
        print $0 > file;
    }
}
' << 'EOF'
# ==========================================
# КАТЕГОРИЯ 1: Гуманитарные науки и Искусство
# ==========================================

# ancient-greek-philosophy.yaml
name: "Древнегреческая философия"
slug: "ancient-greek-philosophy"
description: "Школы и великие мыслители Древней Греции"
sparql: ""

concepts:
  - philosophy
  - greece
  - antiquity

---
# italian-renaissance-painters.yaml
name: "Итальянские художники эпохи Возрождения"
slug: "italian-renaissance-painters"
description: "Знаменитые художники и их шедевры в Италии эпохи Возрождения"
sparql: ""

concepts:
  - painting
  - italy
  - renaissance

---
# french-impressionism.yaml
name: "Французский импрессионизм"
slug: "french-impressionism"
description: "Картины и авторы французского импрессионизма"
sparql: ""

concepts:
  - painting
  - france
  - 19th_century

---
# russian-classical-literature.yaml
name: "Русская классическая литература"
slug: "russian-classical-literature"
description: "Произведения и авторы Золотого века русской литературы"
sparql: ""

concepts:
  - literature
  - russia
  - 19th_century

---
# soviet-cinema.yaml
name: "Советский кинематограф"
slug: "soviet-cinema"
description: "Культовые режиссеры и фильмы советского кинематографа"
sparql: ""

concepts:
  - cinema
  - ussr
  - 20th_century

---
# gothic-architecture-europe.yaml
name: "Готическая архитектура Европы"
slug: "gothic-architecture-europe"
description: "Самые известные готические соборы и здания Европы"
sparql: ""

concepts:
  - architecture
  - europe
  - middle_ages

---
# ancient-egypt-mythology.yaml
name: "Мифология Древнего Египта"
slug: "ancient-egypt-mythology"
description: "Египетские боги, их атрибуты и сферы влияния"
sparql: ""

concepts:
  - mythology
  - egypt
  - antiquity

---
# japanese-animation.yaml
name: "Японская анимация"
slug: "japanese-animation"
description: "Студия Ghibli, известные режиссеры и шедевры аниме"
sparql: ""

concepts:
  - cinema
  - japan
  - 20th_century
  - 21st_century

---
# british-rock-music.yaml
name: "Британская рок-музыка 1960-70х годов"
slug: "british-rock-music"
description: "Легендарные рок-группы и исполнители Великобритании"
sparql: ""

concepts:
  - music
  - uk
  - 20th_century

---
# german-classical-philosophy.yaml
name: "Немецкая классическая философия"
slug: "german-classical-philosophy"
description: "Идеи и труды Канта, Гегеля и других немецких мыслителей"
sparql: ""

concepts:
  - philosophy
  - germany
  - 18th_century
  - 19th_century

# ==========================================
# КАТЕГОРИЯ 2: История и Геополитика
# ==========================================

---
# roman-emperors.yaml
name: "Римские императоры"
slug: "roman-emperors"
description: "Династии, годы правления и факты о римских императорах"
sparql: ""

concepts:
  - history
  - italy
  - europe
  - antiquity

---
# ww2-battles.yaml
name: "Битвы Второй мировой войны"
slug: "ww2-battles"
description: "Ключевые сражения, даты и участники Второй мировой войны"
sparql: ""

concepts:
  - history
  - global
  - 20th_century

---
# us-presidents.yaml
name: "Президенты США"
slug: "us-presidents"
description: "Штаты рождения, партии и годы президентства в США"
sparql: ""

concepts:
  - history
  - usa
  - 18th_century
  - 19th_century
  - 20th_century
  - 21st_century

---
# chinese-dynasties.yaml
name: "Династии китайских императоров"
slug: "chinese-dynasties"
description: "Эпохи правления и правители великих династий Китая"
sparql: ""

concepts:
  - history
  - china
  - antiquity
  - middle_ages

---
# cold-war-crises.yaml
name: "Холодная война"
slug: "cold-war-crises"
description: "Кризисы, договоры и ключевые события эпохи Холодной войны"
sparql: ""

concepts:
  - history
  - usa
  - ussr
  - 20th_century

---
# french-revolution.yaml
name: "Великая французская революция"
slug: "french-revolution"
description: "Деятели, события и последствия Французской революции"
sparql: ""

concepts:
  - history
  - france
  - 18th_century

---
# italian-unification.yaml
name: "Объединение Италии"
slug: "italian-unification"
description: "События и герои Рисорджименто (объединения Италии)"
sparql: ""

concepts:
  - history
  - italy
  - 19th_century

---
# rulers-of-russia.yaml
name: "Правители Киевской Руси и Российской Империи"
slug: "rulers-of-russia"
description: "Князья, цари и императоры в истории России"
sparql: ""

concepts:
  - history
  - russia
  - middle_ages
  - 18th_century
  - 19th_century

---
# age-of-discovery.yaml
name: "Эпоха великих географических открытий"
slug: "age-of-discovery"
description: "Мореплаватели, экспедиции и открытия новых земель"
sparql: ""

concepts:
  - history
  - europe
  - americas
  - renaissance

---
# fall-of-ussr.yaml
name: "Падение Берлинской стены и распад СССР"
slug: "fall-of-ussr"
description: "Ключевые события конца XX века в Восточной Европе"
sparql: ""

concepts:
  - history
  - germany
  - ussr
  - russia
  - 20th_century

# ==========================================
# КАТЕГОРИЯ 3: Естественные науки
# ==========================================

---
# periodic-table-elements.yaml
name: "Химические элементы"
slug: "periodic-table-elements"
description: "Свойства химических элементов периодической таблицы"
sparql: ""

concepts:
  - chemistry
  - global
  - global_time

---
# exoplanets-discovery.yaml
name: "Открытие экзопланет"
slug: "exoplanets-discovery"
description: "Экзопланеты, их типы и космические телескопы"
sparql: ""

concepts:
  - astronomy
  - space
  - 21st_century

---
# north-american-dinosaurs.yaml
name: "Динозавры Северной Америки"
slug: "north-american-dinosaurs"
description: "Виды динозавров, обитавших на территории Северной Америки"
sparql: ""

concepts:
  - biology
  - usa
  - prehistoric

---
# nobel-physics.yaml
name: "Нобелевские лауреаты по физике"
slug: "nobel-physics"
description: "Великие физики и их открытия, удостоенные Нобелевской премии"
sparql: ""

concepts:
  - physics
  - global
  - 20th_century
  - 21st_century

---
# human-anatomy.yaml
name: "Анатомия человека"
slug: "human-anatomy"
description: "Органы, системы и строение человеческого тела"
sparql: ""

concepts:
  - biology
  - global
  - global_time

---
# infectious-diseases.yaml
name: "Инфекционные заболевания и их возбудители"
slug: "infectious-diseases"
description: "Болезни, вирусы, бактерии и способы передачи"
sparql: ""

concepts:
  - medicine
  - global
  - global_time

---
# history-of-vaccines.yaml
name: "История создания вакцин"
slug: "history-of-vaccines"
description: "Ученые-иммунологи и болезни, которые они победили"
sparql: ""

concepts:
  - medicine
  - global
  - 19th_century
  - 20th_century

---
# solar-system-objects.yaml
name: "Объекты Солнечной системы"
slug: "solar-system-objects"
description: "Планеты, их спутники и другие объекты Солнечной системы"
sparql: ""

concepts:
  - astronomy
  - space
  - global_time

---
# hominid-evolution.yaml
name: "Эволюция видов (Гоминиды)"
slug: "hominid-evolution"
description: "Предки человека и эволюционные ветви гоминид"
sparql: ""

concepts:
  - biology
  - global
  - prehistoric

---
# ancient-physicists.yaml
name: "Великие физики античности и средневековья"
slug: "ancient-physicists"
description: "Открытия и труды ранних исследователей природы"
sparql: ""

concepts:
  - physics
  - europe
  - asia
  - antiquity
  - middle_ages

# ==========================================
# КАТЕГОРИЯ 4: IT, Технологии и Математика
# ==========================================

---
# programming-languages-evolution.yaml
name: "Эволюция языков программирования"
slug: "programming-languages-evolution"
description: "Создатели языков, парадигмы и влияние друг на друга"
sparql: ""

concepts:
  - programming
  - global
  - 20th_century
  - 21st_century

---
# history-of-internet.yaml
name: "История Интернета и протоколы"
slug: "history-of-internet"
description: "Развитие сетей от ARPANET до современной паутины"
sparql: ""

concepts:
  - computer_networks
  - usa
  - europe
  - 20th_century

---
# great-math-theorems.yaml
name: "Великие теоремы математики"
slug: "great-math-theorems"
description: "Знаменитые теоремы, их авторы и суть доказательств"
sparql: ""

concepts:
  - math
  - global
  - antiquity
  - 19th_century
  - 20th_century

---
# os-creators.yaml
name: "Создатели операционных систем"
slug: "os-creators"
description: "История создания UNIX, Linux, Windows и других ОС"
sparql: ""

concepts:
  - it
  - usa
  - europe
  - 20th_century

---
# ussr-space-program.yaml
name: "Космическая программа СССР"
slug: "ussr-space-program"
description: "Первые спутники, миссии и космонавты Советского Союза"
sparql: ""

concepts:
  - it
  - astronomy
  - ussr
  - space
  - 20th_century

---
# japanese-video-games.yaml
name: "Японская индустрия видеоигр"
slug: "japanese-video-games"
description: "Игровые консоли, легендарные студии и хиты из Японии"
sparql: ""

concepts:
  - programming
  - art
  - japan
  - 20th_century
  - 21st_century

---
# cryptography-history.yaml
name: "Алгоритмы шифрования и криптография"
slug: "cryptography-history"
description: "Развитие шифров от Энигмы до алгоритма RSA"
sparql: ""

concepts:
  - math
  - it
  - global
  - 20th_century

---
# cpu-architecture.yaml
name: "Архитектура процессоров"
slug: "cpu-architecture"
description: "Развитие процессорных архитектур (x86, ARM) и их создатели"
sparql: ""

concepts:
  - it
  - global
  - 20th_century
  - 21st_century

---
# ai-development.yaml
name: "Развитие искусственного интеллекта"
slug: "ai-development"
description: "Современные модели ИИ, компании-разработчики и прорывы"
sparql: ""

concepts:
  - programming
  - global
  - 21st_century

---
# ancient-greek-mathematicians.yaml
name: "Математики Древней Греции"
slug: "ancient-greek-mathematicians"
description: "Труды Евклида, Пифагора, Архимеда и других ученых"
sparql: ""

concepts:
  - math
  - greece
  - antiquity

# ==========================================
# КАТЕГОРИЯ 5: Кросс-доменные (Междисциплинарные)
# ==========================================

---
# russian-nobel-literature.yaml
name: "Нобелевские премии по литературе (Русские писатели)"
slug: "russian-nobel-literature"
description: "Русские авторы, получившие Нобелевскую премию, и их произведения"
sparql: ""

concepts:
  - literature
  - russia
  - ussr
  - 20th_century

---
# soviet-constructivism.yaml
name: "Архитектура конструктивизма в Советском Союзе"
slug: "soviet-constructivism"
description: "Памятники конструктивизма, архитекторы и здания в СССР"
sparql: ""

concepts:
  - architecture
  - ussr
  - 20th_century

---
# scientific-revolution.yaml
name: "Научная революция 17-18 веков"
slug: "scientific-revolution"
description: "Достижения Ньютона, Лейбница и становление современной науки"
sparql: ""

concepts:
  - physics
  - math
  - history
  - europe
  - 18th_century

---
# ww2-military-tech.yaml
name: "Оружие и военная техника Второй мировой войны"
slug: "ww2-military-tech"
description: "Танки, самолеты и инженерия эпохи Второй мировой"
sparql: ""

concepts:
  - history
  - it
  - global
  - 20th_century

---
# soviet-medicine-achievements.yaml
name: "Достижения советской медицины"
slug: "soviet-medicine-achievements"
description: "Выдающиеся врачи, институты и открытия медицины в СССР"
sparql: ""

concepts:
  - medicine
  - ussr
  - 20th_century

---
# space-race-apollo-soyuz.yaml
name: "Космическая гонка: Аполлон vs Союз"
slug: "space-race-apollo-soyuz"
description: "Противостояние космических программ США и СССР"
sparql: ""

concepts:
  - history
  - astronomy
  - usa
  - ussr
  - space
  - 20th_century

---
# ancient-egypt-astronomy.yaml
name: "Астрономические знания древних египтян"
slug: "ancient-egypt-astronomy"
description: "Календари, наблюдения за звездами и наука Древнего Египта"
sparql: ""

concepts:
  - astronomy
  - egypt
  - antiquity

---
# renaissance-literature.yaml
name: "Литература эпохи Возрождения"
slug: "renaissance-literature"
description: "Творчество Данте, Петрарки, Шекспира и других авторов"
sparql: ""

concepts:
  - literature
  - europe
  - italy
  - uk
  - renaissance

---
# chemical-weapons-history.yaml
name: "История создания химического оружия и антидотов"
slug: "chemical-weapons-history"
description: "Разработка боевых отравляющих веществ и развитие токсикологии"
sparql: ""

concepts:
  - chemistry
  - medicine
  - history
  - global
  - 20th_century

---
# ancient-rome-art.yaml
name: "Монументальное искусство Древнего Рима"
slug: "ancient-rome-art"
description: "Скульптура, фрески и архитектурные памятники Римской Империи"
sparql: ""

concepts:
  - architecture
  - art
  - italy
  - antiquity
EOF

echo "Генерация завершена! Проверь папку $TARGET_DIR/"