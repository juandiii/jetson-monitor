# 🚨 Jetson Monitor

**Jetson** is an HTTP monitoring service used to notify by various messaging platforms such as **Slack** and **Telegram** 

# ⚙️ Configuration

Configuration of Jetson

## config.yml

| Key            | Description           | Optional |
|----------------|-----------------------|----------|
| url            | Fetch URL of a server | No       |
| status_code    | Return Status Code    | No       |
| slack_token    | Token of Slack        | Yes      |
| telegram_token | Token of Telegram Bot | Yes      |
| scheduler      | CronJob               | No       |

# 👨🏻‍💻 Example usage

Copy a configuration sample a rename to `config.yml`.

```sh
# Copy config.sample.yml
cp config.sample.yml config.yml
```

```yml
urls:
  - url: https://google.com/
    status_code: 200 
    slack_token: ""
    telegram_token: ""
    scheduler: "@every 1m"
```

# 🐋 Docker

##  How to use this image

Run `docker`

```sh
 docker run -d --restart always -v $(pwd):/var/jetson-monitor -e LOG_LEVEL='DEBUG' jetson
```

# 😇 Contribuition

TBH

