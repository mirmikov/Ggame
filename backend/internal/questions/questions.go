package questions

import "ggame/backend/internal/models"

var seed = []models.Question{
	{ID: "q9-01", Grade: 9, Topic: "Системы счисления", Text: "Чему равно 101101₂ в десятичной системе?", Options: []string{"41", "43", "45", "47"}, CorrectAnswer: 2, Explanation: "32 + 8 + 4 + 1 = 45.", TimeLimitSeconds: 25, Difficulty: 1},
	{ID: "q9-02", Grade: 9, Topic: "Логика", Text: "Когда выражение A И B истинно?", Options: []string{"Если истинно хотя бы одно", "Только если оба истинны", "Если оба ложны", "Всегда"}, CorrectAnswer: 1, Explanation: "Конъюнкция истинна только при двух истинных операндах.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-03", Grade: 9, Topic: "Информация", Text: "Сколько бит в одном байте?", Options: []string{"4", "8", "16", "32"}, CorrectAnswer: 1, Explanation: "Байт состоит из 8 бит.", TimeLimitSeconds: 15, Difficulty: 1},
	{ID: "q9-04", Grade: 9, Topic: "Алгоритмы", Text: "Сколько раз выполнится цикл for i := 0; i < 5; i++?", Options: []string{"4", "5", "6", "Бесконечно"}, CorrectAnswer: 1, Explanation: "Значения i: 0, 1, 2, 3, 4.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-05", Grade: 9, Topic: "Кодирование", Text: "Какой минимальный объём нужен для 16 различных символов?", Options: []string{"2 бита", "4 бита", "8 бит", "16 бит"}, CorrectAnswer: 1, Explanation: "2⁴ = 16.", TimeLimitSeconds: 25, Difficulty: 2},
	{ID: "q9-06", Grade: 9, Topic: "Системы счисления", Text: "Как записать число 15 в шестнадцатеричной системе?", Options: []string{"E", "F", "10", "15"}, CorrectAnswer: 1, Explanation: "Цифре F соответствует десятичное 15.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-07", Grade: 9, Topic: "Логика", Text: "Чему равно НЕ(истина)?", Options: []string{"Истина", "Ложь", "0 и 1", "Не определено"}, CorrectAnswer: 1, Explanation: "Отрицание меняет логическое значение.", TimeLimitSeconds: 15, Difficulty: 1},
	{ID: "q9-08", Grade: 9, Topic: "Алгоритмы", Text: "Как называется поиск делением отсортированного массива пополам?", Options: []string{"Линейный", "Бинарный", "Пузырьковый", "Случайный"}, CorrectAnswer: 1, Explanation: "Это бинарный поиск.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-09", Grade: 9, Topic: "Информация", Text: "Сколько байт в 2 КиБ?", Options: []string{"1000", "1024", "2000", "2048"}, CorrectAnswer: 3, Explanation: "2 × 1024 = 2048 байт.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-10", Grade: 9, Topic: "Программирование", Text: "Что обычно хранит переменная?", Options: []string{"Команду процессора", "Значение", "Только текст", "Файл"}, CorrectAnswer: 1, Explanation: "Переменная связывает имя со значением.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-11", Grade: 9, Topic: "Системы счисления", Text: "Чему равно 1111₂ + 1₂?", Options: []string{"1110₂", "1111₂", "10000₂", "10001₂"}, CorrectAnswer: 2, Explanation: "15 + 1 = 16 = 10000₂.", TimeLimitSeconds: 25, Difficulty: 2},
	{ID: "q9-12", Grade: 9, Topic: "Логика", Text: "A ИЛИ B ложно, если...", Options: []string{"Оба ложны", "Оба истинны", "A истинно", "B истинно"}, CorrectAnswer: 0, Explanation: "Дизъюнкция ложна только при двух ложных операндах.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-13", Grade: 9, Topic: "Алгоритмы", Text: "Какой алгоритм должен завершаться за конечное число шагов?", Options: []string{"Любой корректный", "Только графический", "Только линейный", "Никакой"}, CorrectAnswer: 0, Explanation: "Конечность — базовое свойство алгоритма.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-14", Grade: 9, Topic: "Сети", Text: "Какой протокол используется для веб-страниц?", Options: []string{"HTTP", "SSH", "SMTP", "DNS"}, CorrectAnswer: 0, Explanation: "HTTP передаёт гипертекст.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-15", Grade: 9, Topic: "Информация", Text: "Как называется уменьшение размера файла?", Options: []string{"Компиляция", "Сжатие", "Шифрование", "Индексация"}, CorrectAnswer: 1, Explanation: "Сжатие уменьшает объём данных.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-16", Grade: 9, Topic: "Программирование", Text: "Что делает условный оператор?", Options: []string{"Повторяет код", "Выбирает ветвь", "Удаляет данные", "Сортирует массив"}, CorrectAnswer: 1, Explanation: "Он выбирает действие в зависимости от условия.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-17", Grade: 9, Topic: "Системы счисления", Text: "Чему равно 64₁₀ в двоичной системе?", Options: []string{"100000₂", "1000000₂", "10000000₂", "111111₂"}, CorrectAnswer: 1, Explanation: "64 = 2⁶.", TimeLimitSeconds: 25, Difficulty: 2},
	{ID: "q9-18", Grade: 9, Topic: "Логика", Text: "Импликация A → B ложна, когда...", Options: []string{"A и B ложны", "A ложно, B истинно", "A истинно, B ложно", "Оба истинны"}, CorrectAnswer: 2, Explanation: "Импликация ложна только из истины в ложь.", TimeLimitSeconds: 25, Difficulty: 2},
	{ID: "q9-19", Grade: 9, Topic: "Кодирование", Text: "Как называется таблица кодов символов латиницы?", Options: []string{"ASCII", "HTTP", "JPEG", "SQL"}, CorrectAnswer: 0, Explanation: "ASCII — одна из систем кодирования символов.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-20", Grade: 9, Topic: "Алгоритмы", Text: "Какой блок-схемный элемент обозначает условие?", Options: []string{"Прямоугольник", "Ромб", "Овал", "Стрелка"}, CorrectAnswer: 1, Explanation: "Условие обозначается ромбом.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-21", Grade: 9, Topic: "Информация", Text: "Сколько различных сообщений кодируют 3 бита?", Options: []string{"3", "6", "8", "9"}, CorrectAnswer: 2, Explanation: "2³ = 8.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-22", Grade: 9, Topic: "Программирование", Text: "Какой тип хранит истину или ложь?", Options: []string{"string", "boolean", "float", "array"}, CorrectAnswer: 1, Explanation: "Логический тип хранит два значения.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-23", Grade: 9, Topic: "Сети", Text: "Что преобразует доменное имя в IP-адрес?", Options: []string{"DNS", "HTML", "USB", "CPU"}, CorrectAnswer: 0, Explanation: "Это задача DNS.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q9-24", Grade: 9, Topic: "Алгоритмы", Text: "Сложность линейного поиска в худшем случае?", Options: []string{"O(1)", "O(log n)", "O(n)", "O(n²)"}, CorrectAnswer: 2, Explanation: "Может потребоваться проверить все n элементов.", TimeLimitSeconds: 25, Difficulty: 2},
	{ID: "q9-25", Grade: 9, Topic: "Информация", Text: "Какой формат обычно хранит изображение?", Options: []string{"PNG", "MP3", "CSV", "EXE"}, CorrectAnswer: 0, Explanation: "PNG — растровый формат изображения.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q10-01", Grade: 10, Topic: "Массивы", Text: "Индекс первого элемента массива в Go?", Options: []string{"-1", "0", "1", "Зависит от массива"}, CorrectAnswer: 1, Explanation: "Индексация начинается с нуля.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q10-02", Grade: 10, Topic: "Сложность", Text: "Сложность бинарного поиска?", Options: []string{"O(1)", "O(log n)", "O(n)", "O(n²)"}, CorrectAnswer: 1, Explanation: "Область поиска делится пополам.", TimeLimitSeconds: 25, Difficulty: 2},
	{ID: "q10-03", Grade: 10, Topic: "SQL", Text: "Какая команда читает строки таблицы?", Options: []string{"SELECT", "UPDATE", "DELETE", "DROP"}, CorrectAnswer: 0, Explanation: "SELECT выбирает данные.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q10-04", Grade: 10, Topic: "Рекурсия", Text: "Что обязательно нужно рекурсивной функции?", Options: []string{"Глобальная переменная", "Базовый случай", "Цикл", "Массив"}, CorrectAnswer: 1, Explanation: "Базовый случай останавливает рекурсию.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q10-05", Grade: 10, Topic: "JavaScript", Text: "Строгое сравнение в JavaScript?", Options: []string{"=", "==", "===", "!="}, CorrectAnswer: 2, Explanation: "=== сравнивает значение и тип.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q10-06", Grade: 10, Topic: "Go", Text: "Как объявить короткую переменную в Go?", Options: []string{"let x = 1", "x := 1", "var: x = 1", "x <- 1"}, CorrectAnswer: 1, Explanation: ":= — короткое объявление.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q10-07", Grade: 10, Topic: "Python", Text: "Что вернёт len([4, 5, 6])?", Options: []string{"2", "3", "4", "15"}, CorrectAnswer: 1, Explanation: "В списке три элемента.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q10-08", Grade: 10, Topic: "Сортировки", Text: "Худшая сложность пузырьковой сортировки?", Options: []string{"O(1)", "O(log n)", "O(n)", "O(n²)"}, CorrectAnswer: 3, Explanation: "Используются вложенные проходы.", TimeLimitSeconds: 25, Difficulty: 2},
	{ID: "q10-09", Grade: 10, Topic: "Функции", Text: "Что передаётся функции при вызове?", Options: []string{"Аргументы", "Только строки", "Классы", "Запросы"}, CorrectAnswer: 0, Explanation: "Функция принимает аргументы через параметры.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q10-10", Grade: 10, Topic: "Структуры данных", Text: "Как работает стек?", Options: []string{"FIFO", "LIFO", "Случайно", "По приоритету всегда"}, CorrectAnswer: 1, Explanation: "Последним пришёл — первым вышел.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q10-11", Grade: 11, Topic: "SQL", Text: "Какое слово фильтрует строки SELECT?", Options: []string{"WHERE", "ORDER", "GROUP", "JOIN"}, CorrectAnswer: 0, Explanation: "WHERE задаёт условие отбора.", TimeLimitSeconds: 20, Difficulty: 1},
	{ID: "q10-12", Grade: 11, Topic: "Графы", Text: "Что находит BFS в невзвешенном графе?", Options: []string{"Кратчайший путь по числу рёбер", "Минимальный вес всегда", "Цикл максимальной длины", "Только вершины-листья"}, CorrectAnswer: 0, Explanation: "BFS обходит граф слоями.", TimeLimitSeconds: 25, Difficulty: 2},
	{ID: "q10-13", Grade: 11, Topic: "Сложность", Text: "Какая структура даёт средний поиск O(1)?", Options: []string{"Связный список", "Хеш-таблица", "Обычный массив без индекса", "Стек"}, CorrectAnswer: 1, Explanation: "Хеш-таблица обычно даёт константный доступ.", TimeLimitSeconds: 25, Difficulty: 2},
	{ID: "q10-14", Grade: 11, Topic: "Go", Text: "Какой примитив запускает конкурентную функцию в Go?", Options: []string{"thread", "async", "go", "spawn"}, CorrectAnswer: 2, Explanation: "Ключевое слово go запускает goroutine.", TimeLimitSeconds: 20, Difficulty: 2},
	{ID: "q10-15", Grade: 11, Topic: "Алгоритмы", Text: "Что такое динамическое программирование?", Options: []string{"Хранение результатов подзадач", "Рисование блок-схем", "Только рекурсия", "Случайный поиск"}, CorrectAnswer: 0, Explanation: "Повторные подзадачи вычисляются один раз.", TimeLimitSeconds: 25, Difficulty: 2},
}

func All() []models.Question { return append([]models.Question(nil), seed...) }

var tasks = []models.TerminalTask{
	{ID: "t-01", Title: "Binary packet", Description: "Переведите 110101₂ в десятичную систему.", ExpectedAnswer: "53", Reward: 250},
	{ID: "t-02", Title: "Loop trace", Description: "Какой вывод даст: sum=0; for i=1..4 sum+=i?", ExpectedAnswer: "10", Reward: 250},
	{ID: "t-03", Title: "Array core", Description: "Найдите максимум массива [7, 2, 19, 4, 11].", ExpectedAnswer: "19", Reward: 250},
	{ID: "t-04", Title: "SQL uplink", Description: "Какое ключевое слово SQL сортирует результат?", ExpectedAnswer: "order by", Reward: 250},
	{ID: "t-05", Title: "Complexity gate", Description: "Укажите сложность бинарного поиска в формате O(...).", ExpectedAnswer: "o(log n)", Reward: 250},
}

func Tasks() []models.TerminalTask { return append([]models.TerminalTask(nil), tasks...) }

func ForGrade(grade int) []models.Question {
	out := make([]models.Question, 0)
	for _, q := range seed {
		if grade == 9 && q.Grade == 9 || grade >= 10 && q.Grade >= 10 {
			out = append(out, q)
		}
	}
	return out
}
