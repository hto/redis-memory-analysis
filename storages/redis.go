package storages

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
)

type RedisClient struct {
	Id   string
	conn redis.Conn
}

func NewRedisClient(host string, port uint16, password string) (*RedisClient, error) {
	var addr bytes.Buffer
	addr.WriteString(host)
	addr.WriteString(":")
	addr.WriteString(strconv.Itoa(int(port)))

	conn, err := redis.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}

	if password != "" {
		_, err := conn.Do("AUTH", password)
		if err != nil {
			return nil, err
		}
	}

	return &RedisClient{addr.String(), conn}, err
}

func (client RedisClient) Select(db uint64) error {
	_, err := client.conn.Do("SELECT", db)
	return err
}

func (client RedisClient) GetDatabases() (map[uint64]uint64, error) {

	var databases = make(map[uint64]uint64)
	reply, err := client.conn.Do("INFO", "Keyspace")
	keyspace, err := redis.String(reply, err)
	keyspace = strings.Trim(keyspace[12:], "\n")
	keyspaces := strings.Split(keyspace, "\r")

	for _, db := range keyspaces {
		dbKeysParse := strings.Split(db, ",")
		if dbKeysParse[0] == "" {
			continue
		}

		dbKeysParsed := strings.Split(dbKeysParse[0], ":")
		dbNo, _ := strconv.ParseUint(dbKeysParsed[0][2:], 10, 64)
		dbKeySize, _ := strconv.ParseUint(dbKeysParsed[1][5:], 10, 64)
		databases[dbNo] = dbKeySize
	}
	return databases, err
}

func (client RedisClient) GetCluster() {

	// ab7628bc43960b4852df55fe1db5177f43bddbba 172.31.37.103:6379@1122 slave 4704b073d11db1512935f2acd587d59f2efa1da0 0 1571211331000 1 connected
	reply, err := client.conn.Do("CLUSTER", "NODES")
	keyspace, err := redis.String(reply, err)
	fmt.Println(keyspace, reply)

	return
}

func (client RedisClient) Scan(cursor *uint64, match string, limit uint64) ([]string, error) {
	reply, err := client.conn.Do("SCAN", *cursor, "MATCH", match, "COUNT", limit)
	result, err := redis.Values(reply, err)

	var keys []string

	for _, v := range result {
		switch v.(type) {
		case []uint8:
			*cursor, _ = redis.Uint64(v, nil)
		case []interface{}:
			keys, _ = redis.Strings(v, nil)
		}
	}
	return keys, err
}

func (client RedisClient) Ttl(key string) (int64, error) {
	reply, err := client.conn.Do("TTL", key)
	ttl, err := redis.Int64(reply, err)
	return ttl, err
}

func (client RedisClient) SerializedLength(key string) (uint64, error) {
	reply, err := client.conn.Do("DEBUG", "OBJECT", key)
	debug, err := redis.String(reply, err)

	if err != nil {
		return 0, err
	}

	debugs := strings.Split(debug, " ")
	items := strings.Split(debugs[4], ":")

	return strconv.ParseUint(items[1], 10, 64)
}

func (client RedisClient) Close() error {
	return client.conn.Close()
}
