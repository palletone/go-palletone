/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package redis

import (
	//toml "github.com/extrame/go-toml-config"
	"github.com/gomodule/redigo/redis"
	"github.com/palletone/go-palletone/dag/dagconfig"
)

////TODO
//
//type server struct {
//	plugins map[string]Plugin
//}

type Plugin interface {
	ParseConfig(prefix string) error
	Init() error
}

type Redis struct {
	address  *string
	password *string
	db       *int64
}

var redisPool *redis.Pool
var PoolMaxIdle = 10

// 从配置文件获取 redis配置信息   config: redis
func Init() {
	var addr, pwd, prefix string

	if dagconfig.DagConfig.RedisAddr == "" {
		addr = "localhost"
	} else {
		addr = dagconfig.DagConfig.RedisAddr
	}

	if dagconfig.DagConfig.RedisPwd == "" {
		pwd = ""
	} else {
		pwd = dagconfig.DagConfig.RedisPwd
	}

	if dagconfig.DagConfig.RedisPrefix == "" {
		prefix = "default"
	} else {
		prefix = dagconfig.DagConfig.RedisPrefix
	}
	r := Redis{address: &addr, password: &pwd}
	r.ParseConfig(prefix)
	r.Init()
}

// 从文件解析 redis配置信息
func (r *Redis) ParseConfig(prefix string) error {
	// r.address = toml.String(prefix+".address", "localhost:6379")
	// r.password = toml.String(prefix+".password", "")
	// r.db = toml.Int64(prefix+".db", 0)
	//TODO please use github.com/naoina/toml
	return nil
}

func (r *Redis) Init() error {
	redisPool = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *r.address)

		if err != nil {
			return nil, err
		}
		if len(*r.password) > 0 {
			if _, err := c.Do("AUTH", *r.password); err != nil {
				c.Close()
				return nil, err
			}
		}
		if _, err := c.Do("SELECT", *r.db); err != nil {
			c.Close()
			return nil, err
		}
		return c, nil
	}, PoolMaxIdle)
	return nil
}

func Store(key, itemKey string, item interface{}) error {
	c := redisPool.Get()
	defer c.Close()
	if _, err := c.Do("HSET", key, itemKey, item); err != nil {
		return err
	}
	return nil
}

func Expire(key string, seconds int) error {
	c := redisPool.Get()
	defer c.Close()
	if _, err := c.Do("EXPIRE", key, seconds); err != nil {
		return err
	}
	return nil
}

func Exists(key, itemKey string) bool {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", key, itemKey))
	return count != 0
}

func Get(userKey, itemKey string) (interface{}, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		res, _ := redis.Values(c.Do("HGET", userKey, itemKey))

		return res, true
	}
}

func GetBool(userKey, itemKey string) (bool, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return false, false
	} else {
		n, _ := redis.Bool(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetBytes(userKey, itemKey string) ([]byte, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		n, _ := redis.Bytes(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetFloat64(userKey, itemKey string) (float64, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return 0, false
	} else {
		n, _ := redis.Float64(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetInt(userKey, itemKey string) (int, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return 0, false
	} else {
		n, _ := redis.Int(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetInt64(userKey, itemKey string) (int64, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return 0, false
	} else {
		n, _ := redis.Int64(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetIntMap(userKey, itemKey string) (map[string]int, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		n, _ := redis.IntMap(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetInt64Map(userKey, itemKey string) (map[string]int64, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		n, _ := redis.Int64Map(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetInts(userKey, itemKey string) ([]int, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		n, _ := redis.Ints(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetString(userKey, itemKey string) (string, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return "", false
	} else {
		n, _ := redis.String(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetStrings(userKey, itemKey string) ([]string, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		n, _ := redis.Strings(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetStringMap(userKey, itemKey string) (map[string]string, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return nil, false
	} else {
		n, _ := redis.StringMap(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func GetUint64(userKey, itemKey string) (uint64, bool) {
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", userKey, itemKey))
	if count == 0 {
		return 0, false
	} else {
		n, _ := redis.Uint64(c.Do("HGET", userKey, itemKey))
		return n, true
	}
}

func RemoveItem(userKey, itemKey string) bool {
	c := redisPool.Get()
	defer c.Close()
	if _, err := c.Do("HDEL", userKey, itemKey); err != nil {
		return false
	}
	return true
}
