package customdata

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

const path = "./daschatascha/%s/"

func SaveCache(keschey string, vaschal []byte) error {
	fmt.Println("SAVECACHE")
	dir := "./daschatascha/reindex/" + keschey
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll("./daschatascha/reindex", 0700)
	}
	key := []byte("passphrasewhichneedstobe32bytes!")

	c, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	err = ioutil.WriteFile(dir+".data", gcm.Seal(nonce, nonce, vaschal, nil), 0777)
	// handle this error
	if err != nil {
		return err
	}
	return nil
}
func DeleteFullCache() error {
	// fmt.Println("DELETEFULLCACHE")
	// dir := "./daschatascha/reindex/"
	// err := os.RemoveAll(dir)
	// if err != nil {
	// 	log.Panic(err)
	// }
	return nil
}
func DeleteExplicitFileCache(keschey string) error {
	fmt.Println("DELETEEXPLICITFILECACHE")
	dir := "./daschatascha/reindex/" + keschey
	defer os.Remove(dir)
	return nil
}
func SaveHex(keschey string, vaschal []byte) error {
	dir := "./daschatascha/blocks/" + keschey
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll("./daschatascha/blocks", 0755)
	}
	key := []byte("passphrasewhichneedstobe32bytes!")

	c, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	err = ioutil.WriteFile(dir+".data", gcm.Seal(nonce, nonce, vaschal, nil), 0777)
	// handle this error
	if err != nil {
		return err
	}
	return nil

}
func SaveHashes(vaschal []byte) error {
	dir := "./daschatascha/hashes/"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
	}
	files, _ := ioutil.ReadDir(dir)
	var count int
	for range files {
		count++
	}
	key := []byte("passphrasewhichneedstobe32bytes!")

	c, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	err = ioutil.WriteFile(dir+strconv.Itoa(count+1)+".data", gcm.Seal(nonce, nonce, vaschal, nil), 0777)
	// handle this error
	if err != nil {
		return err
	}
	return nil
}
func LoadCache() ([]os.FileInfo, error) {
	fmt.Println("LOADCACHE")
	dir := "./daschatascha/reindex"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	return files, nil
}
func DecodeCache(fileName string) ([]byte, error) {
	fmt.Println("DECODECACHE")
	dir := "./daschatascha/reindex/"
	key := []byte("passphrasewhichneedstobe32bytes!")
	ciphertext, err := ioutil.ReadFile(dir + fileName)
	if err != nil {
		fmt.Println(err)
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		fmt.Println(err)
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		fmt.Println(err)
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	bytes, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		fmt.Println(err)
	}
	return bytes, nil
}
func LoadHex(id string) ([]byte, error) {
	dir := "./daschatascha/blocks/" + id
	key := []byte("passphrasewhichneedstobe32bytes!")
	files, _ := ioutil.ReadDir("./daschatascha/blocks")
	var count int
	for range files {
		count++
	}
	if count == 0 {
		return nil, errors.New("nodata")
	}
	ciphertext, err := ioutil.ReadFile(dir + ".data")
	if err != nil {
		fmt.Println(err)
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		fmt.Println(err)
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		fmt.Println(err)
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	bytes, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		fmt.Println(err)
	}
	return bytes, nil
}
func LoadHashes() ([]byte, error) {
	dir := "./daschatascha/hashes/"
	key := []byte("passphrasewhichneedstobe32bytes!")
	files, _ := ioutil.ReadDir(dir)
	var count int
	for range files {
		count++
	}
	if count == 0 {
		return nil, errors.New("nodata")
	}
	ciphertext, err := ioutil.ReadFile(dir + strconv.Itoa(count) + ".data")
	if err != nil {
		fmt.Println(err)
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		fmt.Println(err)
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		fmt.Println(err)
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	bytes, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		fmt.Println(err)
	}
	return bytes, nil
}
