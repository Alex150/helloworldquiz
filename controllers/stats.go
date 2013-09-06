package controllers

import (
  "fmt"
  // "math"
  "langgame/models/variant"
  "langgame/initializers/redisPool"
  "github.com/astaxie/beego"
  "github.com/garyburd/redigo/redis"
)

type StatsController struct {
  beego.Controller
}

type StatsObject struct {
  Variants        string
  Games           string
  CompletionRate  string
  Easiest         *StatsVariant
  Hardest         *StatsVariant
  Highest         string
}

type StatsVariant struct {
  Key             string
  Score           int
  OppositeScore   int
  Name            string
  Value           string
}

func (this *StatsController) Get() {
  conn := redisPool.Get()
  defer conn.Close()
  
  var stats StatsObject
  stats.Variants = fmt.Sprintf("%v", len(variant.Objects))
  
  games, _ := redis.Int(conn.Do("HLEN", "games"))
  stats.Games = fmt.Sprintf("%v", games)
  
  died, _ := redis.Int(conn.Do("GET", "died"))
  survived, _ := redis.Int(conn.Do("GET", "survived"))
  stats.CompletionRate = fmt.Sprintf("%.1f", (float64(died + survived) / float64(games)) * 100)
  
  highest, _ := redis.Int(conn.Do("GET", "highest"))
  stats.Highest = fmt.Sprintf("%v", highest)
  
  stats.Easiest = FindMaxVariant(conn, "success", "failure")
  stats.Hardest = FindMaxVariant(conn, "failure", "success")
  
  this.Data["json"] = stats
  this.ServeJson()
}

func FindMaxVariant(conn redis.Conn, source string, opposite string) (response *StatsVariant) {
  variants, err := redis.Values(conn.Do("HGETALL", source))
  response = &StatsVariant{}
  
  for k, _score := range variants {
    if k % 2 == 1 {
      score, _ := redis.Int(_score, err)
      if score > response.Score {
        key, _ := redis.String(variants[k-1], err)
        
        response.Key = key
        response.Score = score
      }
    }
  }
  
  if response.Key != "" {
    response.OppositeScore, _ = redis.Int(conn.Do("HGET", opposite, response.Key))
    response.Name = variant.Objects[response.Key].Name
    response.Value = fmt.Sprintf("%.1f", (float64(response.Score) / float64(response.Score + response.OppositeScore)) * 100)
  }
  
  return response
}
