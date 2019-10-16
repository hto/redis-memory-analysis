Redis memory analysis
======

🔎  Analyzing memory of redis is to find the keys(prefix) which used a lot of memory, export the analysis result into csv file.

[![GoDoc](https://godoc.org/github.com/hto/redis-memory-analysis?status.svg)](https://godoc.org/github.com/hto/redis-memory-analysis)
[![Go Report Card](https://goreportcard.com/badge/github.com/hto/redis-memory-analysis)](https://goreportcard.com/report/github.com/hto/redis-memory-analysis)
[![Sourcegraph](https://sourcegraph.com/github.com/hto/redis-memory-analysis/-/badge.svg)](https://sourcegraph.com/github.com/hto/redis-memory-analysis?badge)

## Binary File Usage

1. Download the appropriate binary file from [Releases](https://github.com/hto/redis-memory-analysis/releases)

2. Run

```Shell
# help
./redis-memory-analysis-linux-amd64 -h
Usage of ./redis-memory-analysis-darwin-amd64:
  -ip string
    	The host of redis (default "127.0.0.1")
  -password string
    	The password of redis (default "")
  -port uint
    	The port of redis (default 6379)
  -rdb string
    	The rdb file of redis (default "")
  -prefixes string
    	The prefixes list of redis key, be split by '//', special pattern characters need to escape by '\' (default "#//:")
  -reportPath string
    	The csv file path of analysis result (default "./reports")

# run by connecting to redis
./redis-memory-analysis-linux-amd64 -ip="127.0.0.1" -port=6380 -password="abc" -prefixes="#//:"

# run by redis rdb file
./redis-memory-analysis-linux-amd64 -rdb="./6379_dump.rdb" -prefixes="#//:"
```

## Source Code Usage

1. Install

```Shell
//cd your-root-folder-of-project
// dep init
dep ensure -add github.com/hto/redis-memory-analysis@~2.0.0
```

2. Run

- Analyze keys by connecting to redis directly.

```Go
analysis := NewAnalysis()
//Open redis: 127.0.0.1:6379 without password
err := analysis.Open("127.0.0.1", 6379, "")
defer analysis.Close()
if err != nil {
    fmt.Println("something wrong:", err)
    return
}

//Scan the keys which can be split by '#' ':'
//Special pattern characters need to escape by '\'
analysis.Start([]string{"#", ":"})

//Find the csv file in default target folder: ./reports
//CSV file name format: redis-analysis-{host:port}-{db}.csv
//The keys order by count desc
err = analysis.SaveReports("./reports")
if err == nil {
    fmt.Println("done")
} else {
    fmt.Println("error:", err)
}
```

- Analyze keys by redis `RDB` file, but cannot work out the size of key.

```Go
analysis := NewAnalysis()
//Open redis rdb file: ./6379_dump.rdb
err := analysis.OpenRDB("./6379_dump.rdb")
defer analysis.CloseRDB()
if err != nil {
    fmt.Println("something wrong:", err)
    return
}

//Scan the keys which can be split by '#' ':'
//Special pattern characters need to escape by '\'
analysis.StartRDB([]string{"#", ":"})

//Find the csv file in default target folder: ./reports
//CSV file name format: redis-analysis-{host:port}-{db}.csv
//The keys order by count desc
err = analysis.SaveReports("./reports")
if err == nil {
    fmt.Println("done")
} else {
    fmt.Println("error:", err)
}
```

![CSV](https://raw.githubusercontent.com/hto/redis-memory-analysis/master/examples/demo.png)

## Another tool implemented by PHP

[redis-memory-analysis](https://github.com/hhxsv5/redis-memory-analysis)


## License

[MIT](https://github.com/hto/redis-memory-analysis/blob/master/LICENSE)
