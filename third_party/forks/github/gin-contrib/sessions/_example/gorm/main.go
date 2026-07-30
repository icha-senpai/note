package main

import (
	"github.com/icha-senpai/note/third_party/forks/github/gin-contrib/sessions"
	gormsessions "github.com/icha-senpai/note/third_party/forks/github/gin-contrib/sessions/gorm"
	"github.com/icha-senpai/note/third_party/forks/github/gin-gonic/gin"
	"github.com/icha-senpai/note/third_party/forks/external/gorm.io/driver/sqlite"
	"github.com/icha-senpai/note/third_party/forks/external/gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	store := gormsessions.NewStore(db, true, []byte("secret"))

	r := gin.Default()
	r.Use(sessions.Sessions("mysession", store))

	r.GET("/incr", func(c *gin.Context) {
		session := sessions.Default(c)
		var count int
		v := session.Get("count")
		if v == nil {
			count = 0
		} else {
			count = v.(int)
			count++
		}
		session.Set("count", count)
		session.Save()
		c.JSON(200, gin.H{"count": count})
	})
	r.Run(":8000")
}
