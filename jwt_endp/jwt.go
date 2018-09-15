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
	signing_key := fromEnv("SIGNING_KEY", "yYTMYjyxvY5dv6luoRbZwkvy0NvFuYWIkPQlUUXHsAY=")
	signing_key_age := fromEnv("SIGNING_AGE", "86400")
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
				c.JSON(500, "Server is experiencing database connectivity issues.")
				return
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

			fmt.Println("Formatting email")
			formattedEmail := ctx.formatJwtEmail(jwt, senderDetails)

			fmt.Println("Formatted email for sender " + email + ":\n" + formattedEmail.message)

			fmt.Println("Spawning email sender")
			go sendFormattedEmail(formattedEmail)

			c.JSON(200, gin.H{"login": ctx.email})
		}
	})

	r.POST("/verifytoken", func(c *gin.Context) {
		//tokenString := c.Param("token")
		tokenString := c.DefaultQuery("token", "")
		skey := jwtverify.SigningKey{signing_key}
		if valid, _ := skey.Verify(tokenString); valid {
			c.SetCookie("auth_token", tokenString, 30000, "", "", false, false)
			c.Writer.Header().Add("Cache-Control", "max-age=0, no-cache, no-store, must-revalidate, proxy-revalidate")
			c.Writer.Header().Add("Pragma", "no-cache")
			c.Writer.Header().Add("Expires", "Tue, 03 Jul 2001 06:00:00 GMT")
			c.Redirect(303, "/")
		} else {
			c.JSON(401, gin.H{"error": "Unauthorized"})
		}

	})

	r.GET("/verifytoken", func(c *gin.Context) {
		//tokenString := c.Param("token")
		tokenString := c.DefaultQuery("token", "")
		skey := jwtverify.SigningKey{signing_key}
		if valid, _ := skey.Verify(tokenString); valid {
			c.SetCookie("auth_token", tokenString, 30000, "", "", false, false)
			c.Writer.Header().Add("Cache-Control", "max-age=0, no-cache, no-store, must-revalidate, proxy-revalidate")
			c.Writer.Header().Add("Pragma", "no-cache")
			c.Writer.Header().Add("Expires", "Tue, 03 Jul 2001 06:00:00 GMT")
			c.Redirect(303, "/")
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

type FormattedEmail struct {
	senderDetails SenderDetails
	message       string
	to_email      string
}

func (c LoginContext) formatJwtEmail(jwt string, senderDetails SenderDetails) FormattedEmail {
	server_path_entry, has_server_path := c.ghttp.Request.Header["X-Request-Server"]
	server_path := ""
	if has_server_path {
		server_path = server_path_entry[0]
	}

	baseurl := server_path + "/verifytoken?token="
	body := "Your Login Token: \n" + baseurl + jwt

	msg := "From: " + senderDetails.from + "\n" +
		"To: " + c.email + "\n" +
		"Subject: Login Information\n\n" +
		body

	formatted_email := FormattedEmail{senderDetails, msg, c.email}
	return formatted_email
}

func sendFormattedEmail(email FormattedEmail) {
	fmt.Println("Attempting to send email to " + email.to_email)
	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", email.senderDetails.from, email.senderDetails.password, "smtp.gmail.com"),
		email.senderDetails.from, []string{email.to_email}, []byte(email.message))

	if err != nil {
		fmt.Println("An error occured sending an email", err)
	} else {
		fmt.Println("Email successfully sent to " + email.to_email)
	}
}

func fromEnv(variable string, defaultValue string) string {
	if value := os.Getenv(variable); value != "" {
		return value
	} else {
		return defaultValue
	}

}
