package gorma

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/hto/redis-memory-analysis/storages"
)

var Reports = DBReports{}
var redisClient *redis.Client

func Connection(host string, port string, password string) {
	client := redis.NewClient(&redis.Options{
		Addr:        host + ":" + port,
		Password:    password, // no password set
		DB:          0,        // use default DB
		ReadTimeout: 1 * time.Minute,
	})

	pong, err := client.Ping().Result()

	if err != nil || pong == "" {
		log.Fatal("\n\nREDIS NOT CONNECT : ", err)
	}

	redisClient = client
}

func Close() {
	if redisClient != nil {
		redisClient.Close()
	}
}

func Start(delimiters []string) {
	fmt.Println("Starting analysis")

	databases := GetDatabases()

	var (
		// r Report
		// f      float64
		// ttl    int64
		sr SortBySizeReports
		mr KeyReports
	)

	for db, keyCount := range databases {

		fmt.Println("Analyzing db", db)
		mr = KeyReports{}

		keys, _, err := redisClient.Scan(db, "*", keyCount).Result()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("Key Prefix : %s  ResultCount: %d\n", "*", len(keys))

		groupKey := ""
		for _, key := range keys {
			for _, delimiter := range delimiters {
				tmp := strings.Split(key, delimiter)
				if len(tmp) > 1 {
					groupKey = strings.Join(tmp[0:len(tmp)-1], delimiter) + delimiter + "*"
					break
				} else {
					groupKey = key
				}
			}

			// lenght, _ := redisClient.MemoryUsage(key).Result()
			reply, err := redisClient.DebugObject(key).Result()

			lenght, _ := strconv.ParseUint("0", 10, 64)
			if err != nil {

			} else {
				debugs := strings.Split(reply, " ")
				items := strings.Split(debugs[4], ":")
				lenght, _ = strconv.ParseUint(items[1], 10, 64)
			}

			r := Report{}
			if _, ok := mr[groupKey]; ok {
				r = mr[groupKey]
			} else {
				r = Report{groupKey, 0, 0, 0, 0}
			}
			r.Size += uint64(lenght)
			r.Count++
			mr[groupKey] = r
		}

		//Sort by size
		sr = SortBySizeReports{}
		for _, report := range mr {
			sr = append(sr, report)
		}
		sort.Sort(sr)

		Reports[db] = sr
	}
}

func SaveReports(folder string) error {
	fmt.Println("Saving the results of the analysis into...", folder)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		_ = os.MkdirAll(folder, os.ModePerm)
	}

	var (
		str      string
		filename string
		size     float64
		unit     string
	)
	template := fmt.Sprintf("%s%sredis-analysis-%s%s", folder, string(os.PathSeparator), strings.Replace("analysis.redis.Id", ":", "-", -1), "-%d.csv")
	for db, reports := range Reports {
		filename = fmt.Sprintf(template, db)
		fp, err := storages.NewFile(filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		_, _ = fp.Append([]byte("Key,Count,Size\n"))
		for _, value := range reports {
			size, unit = HumanSize(value.Size)
			str = fmt.Sprintf("%s,%d,%s\n",
				value.Key,
				value.Count,
				fmt.Sprintf("%0.3f %s", size, unit))
			_, _ = fp.Append([]byte(str))
		}
		fp.Close()
	}
	return nil
}

func GetDatabases() map[uint64]int64 {

	var databases = make(map[uint64]int64)
	reply, _ := redisClient.Info("keyspace").Result()
	keyspace := strings.Trim(reply[12:], "\n")
	keyspaces := strings.Split(keyspace, "\r")

	for _, db := range keyspaces {
		dbKeysParse := strings.Split(db, ",")
		if dbKeysParse[0] == "" {
			continue
		}

		dbKeysParsed := strings.Split(dbKeysParse[0], ":")
		dbNo, _ := strconv.ParseUint(dbKeysParsed[0][2:], 10, 64)
		dbKeySize, _ := strconv.ParseInt(dbKeysParsed[1][5:], 10, 64)
		databases[dbNo] = dbKeySize
	}
	return databases
}

func HumanSize(byte uint64) (float64, string) {
	units := []string{"Bytes", "KB", "MB", "GB", "TB", "PB", "EB"}
	fb := float64(byte)
	i := 0
	for ; fb >= 1024; i++ {
		fb /= 1024
	}
	size, _ := strconv.ParseFloat(fmt.Sprintf("%0.3f", fb), 64)
	return size, units[i]
}
