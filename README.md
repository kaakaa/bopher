# Bopher

`Bot` + `Gopher` = `Bopher`

Bopher is client application connecting Mattermost(3.3.0~) as Bot.
You can call Gopher from chat!

## Requirements

* Mattermost 3.3.0~
* [mattn/gopher](https://github.com/mattn/gopher) binary file (named `gopher.exe`)

## Setup

#### 1. Build gopher binary
* [mattn/gopher](https://github.com/mattn/gopher)

#### 2. Write config.json
```
{
  // Your Mattermost settings
  "mattermost": {
    "host": "localhost",
    "port": "8065",
    "bot": {
      "email": "admin@example.com",
      "password": "admin",
      "name": "Bot",
      "first_name": "Go",
      "last_name": "Bot"
    },
    "team": "tttt",
    "channel": {
      "name":"botting",
      "display_name": "BotRoom",
      "purpose": "bot_test"
    }
  },
  // path to gopher.exe (Even if you use windows, this value is like `$GOPATH` not `%GOPATH%`)
  // details -> https://github.com/golang/go/issues/8469
  "gopher": "$GOPATH/bin/gopher.exe",

  // The maximum number of gophers on display
  // If you set a large number, gopher will eat all of your machine resource)
  "max_of_gophers": 10                 
}
```

#### 3. Run bopher

```
go run main.go
```

#### 4. Call gopher on mattermost
![](https://raw.githubusercontent.com/kaakaa/bopher/master/bopher.gif)

## Commands

Message         | Action
----------------|----------------
`gopher`        | Call a gopher
`bye gopher`    | kill all gophers :scream:
`jump gopher`   | jump all existing gophers
`hello gopher`  | a gopher say hello 
`whats gopher?` | r u sure?
