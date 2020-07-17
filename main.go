package main

import (
	//"encoding/json"
	"database/sql"
	"fmt"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type CustomPayload struct {
	jwt.Payload
	UserId string `json:"user_id,omitempty"`
}

var hs = jwt.NewHS256([]byte(os.Getenv("JWT_PASS")))

func main() {

	gin.DefaultWriter = ioutil.Discard
	r := gin.Default()

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Fatalln(err)
		}
	}

	connString := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s  password=%s sslmode=disable",
		os.Getenv("DB_SERVER"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
	)

	_, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
	}

	go h.run()

	pgListerner := pq.NewListener(connString, 10*time.Second, time.Minute, reportProblem)
	pgListerner.Listen("hustak_jizdy")
	go postgresListen(pgListerner)

	r.GET("/", func(c *gin.Context) {
		cookie, err := c.Cookie("token")
		if err != nil {
			c.AbortWithStatus(403)
			fmt.Println(fmt.Sprintf("Missing token cookie\n"))
			return
		}

		var pl CustomPayload
		_, err = jwt.Verify([]byte(cookie), hs, &pl)
		if err != nil {
			c.AbortWithStatus(403)
			fmt.Println(fmt.Sprintf("Cannot verify user\n"))
			return
		}

		fmt.Println(fmt.Sprintf("User: %s", pl.UserId))

		serveWs(c, pgListerner)
	})
	r.Run(os.Getenv("LISTEN_ADDR"))
}

func postgresListen(l *pq.Listener) {
	for {
		select {
		case n := <-l.Notify:
			log.Printf("Channel: %s; Data: %s\n", n.Channel, n.Extra)
			h.broadcast <- n.Extra
		case <-time.After(10 * time.Second):
			go func() {
				l.Ping()
			}()
		}
	}
}
