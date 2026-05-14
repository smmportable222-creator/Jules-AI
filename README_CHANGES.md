# Pavuk Crawler Updates

## Overview of Changes / Обзор изменений

This document outlines the major overhauls and enhancements deployed to `pavuk5_refactored` per the recent requests. The primary goals were increasing target accuracy, bypassing modern anti-bot protections, improving UI accessibility, and ensuring output categorization is precise and flexible.

Этот документ описывает основные переработки и улучшения, примененные к `pavuk5_refactored` согласно последним запросам. Основными целями были увеличение точности поиска целевых страниц, обход современных средств защиты от ботов, улучшение доступности интерфейса и обеспечение точной и гибкой категоризации результатов.

### 1. Web GUI (User-Friendly Interface) / Веб-интерфейс
- **EN:** A built-in web interface has been added. Running the crawler with the `-web` flag starts a server on `http://localhost:8080`. Users can insert domains, proxies, and crawler parameters directly into a browser form and monitor the log feed dynamically.
- **RU:** Добавлен встроенный веб-интерфейс. Запуск краулера с флагом `-web` запускает сервер на `http://localhost:8080`. Пользователи могут вставлять домены, прокси и параметры краулера прямо в форму в браузере и динамически отслеживать логи.

### 2. Captcha & Anti-bot Bypass / Обход капчи и антибот систем
- **EN:** The standard HTTP client has been replaced with `bogdanfinn/tls-client` specifically designed to spoof TLS handshakes and avoid Cloudflare 403s.
- **EN:** Client profiles randomly rotate between Chrome, Firefox, Safari, and Opera to maintain indistinguishable headers.
- **EN:** Hardcoded randomized wait times (500-2500ms) have been implemented, as well as specific 403 / 429 response catching to halt and retry over different profiles/proxies properly.
- **RU:** Стандартный HTTP-клиент был заменен на `bogdanfinn/tls-client`, специально разработанный для подделки TLS-рукопожатий и избегания блокировок Cloudflare 403.
- **RU:** Профили клиента случайным образом меняются между Chrome, Firefox, Safari и Opera, чтобы сохранять неразличимые заголовки.
- **RU:** Внедрены рандомизированные задержки (500-2500 мс) для имитации поведения человека. Включен перехват ответов 403 / 429 для приостановки и повтора через другие профили/прокси.

### 3. Categorization & External Patterns / Категоризация и внешние паттерны
- **EN:** Instead of abstract "probability tier files", results are directly saved to `1_comments.jsonl`, `2_emails.jsonl`, `3_forums.jsonl`, `4_contacts.jsonl`, and `5_errors.jsonl`.
- **EN:** All heuristic rules (RegEx for comments, URLs, form field logic) have been externalized to `patterns.json`. This JSON dynamically loads at runtime without needing recompilation.
- **EN:** This logic prevents "absolute drops" (where a page is discarded due to 1 missing factor). "Soft boosts" are now used.
- **RU:** Вместо абстрактных файлов с "уровнями вероятности", результаты напрямую сохраняются в `1_comments.jsonl`, `2_emails.jsonl`, `3_forums.jsonl`, `4_contacts.jsonl` и `5_errors.jsonl`.
- **RU:** Все эвристические правила (RegEx для комментариев, URL, логика полей форм) вынесены в файл `patterns.json`. Этот JSON динамически загружается во время выполнения без необходимости перекомпиляции.
- **RU:** Эта логика предотвращает "абсолютные отказы" (когда страница отбрасывается из-за отсутствия 1 фактора). Теперь используются "мягкие бусты" (soft boosts).

### 4. Multilingual & Funnel Depth / Мультиязычность и глубина воронки
- **EN:** `patterns.json` now includes keyword detection for Spanish, French, and German for comment semantics, forms, and subjects to expand global domain coverage.
- **EN:** URLs with deep pagination (`/page/`) are now given a massive heuristic forgiveness score to prioritize probing actual fresh pagination routes instead of penalizing them as aggregators. A massive boost is given to dates within the current 12 months.
- **RU:** В `patterns.json` теперь добавлено распознавание ключевых слов на испанском, французском и немецком языках для семантики комментариев, форм и тем, что расширяет охват глобальных доменов.
- **RU:** Ссылки с глубокой пагинацией (`/page/`) теперь получают масштабное "прощение" в скоринге эвристики: приоритет отдается прохождению реальных свежих маршрутов пагинации, а не штрафованию их как агрегаторов. Огромный бонус получают даты в пределах текущих 12 месяцев.

### 5. Target Limiting / Лимиты на целевые страницы
- **EN:** The crawler intelligently tracks secondary indicators (emails, contact forms, errors) but doesn't count them against the domain threshold limit. It stops purely when it finds 10 high-quality target pages capable of placements (Comments or Forums).
- **RU:** Краулер интеллектуально отслеживает вторичные индикаторы (email-адреса, контактные формы, ошибки), но не учитывает их в лимите порога домена. Он останавливается строго после нахождения 10 целевых страниц высокого качества, пригодных для размещения (Комментарии или Форумы).

### 6. Additional Bug Fixes (User Request)
- **BUG 1:** Profile rotation is now functioning perfectly. The crawler rotates profiles randomly (Chrome, Firefox, Safari, Opera) upon initialization and dynamically updates/swaps the profile specifically whenever a 403 or 429 response is encountered, bypassing Captchas.
- **BUG 2:** The target counters strictly only increment and respect stopping boundaries when `comment` or `forum` placement possibilities are found, allowing unhindered collection of emails and errors until 10 target pages are acquired.
- **BUG 3:** `log.Printf` inside `DomainCrawler` mechanisms was replaced with `AddLog(fmt.Sprintf(...))` to natively and correctly pipe the outputs into the Web GUI terminal in real-time.
- **NEW FEATURE:** Introduced a `fetchFreshURLs` queue that scrapes `sitemap_news.xml`, `/feed/`, and `/rss/` links up-front, aggressively appending 20 fresh endpoints to the front of the queue to analyze recent activity and guarantee current material coverage.

### 6. Дополнительные исправления ошибок (по запросу пользователя)
- **ОШИБКА 1:** Ротация профилей теперь работает идеально. Краулер случайным образом меняет профили браузеров (Chrome, Firefox, Safari, Opera) при инициализации и динамически обновляет/заменяет профиль каждый раз, когда встречается ответ 403 или 429, тем самым обходя капчи.
- **ОШИБКА 2:** Счетчики целей строго увеличиваются и учитывают границы остановки только тогда, когда найдены возможности для размещения комментариев (`comment`) или форумов (`forum`), что позволяет беспрепятственно собирать электронные письма и ошибки до тех пор, пока не будет получено 10 целевых страниц.
- **ОШИБКА 3:** `log.Printf` внутри механизмов `DomainCrawler` был заменен на `AddLog(fmt.Sprintf(...))` для корректной передачи вывода в терминал Web GUI в режиме реального времени.
- **НОВАЯ ФУНКЦИЯ:** Введена очередь `fetchFreshURLs`, которая заранее извлекает ссылки из `sitemap_news.xml`, `/feed/` и `/rss/`, агрессивно добавляя 20 свежих конечных точек в начало очереди для анализа недавней активности и обеспечения охвата актуального материала.
