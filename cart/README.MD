# Запуск проекта

### 1. Клонировать репозиторий
```bash   
git clone https://gitlab.ozon.dev/maksimovis74/homework.git   
```
### 2. Перейти в директорию проекта (если Вы не в ней).  
```bash  
cd homework  
```  
### 3. Из дериктории проекта выполнить команды, сервис будет доступен по порту 8080:  
```bash      
make run-all   
```  
## 🛠Реализация   
- Язык программирование: Golang 1.23.1
- Для реализации http сервера использовалась библиотека go
- Логер zerolog  

# ✅ Условия  
Уменьшить время ответа cart.list.
1. Распараллельные вызовы ручки продакт сервиса ```cart/internal/business/service/cart/get_items.go``` ✅
2. Самим написать аналог https://pkg.go.dev/golang.org/x/sync/errgroup и использовать его ```cart/internal/infra/errgroup/errgroup.go``` ✅
3. В случае ошибки - отменять все текущие запросы и вернуть ошибку из errgroup ✅
4. При общении с Product Service необходимо использовать лимит 10 RPS на клиентской стороне (реализован по средствам https://pkg.go.dev/golang.org/x/time/rate) ✅
5. Группа живет в рамках одного запроса, группа не переиспользуется между запросами ✅
6. in-memory репозитории защитить мьютексами ✅

# ✅ Дополнительное задание  
1. Тесты на многопоточность in-memory репозитория ✅
2. Грейсфул завершение приложения, использовать signal.Notify (реализован через closer - `cart/internal/infra/closer/closer.go`) ✅
3. Контроль утечек горутин через https://github.com/uber-go/goleak (проверка добавлена в тесты с многопоточностью) ✅
4. Контроль рейзов через тесты с флагом -race `make/service_make.mk` ✅
5. Реализовать параллельный запуск тестов (тесты запускаются параллельно, установлены линтеры paralleltest, tparallel) ✅
 # Примеры ответов на запросы согласно сценарию

```json
POST http://localhost:8080/user/31337/cart/1076963
Content-Type: application/json

{
  "count": 1
}

### expected {} 200 OK; must add 1 item
```
```json
POST http://localhost:8080/user/31337/cart/1076963
Content-Type: application/json

{
  "count": 5
}
### expected {} 200 OK; must add 5 more item, 1076963 - must be 6 items

### result GET http://localhost:8080/user/31337/cart
{
    "items": [
        {
            "sku": 1076963,
            "name": "Теория нравственных чувств | Смит Адам",
            "count": 6,
            "price": 3379
        }
    ],
    "total_price": 20274
}
```
```json
### add unknown sku to cart
POST http://localhost:8080/user/31337/cart/1076963000
Content-Type: application/json

{
  "count": 1
}
### expected {} 412 Precondition Failed; invalid sku

### result:
{
    "Message": "SKU должен существовать в сервисе"
}
```
```json
### add another sku to cart
POST http://localhost:8080/user/31337/cart/1148162
Content-Type: application/json

{
  "count": 1
}
### expected {} 200 OK; must add 1 item

### result GET http://localhost:8080/user/31337/cart
{
    "items": [
        {
            "sku": 1148162,
            "name": "Кулинар Гуров",
            "count": 1,
            "price": 2931
        },
        {
            "sku": 1076963,
            "name": "Теория нравственных чувств | Смит Адам",
            "count": 6,
            "price": 3379
        }
    ],
    "total_price": 23205
}
```
```json
### invalid user
POST http://localhost:8080/user/0/cart/1148162
Content-Type: application/json

{
  "count": 1
}
### expected {} 400 Bad Request

### result:
{
    "Message": "идентификатор пользователя должен быть натуральным числом (больше нуля)"
}
```
```json
### invalid sku
POST http://localhost:8080/user/31337/cart/0
Content-Type: application/json

{
  "count": 1
}
### expected {} 400 Bad Request

### result:
{
    "Message": "SKU должен быть натуральным числом (больше нуля)"
}
```
```json
### invalid count
POST http://localhost:8080/user/31337/cart/1148162
Content-Type: application/json

{
  "count": 0
}
### expected {} 400 Bad Request

###result:
{
    "Message": "validation error: поле 'Count' должно быть больше 0"
}
```
```json
### delete whole sku from cart
DELETE http://localhost:8080/user/31337/cart/1076963
Content-Type: application/json
### expected {} 200 OK; must delete item from cart

### result GET http://localhost:8080/user/31337/cart
{
    "items": [
        {
            "sku": 1148162,
            "name": "Кулинар Гуров",
            "count": 1,
            "price": 2931
        }
    ],
    "total_price": 2931
}
```
```json
### delete whole cart
DELETE http://localhost:8080/user/31337/cart
Content-Type: application/json
### expected {} 204 No Content; must delete cart

### result GET http://localhost:8080/user/31337
404 Not Found
```
```json
### get invalid list of cart
GET http://localhost:8080/user/0/cart
Content-Type: application/json
### 400 bad request

### result:
{
    "Message": "идентификатор пользователя должен быть натуральным числом (больше нуля)"
}
```