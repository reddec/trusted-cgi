package cases

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/reddec/trusted-cgi/internal"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type sshLoader struct {
	keyLock        sync.RWMutex
	publicKey      []byte
	privateKeyFile string
}

func (sshLoader *sshLoader) PublicSSHKey() ([]byte, error) {
	sshLoader.keyLock.RLock()
	defer sshLoader.keyLock.RUnlock()
	if len(sshLoader.publicKey) == 0 {
		return nil, fmt.Errorf("key not defined")
	}
	cp := make([]byte, len(sshLoader.publicKey))
	copy(cp, sshLoader.publicKey)
	return cp, nil
}

func (sshLoader *sshLoader) SetPrivateSSHKeyFile(privateKeyFile string) error {
	pmData, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return fmt.Errorf("read private key file: %w", err)
	}
	info, _ := pem.Decode(pmData)
	priv, err := x509.ParsePKCS1PrivateKey(info.Bytes)
	if err != nil {
		return fmt.Errorf("parse private key: %w", err)
	}
	publicKey, err := ssh.NewPublicKey(priv.Public())
	if err != nil {
		return fmt.Errorf("derive public key from private: %w", err)
	}
	sshLoader.keyLock.Lock()
	defer sshLoader.keyLock.Unlock()
	sshLoader.publicKey = ssh.MarshalAuthorizedKey(publicKey)
	sshLoader.privateKeyFile = privateKeyFile
	return nil
}

func (sshLoader *sshLoader) SetOrCreatePrivateSSHKeyFile(privateKeyFile string) error {
	err := sshLoader.SetPrivateSSHKeyFile(privateKeyFile)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	_, err = sshLoader.generateSSHKeys(privateKeyFile)
	if err != nil {
		return fmt.Errorf("generate ssh key: %w", err)
	}
	return sshLoader.SetPrivateSSHKeyFile(privateKeyFile)
}

func (sshLoader *sshLoader) generateSSHKeys(file string) (*rsa.PrivateKey, error) {
	log.Println("generating ssh key to", file)
	privateKey, err := rsa.GenerateKey(rand.Reader, internal.SSHKeySize)
	if err != nil {
		return nil, err
	}
	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)
	err = ioutil.WriteFile(file, pemdata, 0600)
	if err != nil {
		return privateKey, err
	}
	return privateKey, nil
}
