## Веб - сервис умного воображаемого калькулятора
### Выполнено по заданию Яндекс

#### В задании реализована back-end часть

Пользователь хочет считать арифметические выражения. Он вводит строку 2 + 2 * 2 и хочет получить в ответ 6. Но наши
операции сложения и умножения (также деления и вычитания) выполняются "очень-очень" долго. Поэтому вариант, при
котором пользователь делает Нир-запроси получает в качестве ответа результат, невозможна. Более того: вычисление
каждой такой операции в нашей “альтернативной реальности" занимает "гигантские" вычислительные мощности.
Соответственно, каждое действие мы должны уметь выполнять отдельно и масштабировать эту систему можем
добавлением вычислительных мощностей в нашу систему в виде новых "машин". Поэтому пользователь, поэтому
выражение, получает в ответ идентификатор выражения и может с какой-то периодичностью уточнять у сервера "не
посчиталось ли выражение"? Если выражение наконец будет вычислено - то он получит результат. Помните, что
некоторые части арифметического выражения можно вычислять параллельно.

Требования:

* Оркестратор может перезапускаться без потери состояния, Все выражения храним в СУБД.

* Оркестратор должен отслеживать задачи, которые выполняются слишком долго (вычислитель тоже может уйти
  о связи) и делать их повторно доступными для вычислений.

Back-end часть

Состоит из 2 элементов: 

* Сервер, который принимает арифметическое выражение, переводит его в набор последовательных задачи
  обеспечивает порядок их выполнения. Далее будем называть его оркестратором.

* Вычислитель, который может получить от оркестратора задачу, выполнить его и вернуть серверу результат,
  Далее будем называть его агентом.

Оркестратор

Сервер, который имеет следующие endpoint-ы:

* Добавление вычисления арифметического выражения.

* Получение списка выражений со статусами.

* Получение значения выражения по его идентификатору.

* Получение списка доступных операций со временем их выполнения.
* Получение задачи для выполнения.

* Приём результата обработки данных.

Агент

Демон, который получает выражение для вычисления с сервера, вычисляет его и отправляет на сервер результат
выражения. При старте демон запускает несколько горутин, каждая из которых выступает в роли независимого
вычислителя. Количество горутин регулируется переменной средь.