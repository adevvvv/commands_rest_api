## Инструкция

---
Клонируй проект:
```
git clone https://github.com/adevvvv/commands_rest_api
```
Открой проект в терминале и запусти команду:
```
docker-compose up --build
```
Открой postman для просмотра возможностей и тестирования проекта:
```
POST 
http://localhost:8000/commands
http://localhost:8000/commands/1/stop

GET
http://localhost:8000/commands
http://localhost:8000/commands/1

{
    "command": "echo Hello"
}
```
---
### Реализовано:
- [x]  Создание новой команды. Запускает переданную bash-команду, сохраняет результат выполнения в БД;
- [x]  Получение списка команд;
- [x]  Получение одной команды;
- [x]  Метод для остановки команды.
