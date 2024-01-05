# Sensor Streaming (Microservice A)
### version: 1.0.0
Sensor streming or microservice A is service that collect and generate data stream to other microservices

## Code Architecture
This service made to containerize the allcare core service make a clear separation
of concern within the architecture with [Uncle Bob Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) interpretation.

## Build with

- [Go 1.20](https://adoptopenjdk.net/)
- [Echo](https://github.com/labstack/echo)
- [Swagger](https://github.com/swaggo/echo-swagger)
- [Sqlx](https://github.com/jmoiron/sqlx)
- [Goland IDEA](https://www.jetbrains.com/idea)
- [Squirrel](https://github.com/Masterminds/squirrel)
- [Mqtt](https://github.com/eclipse/paho.mqtt.golang)
- [sql-migrate](https://github.com/rubenv/sql-migrate)

## Dependency Rule
Based on Uncle Bob, there are 4 layers:
* Entity
* Use Cases
* Interface Adapters
* Frameworks and Drivers.

In this repository, we also using 4 layers (with modification) like this:
* Entity.
* Use Cases (Implementation).
* Interface Adapters. Will be splitted into two:
    * Repository Interface. Bridging repository implementation and use cases layer.
    * Use Case Interface. Bridging handler and use cases layer.
* Driver Layer
    * Handler
    * Repository implementation

## Standardization used to build this project

- [Google Golang Format](https://google.github.io/styleguide/go) used for code style formatting
- [Sonarlint](https://www.sonarlint.org/) used to check code cleanness and help prevent NPE
- [Resource Naming Guide](https://restfulapi.net/resource-naming/) used to make a clear resource naming

## Guidelines
### Step-by-step writing code using this pattern
- Setup skeleton of the microservices (including: main.go, cmd, server, migrations, config, platform. handler)
- Defining the `entity`
- Defining `usecase` (interface & implementation) in folder `module/sensor/useaces`. We're gonna focus in this folder since the business logic will be written here.
- When the usecase need to communicate to the external agency (Database) then write it to the `repository interface`
- After the usecase layer was done, now time to write `repository implementation`
- Put it up together + register to the module/sensor/service.go and init module dependency in server.go

## Environment

### Environment App
  Please setup the backend environment in this file [app.yaml](config%2Fapp.yaml).

    $ sql-migrate --help
    app:
      port: 8081
      name: sensor_streaming
      secret_key: "oksecret!1"
    mysql:
      host: host.docker.internal
      port: 3306
      driver_name: mysql
      username: cabin
      password: cabinpw
      charset: utf8
      db_name: sensor
      max_open_connection: 5
      max_idle_connection: 2
      max_lifetime_connection: 60s
      db_migrate: false
    kafka:
      host: host.docker.internal
      port: 9092
    mqtt:
      host: host.docker.internal
      port: 1883

### Environment docker
Please setup the docker environment in this file [.env.docker](.env.docker).

      MYSQL_ROOT_PASSWORD=rootpw
      MYSQL_DATABASE=sensor
      MYSQL_USER=cabin
      MYSQL_PASSWORD=cabinpw

## Migration tools
#### As a standalone tool
[dbconfig.yml](dbconfig.yml) config file should specify like this,

    development:
    dialect: mysql
    datasource: dbname=postgres port=5432 user=routesroot sslmode=disable
    dir: internal/platform/storage/migration

    $ sql-migrate --help
    usage: sql-migrate [--version] [--help] <command> [<args>]

command migration,

    Available commands are:
    down      Undo a database migration
    new       Create a new migration
    redo      Reapply the last migration
    status    Show migration status
    up        Migrates the database to the most recent version available

last, set the db_migrate flag in [app.yaml](config%2Fapp.yaml) of the migration it will be execute when app is starting

	db_migrate: true


## 🏗️ How To Run
open terminal cd to this project's root folder and run script below

    docker-compose up --build -d

## Swagger URL
http://0.0.0.0:8081/swagger/index.html