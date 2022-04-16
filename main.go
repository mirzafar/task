package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"main/connection"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Person struct {
	Id        int    `json:"id" form:"id"`
	FirstName string `json:"first_name" form:"first_name"`
	LastName  string `json:"last_name" form:"last_name"`
}

type Email struct {
	Id    int
	Email string
}

func lengthOf(s string) int {
	lastOcc := make(map[byte]int)
	srart := 0
	maxLength := 0
	for i, ch := range []byte(s) {
		if lastI, ok := lastOcc[ch]; ok && lastI >= srart {
			srart = lastI + 1
		}
		if i-srart+1 > maxLength {
			maxLength = i - srart + 1
		}
		lastOcc[ch] = i
	}
	return maxLength
}
func task1(c *gin.Context) {
	messge := c.Request.FormValue("message")
	result := lengthOf(messge)
	c.JSON(http.StatusOK, gin.H{
		"max_len": result,
	})
}

func indexHandler(c *gin.Context) {
	c.String(http.StatusOK, "Hello world")
}

func createuser(c *gin.Context) {

	db := connection.SetupDB()

	firstName := c.Request.FormValue("first_name")
	lastName := c.Request.FormValue("last_name")

	rs, err := db.Exec("INSERT INTO `personal`(`firstname`, `lastname`) VALUES (?, ?)", firstName, lastName)
	if err != nil {
		log.Fatalln(err)
	}

	id, err := rs.LastInsertId()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("insert person Id {}", id)
	msg := fmt.Sprintf("insert successful %d", id)
	c.JSON(http.StatusOK, gin.H{
		"msg": msg,
	})
}

func getuser(c *gin.Context) {

	vars := c.Param("id")

	db := connection.SetupDB()
	results, err := db.Query(fmt.Sprintf("SELECT `id`,`firstname`, `lastname` FROM `personal` WHERE `id` = '%s'", vars))
	if err != nil {
		panic(err.Error())
	}
	showEdit := Person{}
	for results.Next() {
		var tag Person
		err = results.Scan(&tag.Id, &tag.FirstName, &tag.LastName)
		if err != nil {
			panic(err.Error())
		}
		showEdit = tag
	}

	c.JSON(http.StatusOK, gin.H{
		"person": showEdit,
	})
}

func deluser(c *gin.Context) {
	db := connection.SetupDB()
	id := c.Param("id")
	drop, err := db.Prepare("DELETE FROM `personal` WHERE `id`=?")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer drop.Close()

	_, err = drop.Exec(id)
	if err != nil {
		log.Panic(err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "succes",
	})
}

func edituser(c *gin.Context) {
	db := connection.SetupDB()
	cid := c.Param("id")
	id, err := strconv.Atoi(cid)
	if err != nil {
		log.Fatalln(err)
	}

	first_name := "aaaaaaaa"
	last_name := "bbbbbbbb"

	stmt, err := db.Prepare("UPDATE `personal` SET `firstname`=?, `lastname`=? WHERE `id`=?")

	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(first_name, last_name, id)

	if err != nil {
		log.Panic(err.Error())
	}
	msg := "succes"
	c.JSON(http.StatusOK, gin.H{
		"msg": msg,
	})
}

func task2(c *gin.Context) {

	// db := connection.SetupDB()

	var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	email := c.Request.FormValue("email")
	result := fmt.Sprintf("Email: %v :%v", email, emailRegex.MatchString(email))

	// results, err := db.Query("SELECT `email` FROM `emails` WHERE `email` LIKE concat('%', '?', '%')", email)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// ema := []Email{}
	// for results.Next() {
	// 	var elem Email
	// 	err = results.Scan(&elem.Id, &elem.Email)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// 	ema = append(ema, elem)
	// }
	c.JSON(http.StatusOK, gin.H{
		"result": result,
		// "mmm": ema,
	})
}

func decrypt(encryptedString string, keyString string) (decryptedString string) {

	key, _ := hex.DecodeString(keyString)
	enc, _ := hex.DecodeString(encryptedString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return fmt.Sprintf("%s", plaintext)
}
func encrypt(stringToEncrypt string, keyString string) (encryptedString string) {

	//Since the key is in string, we need to convert decode it to byteslog.Println(password)
	key, _ := hex.DecodeString(keyString)
	plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext)
}

func calc(c *gin.Context) {

	message := c.Request.FormValue("message")
	key := hex.EncodeToString(make([]byte, 32))
	encryptpwd := encrypt(message, key)
	c.JSON(http.StatusOK, gin.H{
		"pass_hex": encryptpwd,
	})
}

func resultcalc(c *gin.Context) {
	key := hex.EncodeToString(make([]byte, 32))
	cid := c.Param("id")
	decrypted := decrypt(cid, key)
	c.JSON(http.StatusOK, gin.H{
		"pass": decrypted,
	})
}

func main() {

	router := gin.Default()

	router.GET("/", indexHandler)
	router.POST("/rest/substr/find", task1)
	router.POST("/rest/email/check", task2)
	router.GET("/rest/hash/result/$id", resultcalc)
	router.POST("/rest/hash/calc", calc)
	router.POST("/rest/user", createuser)
	router.GET("/rest/user/:id", getuser)
	router.DELETE("/rest/user/:id", deluser)
	router.PUT("/rest/user/:id", edituser)
	router.Run(":8000")
}
