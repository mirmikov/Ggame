package questions

import "ggame/backend/internal/models"

var bank = []models.Question{
	{ID: "q9-1", Grade: 9, Topic: "Алгоритмы", Text: "Алгоритм выполняет: x=1; для i от 1 до 4: x=x*2+i. Чему равен x?", Options: []string{"31", "42", "38", "26"}, CorrectAnswer: 1, Explanation: "Последовательно получаем 3, 8, 19 и 42.", TimeLimitSeconds: 60, Difficulty: 2},
	{ID: "q9-2", Grade: 9, Topic: "Логика", Text: "Высказывание A истинно, B ложно. Какое выражение ложно?", Options: []string{"A ИЛИ B", "НЕ B", "A И B", "A И НЕ B"}, CorrectAnswer: 2, Explanation: "Конъюнкция A И B ложна, потому что B ложно.", TimeLimitSeconds: 45, Difficulty: 1},
	{ID: "q9-3", Grade: 9, Topic: "Системы счисления", Text: "Какое десятичное число соответствует двоичному 101101?", Options: []string{"43", "45", "46", "53"}, CorrectAnswer: 1, Explanation: "32+8+4+1=45.", TimeLimitSeconds: 45, Difficulty: 1},
	{ID: "q9-4", Grade: 9, Topic: "Графы", Text: "В полном графе 6 вершин. Сколько в нём рёбер?", Options: []string{"12", "15", "18", "30"}, CorrectAnswer: 1, Explanation: "В полном графе n(n−1)/2 рёбер: 6·5/2=15.", TimeLimitSeconds: 60, Difficulty: 2},
	{ID: "q9-5", Grade: 9, Topic: "Кодирование", Text: "Сообщение состоит из 80 символов алфавита мощностью 16. Каков минимальный объём в битах?", Options: []string{"160", "240", "320", "640"}, CorrectAnswer: 2, Explanation: "На символ нужно log2(16)=4 бита, всего 80·4=320.", TimeLimitSeconds: 60, Difficulty: 2},
	{ID: "q9-6", Grade: 9, Topic: "Трассировка", Text: "Список [3,1,4,1,5]. После одного полного прохода пузырьковой сортировки по возрастанию каким станет последний элемент?", Options: []string{"1", "3", "4", "5"}, CorrectAnswer: 3, Explanation: "За один проход максимальный элемент 5 всплывает в конец.", TimeLimitSeconds: 50, Difficulty: 2},

	{ID: "q10-1", Grade: 10, Topic: "Сложность", Text: "Два вложенных цикла выполняются n и n/2 раз. Какая асимптотика числа операций?", Options: []string{"O(n)", "O(log n)", "O(n²)", "O(2ⁿ)"}, CorrectAnswer: 2, Explanation: "n·(n/2)=n²/2, постоянный множитель не влияет на O(n²).", TimeLimitSeconds: 60, Difficulty: 2},
	{ID: "q10-2", Grade: 10, Topic: "Поиск", Text: "В отсортированном массиве 1024 элементов сколько сравнений в худшем случае нужно бинарному поиску?", Options: []string{"10", "11", "512", "1024"}, CorrectAnswer: 1, Explanation: "Нужно до floor(log2(1024))+1 = 11 сравнений.", TimeLimitSeconds: 60, Difficulty: 2},
	{ID: "q10-3", Grade: 10, Topic: "SQL", Text: "Какой запрос вернёт число учеников в каждом классе?", Options: []string{"SELECT class, COUNT(*) FROM students GROUP BY class", "SELECT COUNT(class) FROM students", "SELECT class FROM students ORDER BY class", "SELECT DISTINCT COUNT(*) FROM students"}, CorrectAnswer: 0, Explanation: "Для подсчёта по группам нужны COUNT(*) и GROUP BY class.", TimeLimitSeconds: 70, Difficulty: 2},
	{ID: "q10-4", Grade: 10, Topic: "Сети", Text: "Устройство имеет адрес 192.168.5.130/26. Какой адрес сети?", Options: []string{"192.168.5.0", "192.168.5.64", "192.168.5.128", "192.168.5.192"}, CorrectAnswer: 2, Explanation: "При /26 блоки по 64 адреса; 130 попадает в диапазон 128–191.", TimeLimitSeconds: 80, Difficulty: 3},
	{ID: "q10-5", Grade: 10, Topic: "Рекурсия", Text: "Функция f(n): если n=0 вернуть 0, иначе вернуть n+f(n-1). Чему равно f(5)?", Options: []string{"10", "15", "20", "25"}, CorrectAnswer: 1, Explanation: "Это сумма 1+2+3+4+5=15.", TimeLimitSeconds: 50, Difficulty: 1},
	{ID: "q10-6", Grade: 10, Topic: "Безопасность", Text: "Какой вариант лучше всего защищает пароль, если база утечёт?", Options: []string{"Хранить открытым текстом", "Шифровать одним общим ключом", "Хранить salted hash медленной функцией", "Кодировать Base64"}, CorrectAnswer: 2, Explanation: "Salted hash с Argon2/bcrypt/scrypt затрудняет перебор и радужные таблицы.", TimeLimitSeconds: 60, Difficulty: 2},

	{ID: "q11-1", Grade: 11, Topic: "Динамическое программирование", Text: "Сколькими способами можно подняться на 5 ступеней, делая шаг на 1 или 2 ступени?", Options: []string{"5", "8", "10", "13"}, CorrectAnswer: 1, Explanation: "Количество способов образует последовательность Фибоначчи: 1,2,3,5,8.", TimeLimitSeconds: 75, Difficulty: 2},
	{ID: "q11-2", Grade: 11, Topic: "Графы", Text: "Для поиска кратчайших путей из одной вершины в графе с неотрицательными весами обычно применяют...", Options: []string{"DFS", "Дейкстру", "Флойда только", "Топологическую сортировку"}, CorrectAnswer: 1, Explanation: "Алгоритм Дейкстры решает задачу при неотрицательных весах.", TimeLimitSeconds: 50, Difficulty: 1},
	{ID: "q11-3", Grade: 11, Topic: "Параллелизм", Text: "Два потока одновременно увеличивают общий счётчик без блокировки. Какая ошибка наиболее вероятна?", Options: []string{"SQL-инъекция", "Гонка данных", "Утечка DNS", "Переполнение стека обязательно"}, CorrectAnswer: 1, Explanation: "Операция чтение-изменение-запись не атомарна и приводит к data race.", TimeLimitSeconds: 50, Difficulty: 2},
	{ID: "q11-4", Grade: 11, Topic: "Базы данных", Text: "Таблица находится в 3НФ. Какое утверждение наиболее близко к определению?", Options: []string{"Нет повторяющихся строк", "Каждый неключевой атрибут зависит только от ключа и не зависит транзитивно", "Все поля числовые", "У таблицы один столбец"}, CorrectAnswer: 1, Explanation: "3НФ устраняет транзитивные зависимости неключевых атрибутов от ключа.", TimeLimitSeconds: 75, Difficulty: 3},
	{ID: "q11-5", Grade: 11, Topic: "Алгоритмы", Text: "Какой результат даст выражение sorted({3,1,2,3,2}) в Python?", Options: []string{"[3,1,2,3,2]", "[1,2,3]", "{1,2,3}", "[1,2,2,3,3]"}, CorrectAnswer: 1, Explanation: "Множество удаляет повторы, sorted возвращает отсортированный список.", TimeLimitSeconds: 45, Difficulty: 1},
	{ID: "q11-6", Grade: 11, Topic: "Вероятность", Text: "Хеш-функция равномерно распределяет ключи по 100 ячейкам. Какова вероятность попадания одного ключа в конкретную ячейку?", Options: []string{"1/10", "1/50", "1/100", "1/10000"}, CorrectAnswer: 2, Explanation: "При равномерном распределении каждая из 100 ячеек имеет вероятность 1/100.", TimeLimitSeconds: 45, Difficulty: 1},
}

var tasks = []models.TerminalTask{
	{ID: "py-1", Title: "Чётные числа", Description: "Напишите Python-выражение, которое возвращает список чётных чисел из nums.", Language: "python", StarterCode: "nums = [1, 2, 3, 4, 5, 6]\nresult = ...", AcceptedAnswers: []string{"[x for x in nums if x % 2 == 0]", "[x for x in nums if x%2==0]"}, Reward: 260, Difficulty: 2},
	{ID: "py-2", Title: "Проверка палиндрома", Description: "Допишите выражение, которое проверяет, является ли строка s палиндромом.", Language: "python", StarterCode: "s = input().strip()\nis_palindrome = ...", AcceptedAnswers: []string{"s == s[::-1]", "s==s[::-1]"}, Reward: 220, Difficulty: 1},
	{ID: "py-3", Title: "Частоты", Description: "Напишите одну строку Python, создающую словарь частот символов строки s.", Language: "python", StarterCode: "s = 'abracadabra'\nfreq = ...", AcceptedAnswers: []string{"{c: s.count(c) for c in set(s)}", "{c:s.count(c) for c in set(s)}"}, Reward: 320, Difficulty: 3},
	{ID: "sql-1", Title: "Лучшие результаты", Description: "Выведите name и score трёх игроков с максимальным score из таблицы players.", Language: "sql", StarterCode: "SELECT ...", AcceptedAnswers: []string{"SELECT name, score FROM players ORDER BY score DESC LIMIT 3", "select name,score from players order by score desc limit 3"}, Reward: 260, Difficulty: 2},
	{ID: "sql-2", Title: "Средний балл класса", Description: "Выведите class и средний score для каждого класса из таблицы results.", Language: "sql", StarterCode: "SELECT ...", AcceptedAnswers: []string{"SELECT class, AVG(score) FROM results GROUP BY class", "select class,avg(score) from results group by class"}, Reward: 300, Difficulty: 2},
	{ID: "py-4", Title: "Бинарный поиск", Description: "Какое условие цикла используется в классическом бинарном поиске по индексам left и right?", Language: "python", StarterCode: "while ...:\n    mid = (left + right) // 2", AcceptedAnswers: []string{"left <= right", "left<=right"}, Reward: 180, Difficulty: 1},
}

func All() []models.Question { return append([]models.Question(nil), bank...) }
func ForGrade(grade int) []models.Question {
	if grade < 9 || grade > 11 {
		return All()
	}
	out := make([]models.Question, 0)
	for _, q := range bank {
		if q.Grade == grade {
			out = append(out, q)
		}
	}
	return out
}
func Tasks() []models.TerminalTask { return append([]models.TerminalTask(nil), tasks...) }
