package service_test

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"testing"

	"github.com/yousefvand/secret-service/pkg/client"
)

////////////////////////////// CreateSession //////////////////////////////

func Test_SecretServiceCreateSession(t *testing.T) {

	/*
		CreateSession ( IN String algorithm,
		                IN Variant input,
		                OUT Variant output,
		                OUT String serialnumber);
	*/

	t.Run("dh-ietf1024-sha256-aes128-cbc-pkcs7 algorithm", func(t *testing.T) {

		ssClient, _ := client.New()
		err := ssClient.SecretServiceCreateSession(client.Dh_ietf1024_sha256_aes128_cbc_pkcs7)

		if err != nil {
			t.Errorf("CreateSession failed. Error: %v", err)
		}

		if len(Service.SecretService.Session.SerialNumber) != 32 {
			t.Errorf("Unexpected CLI serialnumber length. Expected 32, got '%d'",
				len(Service.SecretService.Session.SerialNumber))
		}

		if len(ssClient.SecretService.Session.SerialNumber) != 32 {
			t.Errorf("Unexpected CLI serialnumber length. Expected 32, got '%d'",
				len(ssClient.SecretService.Session.SerialNumber))
		}

		if result := bytes.Compare(Service.SecretService.Session.SymmetricKey, ssClient.SecretService.Session.SymmetricKey); result != 0 {
			t.Errorf("Symmetric keys are not equal!")
		}

	})

}

////////////////////////////// SetPassword //////////////////////////////

func Test_SecretServiceSetPassword(t *testing.T) {

	/*
		SetPassword ( IN  String      serialnumber
			            IN  Array<Byte> oldpassword,
									IN  Array<Byte> oldpassword_iv,
									IN  Array<Byte> newpassword,
									IN  Array<Byte> newpassword_iv,
									IN  Array<Byte> salt,
									IN  Array<Byte> salt_iv
									OUT String result);
	*/

	t.Run("SetPassword - empty", func(t *testing.T) {

		ssClient, _ := client.New()
		err := ssClient.SecretServiceCreateSession(client.Dh_ietf1024_sha256_aes128_cbc_pkcs7)

		if err != nil {
			t.Errorf("CreateSession failed. Error: %v", err)
		}

		if result := bytes.Compare(Service.SecretService.Session.SymmetricKey, ssClient.SecretService.Session.SymmetricKey); result != 0 {
			t.Errorf("Symmetric keys are not equal!")
		}

		////////////////////////////// SetPassword (empty) //////////////////////////////

		oldPassword_iv, oldPassword_cipher, _ := client.AesCBCEncrypt([]byte(""), ssClient.SecretService.Session.SymmetricKey)
		newPassword_iv, newPassword_cipher, _ := client.AesCBCEncrypt([]byte("Victoria"), ssClient.SecretService.Session.SymmetricKey)
		oldSalt_iv, oldSaltCipher, _ := client.AesCBCEncrypt([]byte(""), ssClient.SecretService.Session.SymmetricKey)
		newSalt_iv, newSaltCipher, _ := client.AesCBCEncrypt([]byte("Salt"), ssClient.SecretService.Session.SymmetricKey)

		result, err := ssClient.SecretServiceSetPassword(ssClient.SecretService.Session.SerialNumber,
			oldPassword_cipher, oldPassword_iv, newPassword_cipher, newPassword_iv,
			oldSaltCipher, oldSalt_iv, newSaltCipher, newSalt_iv)

		if err != nil {
			t.Errorf("SetPassword Failed.Error: %v", err)
		}

		if result != "ok" {
			t.Errorf("Expected 'ok' got: %s", result)
		}
	})

	t.Run("SetPassword - change", func(t *testing.T) {

		ssClient, _ := client.New()
		err := ssClient.SecretServiceCreateSession(client.Dh_ietf1024_sha256_aes128_cbc_pkcs7)

		if err != nil {
			t.Errorf("CreateSession failed. Error: %v", err)
		}

		secret := "OldSalt" + "OldVictoria"
		hasher := sha512.New()
		hasher.Write([]byte(secret))
		hash := hex.EncodeToString(hasher.Sum(nil))

		err = Service.WritePasswordFile(hash)

		if err != nil {
			t.Errorf("Cannot write password file. Error: %v", err)
		}

		oldPassword_iv, oldPassword_cipher, _ := client.AesCBCEncrypt([]byte("OldVictoria"), ssClient.SecretService.Session.SymmetricKey)
		newPassword_iv, newPassword_cipher, _ := client.AesCBCEncrypt([]byte("NewVictoria"), ssClient.SecretService.Session.SymmetricKey)
		oldSalt_iv, oldSaltCipher, _ := client.AesCBCEncrypt([]byte("OldSalt"), ssClient.SecretService.Session.SymmetricKey)
		newSalt_iv, newSaltCipher, _ := client.AesCBCEncrypt([]byte("NewSalt"), ssClient.SecretService.Session.SymmetricKey)

		result, err := ssClient.SecretServiceSetPassword(ssClient.SecretService.Session.SerialNumber,
			oldPassword_cipher, oldPassword_iv, newPassword_cipher, newPassword_iv,
			oldSaltCipher, oldSalt_iv, newSaltCipher, newSalt_iv)

		if err != nil {
			t.Errorf("SetPassword Failed.Error: %v", err)
		}

		if result != "ok" {
			t.Errorf("Expected 'ok' got: %s", result)
		}

	})
}

////////////////////////////// Login //////////////////////////////

func Test_SecretServiceLogin(t *testing.T) {

	t.Run("Login", func(t *testing.T) {

		ssClient, _ := client.New()
		err := ssClient.SecretServiceCreateSession(client.Dh_ietf1024_sha256_aes128_cbc_pkcs7)

		if err != nil {
			t.Errorf("CreateSession failed. Error: %v", err)
		}

		// write a password file
		secret := "Alpha" + "Bravo"
		hasher := sha512.New()
		hasher.Write([]byte(secret))
		hash := hex.EncodeToString(hasher.Sum(nil))

		err = Service.WritePasswordFile(hash)

		if err != nil {
			t.Errorf("Cannot write password file. Error: %v", err)
		}

		// try login
		iv, cipher, err := client.AesCBCEncrypt([]byte(hash),
			ssClient.SecretService.Session.SymmetricKey)

		if err != nil {
			t.Errorf("Encryption failed. Error: %v", err)
		}

		encryptedCookie, cookie_iv, result, err := ssClient.SecretServiceLogin(
			ssClient.SecretService.Session.SerialNumber, cipher, iv)

		if err != nil {
			t.Errorf("Login returned error: %v", err)
		}

		if result != "ok" {
			t.Errorf("Expected 'ok' got: %v", result)
		}

		cookie, err := client.AesCBCDecrypt(cookie_iv, encryptedCookie, ssClient.SecretService.Session.SymmetricKey)

		if err != nil {
			t.Errorf("Decryption failed. Error: %v", err)
		}

		if len(cookie) != 64 {
			t.Errorf("expected cookie of length 64, got: %d", len(cookie))
		}
	})
}

////////////////////////////// Command //////////////////////////////

func Test_SecretServiceCommand(t *testing.T) {

	t.Run("Ping", func(t *testing.T) {

		ssClient, _ := client.New()
		err := ssClient.SecretServiceCreateSession(client.Dh_ietf1024_sha256_aes128_cbc_pkcs7)

		if err != nil {
			t.Errorf("CreateSession failed. Error: %v", err)
		}

		// write a password file
		secret := "Alpha" + "Bravo"
		hasher := sha512.New()
		hasher.Write([]byte(secret))
		hash := hex.EncodeToString(hasher.Sum(nil))

		err = Service.WritePasswordFile(hash)

		if err != nil {
			t.Errorf("Cannot write password file. Error: %v", err)
		}

		// try login
		iv, cipher, err := client.AesCBCEncrypt([]byte(hash),
			ssClient.SecretService.Session.SymmetricKey)

		if err != nil {
			t.Errorf("Encryption failed. Error: %v", err)
		}

		encryptedCookie, cookie_iv, result, err := ssClient.SecretServiceLogin(
			ssClient.SecretService.Session.SerialNumber, cipher, iv)

		if err != nil {
			t.Errorf("Login returned error: %v", err)
		}

		if result != "ok" {
			t.Errorf("Expected 'ok' got: %v", result)
		}

		cookie, err := client.AesCBCDecrypt(cookie_iv, encryptedCookie, ssClient.SecretService.Session.SymmetricKey)

		if string(cookie) != Service.SecretService.Session.Cookie.Value {
			t.Errorf("Cookie mismatch. Error: %v", err)
		}
		if err != nil {
			t.Errorf("Decryption failed. Error: %v", err)
		}

		if len(cookie) != 64 {
			t.Errorf("expected cookie of length 64, got: %d", len(cookie))
		}

		// Command
		// ssClient.SecretServiceCommand(ssClient.SecretService.Session.SerialNumber)

	})
}