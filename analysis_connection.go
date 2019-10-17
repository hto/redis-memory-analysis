package gorma

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/hto/redis-memory-analysis/storages"
)

type AnalysisConnection struct {
	redis   *storages.RedisClient
	Reports DBReports
}

func NewAnalysisConnection(host string, port uint16, password string) (*AnalysisConnection, error) {
	redis, err := storages.NewRedisClient(host, port, password)
	if err != nil {
		return nil, err
	}
	return &AnalysisConnection{redis, DBReports{}}, nil
}

func (analysis *AnalysisConnection) Close() {
	if analysis.redis != nil {
		_ = analysis.redis.Close()
	}
}

func (analysis AnalysisConnection) Start(delimiters []string) {
	fmt.Println("Starting analysis")

	databases, _ := analysis.redis.GetDatabases()

	var (
		cursor uint64
		r      Report
		f      float64
		ttl    int64
		length uint64
		sr     SortBySizeReports
		mr     KeyReports
	)

	for db, keyCount := range databases {
		fmt.Println("Analyzing db", db)
		cursor = 0
		mr = KeyReports{}

		_ = analysis.redis.Select(db)

		for {
			keys, _ := analysis.redis.Scan(&cursor, "*", keyCount)
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

				if _, ok := mr[groupKey]; ok {
					r = mr[groupKey]
				} else {
					r = Report{groupKey, 0, 0, 0, 0}
				}

				ttl, _ = analysis.redis.Ttl(key)

				switch ttl {
				case -2:
					continue
				case -1:
					r.NeverExpire++
					r.Count++
				default:
					f = float64(r.AvgTtl*(r.Count-r.NeverExpire)+uint64(ttl)) / float64(r.Count+1-r.NeverExpire)
					ttl, _ := strconv.ParseUint(fmt.Sprintf("%0.0f", f), 10, 64)
					r.AvgTtl = ttl
					r.Count++
				}

				length, _ = analysis.redis.SerializedLength(key)
				r.Size += length

				mr[groupKey] = r
			}

			if cursor == 0 {
				break
			}
		}

		//Sort by size
		sr = SortBySizeReports{}
		for _, report := range mr {
			sr = append(sr, report)
		}
		sort.Sort(sr)

		analysis.Reports[db] = sr

		// fmt.Println(analysis.Reports)
	}
}

func (analysis AnalysisConnection) SaveReports(folder string) error {
	// fmt.Println("Saving the results of the analysis into...", folder)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		_ = os.MkdirAll(folder, os.ModePerm)
	}

	var (
		str      string
		filename string
		size     float64
		unit     string
	)
	template := fmt.Sprintf("%s%sredis-analysis-%s%s", folder, string(os.PathSeparator), strings.Replace(analysis.redis.Id, ":", "-", -1), "-%d.csv")
	for db, reports := range analysis.Reports {
		filename = fmt.Sprintf(template, db)
		fp, err := storages.NewFile(filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		_, _ = fp.Append([]byte("Key,Count,Size,NeverExpire,AvgTtl(excluded never expire)\n"))
		for _, value := range reports {
			size, unit = HumanSize(value.Size)
			str = fmt.Sprintf("%s,%d,%s,%d,%d\n",
				value.Key,
				value.Count,
				fmt.Sprintf("%0.3f %s", size, unit),
				value.NeverExpire,
				value.AvgTtl)
			_, _ = fp.Append([]byte(str))
		}
		fp.Close()
	}
	return nil
}
