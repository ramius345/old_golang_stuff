package main

import (
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"pineapple.no-ip.biz/dao"
	"pineapple.no-ip.biz/jwtverify"
)

func main() {
	listen_port := fromEnv("LISTEN_ADDR", "0.0.0.0:5556")
	cluster_name := fromEnv("CLUSTER_NAME", "greengrape.pineapple.no-ip.biz")
	cluster_port := fromEnv("CLUSTER_PORT", "30000")
	keyspace := fromEnv("KEYSPACE", "imageapp")
	signing_key := fromEnv("SIGNING_KEY", "1234567890")
	signing_key_age := fromEnv("SIGNING_AGE", "3600")
	sender_email := fromEnv("SENDER_EMAIL", "imagesharing392@gmail.com")
	sender_pass := fromEnv("SENDER_PASS", "aybabtu1")

	cluster_port_int, err := strconv.ParseInt(cluster_port, 10, 32)
	if err != nil {
		fmt.Println("Error reading cluster port!")
		os.Exit(1)
	}

	signing_key_age_int, err := strconv.ParseInt(signing_key_age, 10, 64)
	if err != nil {
		fmt.Println("Error reading signing key age!")
		os.Exit(1)
	}

	fmt.Println("Connecting to cluster " + cluster_name + " on port " + cluster_port)
	fmt.Println("Using keyspace " + keyspace)
	cluster := gocql.NewCluster(cluster_name)
	cluster.Port = int(cluster_port_int)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	r := gin.Default()

	senderDetails := SenderDetails{sender_email, sender_pass}

	r.POST("/login", func(c *gin.Context) {
		if email, namePresent := c.GetQuery("email"); namePresent {
			session, error := cluster.CreateSession()
			if error != nil {
				fmt.Println(error)
				os.Exit(1)
			}
			defer session.Close()

			ctx := LoginContext{
				email,
				session,
				c,
				signing_key,
				signing_key_age_int,
			}

			fmt.Println("Verifying email")
			if !ctx.verifyEmail() {
				c.JSON(401, gin.H{"error": "Unauthorized"})
				return
			}

			fmt.Println("Signing jwt for email " + email)
			jwt, err := ctx.signJwt()
			if err != nil {
				c.JSON(401, gin.H{"error": "Signing error"})
				return
			}

			fmt.Println("Spawning goroutine to handle email for " + email)
			ctx.handleSendEmail(jwt, senderDetails, email)

			c.JSON(200, gin.H{"login": ctx.email})
		}
	})

	r.GET("/verify/:token", func(c *gin.Context) {
		tokenString := c.Param("token")
		skey := jwtverify.SigningKey{signing_key}
		if valid, _ := skey.Verify(tokenString); valid {
			c.SetCookie("auth_token", tokenString, 30000, "", "", false, false)
			c.Redirect(302, "/")
		} else {
			c.JSON(401, gin.H{"error": "Unauthorized"})
		}

	})

	r.GET("/verify", func(c *gin.Context) {
		if tokenString, error := c.Cookie("auth_token"); error == nil {
			fmt.Println("JWT: " + tokenString)
			skey := jwtverify.SigningKey{signing_key}
			if valid, claims := skey.Verify(tokenString); valid {
				c.SetCookie("auth_token", tokenString, 30000, "", "", false, false)
				c.JSON(200, gin.H{"email": (*claims)["email"], "exp": (*claims)["exp"]})
				return
			}
		}
		c.JSON(401, gin.H{"error": "unauthorized"})
	})

	r.Run(listen_port)
}

type LoginContext struct {
	email            string
	session          *gocql.Session
	ghttp            *gin.Context
	signing_key      string
	signing_age_secs int64
}

func (ctx LoginContext) handleSendEmail(jwt string, senderDetails SenderDetails, email string) {
	fmt.Println("Sending email with jwt to " + email)
	err := ctx.sendJwtEmail(jwt, senderDetails)
	if err != nil {
		fmt.Println("There was an error sending a login email to " + email)
		return
	}
}

func (c LoginContext) verifyEmail() bool {
	perms := dao.GetUserPerms(c.session, c.email)
	fmt.Printf("Permlen %v\n", len(perms))
	return len(perms) > 0
}

func (c LoginContext) signJwt() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"email": c.email,
		"exp":   time.Now().Unix() + c.signing_age_secs,
	})

	tokenString, err := token.SignedString([]byte(c.signing_key))
	return tokenString, err
}

type SenderDetails struct {
	from     string
	password string
}

func (c LoginContext) sendJwtEmail(jwt string, senderDetails SenderDetails) error {
	// for k, v := range c.ghttp.Request.Header {
	// 	fmt.Printf("Header: %v, Value: %v\n", k, v)
	// }

	server_path_entry, has_server_path := c.ghttp.Request.Header["X-Request-Server"]
	server_path := ""
	if has_server_path {
		server_path = server_path_entry[0]
	}

	baseurl := server_path + "/verify/"
	body := "Your Login Token: \n" + baseurl + jwt

	msg := "From: " + senderDetails.from + "\n" +
		"To: " + c.email + "\n" +
		"Subject: Login Information\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", senderDetails.from, senderDetails.password, "smtp.gmail.com"),
		senderDetails.from, []string{c.email}, []byte(msg))

	return err
}

func fromEnv(variable string, defaultValue string) string {
	if value := os.Getenv(variable); value != "" {
		return value
	} else {
		return defaultValue
	}

}
