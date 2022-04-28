# Lfu cache

This project is use for cache image  and use Clean Architecture. We use lfu and redis cache to mange all 
Requested also handling error fo invalid urls

<img src="docs/images/clean.jpg" alt="clean.jpg" width="700">


## Running

to start extracting from all websites run the command below:

```bash

```

## Framework

- Web : [gin](https://github.com/gin-gonic/gin)
- Configuration : [godotenv](https://github.com/joho/godotenv).

- Database : [RED](github.com/go-redis/redis/v8)

## Architecture

Controller -> Service -> Repository