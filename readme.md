# Zip Loader
## 1.  Описание программы
Программа вызывает таблицу gar.gar_stats и смотрит текущую версию гара. Если имеется новая версия или же не имеется совсем, то начинается установка файла output.zip. После установки в таблицу gar.gar_stats добавляется актульная версия файла.
## 2. Установка GO

[Установить можно здесь](https://go.dev/?target=_blank)
## 3. Установка окружения и запуск программы 
1. Установка модулей
      #### ```$ go get "github.com/lib/pq"```
      #### ```$ go get "github.com/jmoiron/sqlx"```
      #### ```$ go get "gopkg.in/yaml.v2"```
     #### Далее в ```config.yaml``` нужно прописать данные о подключение (конфиг должен находиться в одной папке с ```main.go```) 

2. Запуск программы
      ####```$ go run main.go```
      ####Если нужен бинарник
      ####```$ go run main.go```
