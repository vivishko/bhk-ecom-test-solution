# Задача 1. Счетчик кликов.

Есть набор баннеров (от 10 до 100). У каждого есть ИД и название (id, name)

Нужно сделать сервис, который будет считать клики и собирать их в поминутную статистику (timestamp, bannerID, count)

Нужно сделать АПИ с двумя методами:

1. /counter/:bannerID (GET)
   Должен посчитать +1 клик по баннеру с заданным ИД
2. /stats/:bannerID (POST)
   Должен выдать статистику показов по баннеру за указанный промежуток времени (tsFrom, tsTo)

Язык: golang </br>
СУБД: mongo или psql </br>
Сложность:

- junior = кол-во запросов /counter 10-50 в секунду
- middle+ = кол-во запросов /counter 100-500 в секунду

PS: тесты делать не обязательно

---

# Задача 2. Расчет эквивалента USD на заданном адресе сети Ethereum

Нужен простой сервис, который сможет пересчитать кол-во ETH на заданном адресе в USD

Внешние АПИ использовать нельзя (только on-chain). Рекомендуется использовать контракты chainlink (например https://etherscan.io/address/0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419)

Для приложения не нужен веб интерфейс. Оно получает ethereum адрес в параметрах командной строки, выводит кол-во ETH на балансе и эквивалент этой суммы в USD.

Дополнительно: учесть WETH , если таковые есть на балансе

Макс. уровень сложности: вывести все токены и их эквивалент в USD, которые поддерживает chainlink. Только если они есть на балансе адреса (https://docs.chain.link/data-feeds/price-feeds/addresses?network=ethereum&page=3&search=USD)
