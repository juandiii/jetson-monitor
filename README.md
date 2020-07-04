# 🚨 Jetson Monitor

**Jetson** is an HTTP monitoring service used to notify by various messaging platforms such as **Slack**, coming soon **Telegram**

# ⚙️ Configuration

Configuration of Jetson

## Options

| Key            | Description           | Required  |
|----------------|-----------------------|-----------|
| url            | Fetch URL of a server | Yes       |
| status_code    | Return Status Code    | Yes       |
| slack_token    | Token of Slack        | No        |
| scheduler      | CronJob               | Yes       |

# 👨🏻‍💻 Example usage

- Token of Slack: **12345678/12345678/12345678ABCDE**


Copy a configuration sample and rename to `config.yml`.

```sh
# Copy config.sample.yml
cp config.sample.yml config.yml
```

```yml
urls:
  - url: https://google.com/
    status_code: 200 
    slack_token: "12345678/12345678/12345678ABCDE"
    scheduler: "@every 1m"
  - url: https://yahoo.com/
    status_code: 200 
    slack_token: "12345678/12345678/12345678ABCDE"
    scheduler: "*/5 * * * *" # Every 5th minute
```

# 🐋 Docker

##  How to use this image

Run `docker`

```sh
 docker run -d \
  --restart always \ 
  -v $(pwd):/var/jetson-monitor \ 
  -e LOG_LEVEL='DEBUG' \ 
  juandiii/jetson-monitor
```

# 😇 Contribuition

TBH

