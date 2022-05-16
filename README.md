Модуль для отправки и обработки запросов к VK API на Golang, включающий в себя минимум необходимых возможностей:

- Отправка запросов к VK API.
- Лимитирование скорости отправки запросов по токену в секунду.
- Поддержка настраиваемых парсеров, есть возможность настроить парсер для формата [messagepack](https://msgpack.org/), например, для работы с вызовом метода [users.get.msgpack](https://api.vk.com/method/users.get.msgpack)
- Поддержка middleware для выполнения запроса (как на этапе отправки http запроса, так и генерации запроса vk api)
- Поддержка [контекстов](https://pkg.go.dev/context) для настройки таймаутов на выполнение запроса или передачи значений в него.
- Возможность изменять `http.Client` для управления соединениями (настройка прокси, лимитирование количества соединений и т.д).
- Каждый запрос к VK API содержит всю необходимую информацию для его выполнения: метод, параметры, заголовки и url
- Выполнение запросов можно блокировать в middleware
- Запросы можно переиспользовать и изменять в любой момент для экономии памяти. Модуль `executor` гарантирует закрытие соединения после выполнения запроса.
- Возможность ловить результат выполнения методов и переотправлять запрос снова (например, в случае ошибки капчи)

Модуль позиционируется как простейший компонент, на который не должно возлагаться обязательств по управлению сессиями, работе с LongPoll соединениями или поддержке Streaming API, Chat Bot API и т.д. 

Так же, подуль не позиционируется как SDK, но на его основе таковой сделать можно.